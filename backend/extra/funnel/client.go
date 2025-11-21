package funnel

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/blue-monads/turnix/backend/utils/kosher"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type FunnelClientOptions struct {
	LocalHttpPort   int
	RemoteFunnelUrl string
	ServerId        string
}

type FunnelClient struct {
	opts   FunnelClientOptions
	ctx    context.Context
	cancel context.CancelFunc
	conn   net.Conn
}

func NewFunnelClient(opts FunnelClientOptions) *FunnelClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &FunnelClient{
		opts:   opts,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *FunnelClient) Start() error {
	// Parse remote funnel URL

	qq.Println("@FunnelClient/Start/1{REMOTE_FUNNEL_URL}", c.opts.RemoteFunnelUrl)

	u, err := url.Parse(c.opts.RemoteFunnelUrl)
	if err != nil {
		return fmt.Errorf("invalid remote funnel URL: %w", err)
	}

	// Determine websocket URL
	wsScheme := "ws"
	if u.Scheme == "https" {
		wsScheme = "wss"
	}

	u.Scheme = wsScheme

	finalUrl := u.String()

	qq.Println("@FunnelClient/Start/2{FINAL_URL}", finalUrl)

	// Connect to remote funnel via websocket
	conn, _, _, err := ws.Dial(c.ctx, finalUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to funnel: %w", err)
	}
	c.conn = conn

	// Start handling incoming requests from funnel
	err = c.handleFunnelConnection(conn)
	conn.Close()
	return err
}

func (c *FunnelClient) Stop() {
	c.cancel()
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *FunnelClient) handleFunnelConnection(conn net.Conn) error {
	// Read request ID (16 bytes) first, then packet
	reqIdBuf := make([]byte, 16)

	for {
		// Check if context is cancelled
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
		}

		// Set read deadline to allow context cancellation to work
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		_, err := io.ReadFull(conn, reqIdBuf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			// Check if it's a timeout (expected for context cancellation)
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Check context again
				select {
				case <-c.ctx.Done():
					return c.ctx.Err()
				default:
					continue
				}
			}
			return err
		}
		conn.SetReadDeadline(time.Time{}) // Clear deadline

		reqId := kosher.Str(reqIdBuf)

		qq.Println("@FunnelClient/handleFunnelConnection/3{REQ_ID}", reqId)

		// Read header packet
		headerPacket, err := ReadPacket(conn)
		if err != nil {
			return err
		}

		if headerPacket.PType != PTypeSendHeader {
			// Invalid packet type, skip
			continue
		}

		// Parse request
		reader := bytes.NewBuffer(headerPacket.Data)
		req, err := http.ReadRequest(bufio.NewReader(reader))
		if err != nil {
			// Send error response
			continue
		}

		// Check if it's a websocket request
		if req.Header.Get("Upgrade") == "websocket" {
			qq.Println("@FunnelClient/handleFunnelConnection/4{WEBSOCKET_REQUEST}")
			// Handle websocket request
			go c.handleWebSocketRequest(conn, reqId, req)
		} else {
			qq.Println("@FunnelClient/handleFunnelConnection/5{HTTP_REQUEST}")
			// Handle HTTP request
			go c.handleHttpRequest(conn, reqId, req)
		}
	}
}

func (c *FunnelClient) handleHttpRequest(conn net.Conn, reqId string, req *http.Request) {
	// Modify request URL to point to local server
	req.URL.Host = fmt.Sprintf("localhost:%d", c.opts.LocalHttpPort)
	req.URL.Scheme = "http"
	req.RequestURI = ""
	req.Host = fmt.Sprintf("localhost:%d", c.opts.LocalHttpPort)

	// Set up request body reader if needed
	if req.ContentLength > 0 {
		req.Body = &requestReader{
			conn: conn,
		}
	}

	// Make request to local server
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Send error response
		return
	}
	defer resp.Body.Close()

	// Dump response header
	out, err := httputil.DumpResponse(resp, false)
	if err != nil {
		return
	}

	// Write request ID
	_, err = conn.Write([]byte(reqId))
	if err != nil {
		return
	}

	// Write response header packet
	err = WritePacket(conn, &Packet{
		PType:  PTypeSendHeader,
		Offset: 0,
		Total:  int32(resp.ContentLength),
		Data:   out,
	})
	if err != nil {
		return
	}

	// Send response body
	if resp.ContentLength > 0 {
		offset := int32(0)
		fbuf := make([]byte, FragmentSize)

		for {
			n, err := resp.Body.Read(fbuf)
			if err != nil && err != io.EOF {
				return
			}

			if n == 0 {
				// Send EndBody
				err = WritePacket(conn, &Packet{
					PType:  PtypeEndBody,
					Offset: offset,
					Total:  int32(resp.ContentLength),
					Data:   []byte{},
				})
				break
			}

			ptype := PtypeSendBody
			if err == io.EOF {
				ptype = PtypeEndBody
			}

			err = WritePacket(conn, &Packet{
				PType:  ptype,
				Offset: offset,
				Total:  int32(resp.ContentLength),
				Data:   fbuf[:n],
			})
			if err != nil {
				return
			}

			offset += int32(n)

			if err == io.EOF {
				break
			}
		}
	} else if resp.ContentLength == 0 {
		// Explicitly send EndBody for zero-length responses
		err = WritePacket(conn, &Packet{
			PType:  PtypeEndBody,
			Offset: 0,
			Total:  0,
			Data:   []byte{},
		})
		if err != nil {
			return
		}
	} else {
		// ContentLength < 0 means unknown/chunked, read until EOF
		offset := int32(0)
		fbuf := make([]byte, FragmentSize)

		for {
			n, err := resp.Body.Read(fbuf)
			if err != nil && err != io.EOF {
				return
			}

			if n == 0 {
				// Send EndBody
				err = WritePacket(conn, &Packet{
					PType:  PtypeEndBody,
					Offset: offset,
					Total:  -1, // Unknown total
					Data:   []byte{},
				})
				break
			}

			ptype := PtypeSendBody
			if err == io.EOF {
				ptype = PtypeEndBody
			}

			err = WritePacket(conn, &Packet{
				PType:  ptype,
				Offset: offset,
				Total:  -1, // Unknown total
				Data:   fbuf[:n],
			})
			if err != nil {
				return
			}

			offset += int32(n)

			if err == io.EOF {
				break
			}
		}
	}
}

