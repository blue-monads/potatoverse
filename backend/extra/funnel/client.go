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
	"sync"

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
	opts FunnelClientOptions

	pendingRequests map[string]chan *Packet
	prLock          sync.Mutex

	writeChan chan *ServerWrite
}

func NewFunnelClient(opts FunnelClientOptions) *FunnelClient {
	return &FunnelClient{
		opts:            opts,
		pendingRequests: make(map[string]chan *Packet),
		prLock:          sync.Mutex{},
		writeChan:       make(chan *ServerWrite),
	}
}

func (c *FunnelClient) writePackets(conn net.Conn) {

	errCount := 0

	for {

		sw := <-c.writeChan

		if sw == nil {
			break
		}

		err := WritePacketFull(conn, sw.packet, sw.reqId)
		if err != nil {
			qq.Println("@FunnelClient/writePackets/1{ERROR}", err)
			errCount++
			if errCount > 10 {
				qq.Println("@FunnelClient/writePackets/2{BREAK}")
				break
			}
			continue
		}

		errCount = 0

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
	conn, _, _, err := ws.Dial(context.Background(), finalUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to funnel: %w", err)
	}

	go c.writePackets(conn)

	// Start handling incoming requests from funnel
	err = c.handleFunnelConnection(conn)
	if err != nil {
		return fmt.Errorf("failed to handle funnel connection: %w", err)
	}

	conn.Close()
	return err
}

func (c *FunnelClient) Stop() {
	c.writeChan <- nil
	close(c.writeChan)
}

func (c *FunnelClient) handleFunnelConnection(conn net.Conn) error {
	// Read request ID (16 bytes) first, then packet
	reqIdBuf := make([]byte, 16)

	for {
		_, err := io.ReadFull(conn, reqIdBuf)
		if err != nil {
			continue
		}

		reqId := string(reqIdBuf)

		qq.Println("@FunnelClient/handleFunnelConnection/3{REQ_ID}", reqId)

		// Read header packet
		headerPacket, err := ReadPacket(conn)
		if err != nil {
			qq.Println("@FunnelClient/handleFunnelConnection/3{ERROR}", err)
			continue
		}

		if headerPacket.PType != PTypeSendHeader {

			c.prLock.Lock()
			pendingRequest := c.pendingRequests[reqId]
			c.prLock.Unlock()

			if pendingRequest == nil {
				qq.Println("@FunnelClient/handleFunnelConnection/4{PENDING_REQUEST_NOT_FOUND}")
				continue
			}

			pendingRequest <- headerPacket

			if headerPacket.PType == PtypeEndBody || headerPacket.PType == PtypeEndSocket {
				c.prLock.Lock()
				delete(c.pendingRequests, reqId)
				c.prLock.Unlock()
			}

			continue
		}

		// Parse request
		reader := bytes.NewBuffer(headerPacket.Data)
		req, err := http.ReadRequest(bufio.NewReader(reader))
		if err != nil {
			// Send error response
			continue
		}

		pendingReqChan := make(chan *Packet)

		c.prLock.Lock()
		c.pendingRequests[reqId] = pendingReqChan
		c.prLock.Unlock()

		// Check if it's a websocket request
		if req.Header.Get("Upgrade") == "websocket" {
			qq.Println("@FunnelClient/handleFunnelConnection/4{WEBSOCKET_REQUEST}")
			// Handle websocket request
			go c.handleWebSocketRequest(pendingReqChan, reqId, req)
		} else {
			qq.Println("@FunnelClient/handleFunnelConnection/5{HTTP_REQUEST}")
			// Handle HTTP request
			go c.handleHttpRequest(pendingReqChan, reqId, req)
		}
	}
}

func (c *FunnelClient) handleHttpRequest(pch chan *Packet, reqId string, req *http.Request) {
	// Modify request URL to point to local server
	req.URL.Host = fmt.Sprintf("localhost:%d", c.opts.LocalHttpPort)
	req.URL.Scheme = "http"
	req.RequestURI = ""
	req.Host = fmt.Sprintf("localhost:%d", c.opts.LocalHttpPort)

	// Set up request body reader if needed
	if req.ContentLength > 0 {

		defer func() {
			c.prLock.Lock()
			delete(c.pendingRequests, reqId)
			c.prLock.Unlock()
		}()

		req.Body = &requestReader{
			pendingRequest: pch,
			total:          req.ContentLength,
			received:       0,
			buffer:         make([]byte, FragmentSize),
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		qq.Println("@FunnelClient/handleHttpRequest/2{ERROR}", err)
		return
	}
	defer resp.Body.Close()

	// Dump response header
	out, err := httputil.DumpResponse(resp, false)
	if err != nil {
		return
	}

	c.writeChan <- &ServerWrite{
		packet: &Packet{
			PType:  PTypeSendHeader,
			Offset: 0,
			Total:  int32(resp.ContentLength),
			Data:   out,
		},
		reqId: reqId,
	}

	if resp.ContentLength == 0 {

		qq.Println("@handleHttpRequest/case_zero_length")

		c.writeChan <- &ServerWrite{
			packet: &Packet{
				PType:  PtypeEndBody,
				Offset: 0,
				Total:  0,
				Data:   []byte{},
			},
			reqId: reqId,
		}

		return

	}

	// Send response body
	if resp.ContentLength > 0 {
		qq.Println("@handleHttpRequest/case_positive_length")

		offset := int32(0)
		fbuf := make([]byte, FragmentSize)

		for {

			qq.Println("@loop/1")

			n, err := resp.Body.Read(fbuf)
			if err != nil && err != io.EOF {
				qq.Println("@loop/2{ERROR}", err)
				return
			}

			qq.Println("@loop/3{N}", n)

			if n == 0 {
				// Send EndBody
				qq.Println("@loop/4{SEND_END_BODY}")
				c.writeChan <- &ServerWrite{
					packet: &Packet{
						PType:  PtypeEndBody,
						Offset: offset,
						Total:  int32(resp.ContentLength),
						Data:   []byte{},
					},
					reqId: reqId,
				}

				qq.Println("@loop/6{BREAK}")

				break
			}

			qq.Println("@loop/7{SEND_BODY}")

			ptype := PtypeSendBody
			if err == io.EOF {
				qq.Println("@loop/8{SEND_END_BODY}")
				ptype = PtypeEndBody
			}

			qq.Println("@loop/9{SEND_BODY}")

			c.writeChan <- &ServerWrite{
				packet: &Packet{
					PType:  ptype,
					Offset: offset,
					Total:  int32(resp.ContentLength),
					Data:   fbuf[:n],
				},
				reqId: reqId,
			}

			qq.Println("@loop/11{OFFSET}", offset)

			offset += int32(n)

			if err == io.EOF {
				qq.Println("@loop/12{BREAK}")
				break
			}

			qq.Println("@loop/13{LOOP}")
		}
	} else {
		qq.Println("@handleHttpRequest/case_negative_length/chunked")

		offset := int32(0)
		fbuf := make([]byte, FragmentSize)

		for {
			n, err := resp.Body.Read(fbuf)
			if err != nil && err != io.EOF {
				return
			}

			if n == 0 {
				// Send EndBody
				c.writeChan <- &ServerWrite{
					packet: &Packet{
						PType:  PtypeEndBody,
						Offset: offset,
						Total:  -1,
						Data:   []byte{},
					},
					reqId: reqId,
				}
				break
			}

			ptype := PtypeSendBody
			if err == io.EOF {
				ptype = PtypeEndBody
			}

			c.writeChan <- &ServerWrite{
				packet: &Packet{
					PType:  ptype,
					Offset: offset,
					Total:  -1,
					Data:   fbuf[:n],
				},
				reqId: reqId,
			}

			offset += int32(n)

			if err == io.EOF {
				break
			}
		}
	}
}

func (c *FunnelClient) handleWebSocketRequest(pch chan *Packet, reqId string, req *http.Request) {

	defer func() {
		c.prLock.Lock()
		delete(c.pendingRequests, reqId)
		c.prLock.Unlock()
	}()

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

	// After sending the header packet, websocket communication uses packets with request ID
	// Forward from local WS to funnel
	go func() {
		for {
			msg, _, err := wsutil.ReadServerData(localWS)
			if err != nil {
				return
			}

			// Write WebSocket data as packet
			c.writeChan <- &ServerWrite{
				packet: &Packet{
					PType:  PtypeWebSocketData,
					Offset: 0,
					Total:  int32(len(msg)),
					Data:   msg,
				},
				reqId: reqId,
			}
		}
	}()

	// Forward from funnel to local WS
	for {
		packet := <-pch
		if packet == nil {
			break
		}

		err = wsutil.WriteServerBinary(localWS, packet.Data)
		if err != nil {
			break
		}

	}
}

// requestReader reads request body from packets
type requestReader struct {
	pendingRequest chan *Packet
	total          int64
	received       int64
	buffer         []byte
}

func (r *requestReader) Read(p []byte) (int, error) {
	// Check if we've already read all the data
	if r.total > 0 && r.received >= r.total {
		return 0, io.EOF
	}

	// If we have buffered data, return it first
	if len(r.buffer) > 0 {
		// Limit copy to remaining bytes if total is set
		toCopy := r.buffer
		if r.total > 0 {
			remaining := r.total - r.received
			if int64(len(toCopy)) > remaining {
				toCopy = toCopy[:remaining]
			}
		}
		n := copy(p, toCopy)
		r.buffer = r.buffer[n:]
		r.received += int64(n)

		// Check if we're done
		if r.total > 0 && r.received >= r.total {
			return n, io.EOF
		}
		return n, nil
	}

	// Read next packet
	packet, ok := <-r.pendingRequest
	if !ok {
		return 0, io.EOF
	}

	// Limit data to copy based on total if set
	dataToCopy := packet.Data
	if r.total > 0 {
		remaining := r.total - r.received
		if int64(len(dataToCopy)) > remaining {
			dataToCopy = dataToCopy[:remaining]
		}
	}

	// Copy data to buffer
	n := copy(p, dataToCopy)
	r.received += int64(n)

	// If there's remaining data in dataToCopy that we couldn't fit in p, buffer it
	if n < len(dataToCopy) {
		r.buffer = dataToCopy[n:]
	}

	// Check if we're done
	if packet.PType == PtypeEndBody || (r.total > 0 && r.received >= r.total) {
		if len(r.buffer) == 0 {
			return n, io.EOF
		}
	}

	return n, nil
}

func (r *requestReader) Close() error {
	return nil
}