func (c *FunnelClient) handleWebSocketRequest(conn net.Conn, reqId string, req *http.Request) {
	// Parse local websocket URL
	port := strconv.Itoa(c.opts.LocalHttpPort)
	wsUrl := fmt.Sprintf("ws://localhost:%s%s", port, req.URL.Path)

	// Connect to local websocket server using gobwas/ws
	localWS, _, _, err := ws.Dial(context.TODO(), wsUrl)
	if err != nil {
		// Could not connect to local websocket
		return
	}
	defer localWS.Close()

	reqIdBytes := []byte(reqId)

	// After sending the header packet, websocket communication uses packets with request ID
	// Forward from local WS to funnel
	go func() {
		for {
			msg, _, err := wsutil.ReadServerData(localWS)
			if err != nil {
				return
			}

			// Write request ID
			_, err = conn.Write(reqIdBytes)
			if err != nil {
				return
			}

			// Write WebSocket data as packet
			err = WritePacket(conn, &Packet{
				PType:  PtypeWebSocketData,
				Offset: 0,
				Total:  int32(len(msg)),
				Data:   msg,
			})
			if err != nil {
				return
			}
		}
	}()

	// Forward from funnel to local WS
	for {
		// Read request ID first
		reqIdBuf := make([]byte, 16)
		_, err := io.ReadFull(conn, reqIdBuf)
		if err != nil {
			if err != io.EOF {
				// Connection error
			}
			break
		}

		// Verify this is for our request
		if string(reqIdBuf) != reqId {
			// This message is for a different request, skip it
			// Read the packet to consume it
			packet, err := ReadPacket(conn)
			if err != nil {
				break
			}
			// Skip if not WebSocket data
			if packet.PType != PtypeWebSocketData {
				continue
			}
			// This shouldn't happen, but if it does, we've consumed the packet
			continue
		}

		// Read WebSocket data packet
		packet, err := ReadPacket(conn)
		if err != nil {
			break
		}

		if packet.PType != PtypeWebSocketData {
			// Invalid packet type
			break
		}

		err = wsutil.WriteClientBinary(localWS, packet.Data)
		if err != nil {
			break
		}
	}
}

// requestReader reads request body from packets
type requestReader struct {
	conn     net.Conn
	total    int64
	received int64
	buffer   []byte
}

func (r *requestReader) Read(p []byte) (int, error) {
	// If we have buffered data, return it first
	if len(r.buffer) > 0 {
		n := copy(p, r.buffer)
		r.buffer = r.buffer[n:]
		r.received += int64(n)
		return n, nil
	}

	// Read next packet
	packet, err := ReadPacket(r.conn)
	if err != nil {
		return 0, err
	}

	if packet.PType != PtypeSendBody && packet.PType != PtypeEndBody {
		return 0, io.ErrUnexpectedEOF
	}

	// Copy data to buffer
	n := copy(p, packet.Data)
	r.received += int64(n)

	// If there's remaining data, buffer it
	if n < len(packet.Data) {
		r.buffer = packet.Data[n:]
	}

	// Check if we're done
	if packet.PType == PtypeEndBody {
		if len(r.buffer) == 0 {
			return n, io.EOF
		}
	}

	return n, nil
}

func (r *requestReader) Close() error {
	return nil
}
