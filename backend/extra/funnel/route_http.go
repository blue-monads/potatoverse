package funnel

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"maps"
	"net/http"
	"net/http/httputil"

	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

// Route routes an HTTP request to the specified server and writes the response back to gin.Context
func (f *Funnel) routeHttp(serverId string, c *gin.Context) {
	qq.Println("@routeHttp/1", serverId)

	// Get server connection
	f.scLock.RLock()
	serverConn, exists := f.serverConnections[serverId]
	f.scLock.RUnlock()

	qq.Println("@routeHttp/2")

	if !exists {
		qq.Println("@routeHttp/2{SERVER_NOT_CONNECTED}")
		c.Error(errors.New("server not connected"))
		return
	}

	qq.Println("@routeHttp/2.1")

	// Generate request ID
	reqId := GetRequestId()

	qq.Println("@routeHttp/2.2")

	pendingReqChan := make(chan *Packet)
	f.pendingReqLock.Lock()
	f.pendingReq[reqId] = pendingReqChan
	f.pendingReqLock.Unlock()

	defer func() {
		qq.Println("@cleanup/1{REQ_ID}", reqId)
		f.pendingReqLock.Lock()
		delete(f.pendingReq, reqId)
		f.pendingReqLock.Unlock()
	}()

	qq.Println("@routeHttp/3")

	// Dump request
	req := c.Request
	out, err := httputil.DumpRequest(req, false)
	if err != nil {
		c.Error(err)
		return
	}

	qq.Println("@routeHttp/4")

	// Write request header packet

	serverConn.writeChan <- &ServerWrite{
		packet: &Packet{
			PType:  PTypeSendHeader,
			Offset: 0,
			Total:  int32(req.ContentLength),
			Data:   out,
		},
		reqId: reqId,
	}

	qq.Println("@routeHttp/6")

	if req.ContentLength > 0 {

		qq.Println("@routeHttp/7")

		fbuf := make([]byte, FragmentSize)
		offset := int32(0)

		for {

			qq.Println("@routeHttp/8")

			last := false
			n, err := req.Body.Read(fbuf)
			if err != nil {
				if err == io.EOF {
					log.Println("EOF")
					last = true
				} else {
					log.Println("@err/Read", err.Error())
					panic(err)
				}
			}

			ptype := PtypeSendBody
			if last {
				ptype = PtypeEndBody
			}

			toSend := fbuf[:n]

			err = WritePacket(serverConn.conn, &Packet{
				PType:  ptype,
				Offset: int32(offset),
				Total:  int32(req.ContentLength),
				Data:   toSend,
			})

			if err != nil {
				c.Error(err)
				return
			}

			offset += int32(n)

			if offset >= int32(req.ContentLength) {
				break
			}

			if last {
				break
			}

		}

	}

	wpack := <-pendingReqChan
	if wpack.PType != PTypeSendHeader {
		c.Error(errors.New("invalid packet type"))
		return
	}

	reader := bytes.NewReader(wpack.Data)
	resp, err := http.ReadResponse(bufio.NewReader(reader), c.Request)
	if err != nil {
		log.Println("@err/ReadResponse", err.Error())
		panic(err)
	}

	header := c.Writer.Header()
	if resp.ContentLength > -1 {
		header.Del("Content-Length")
	}

	maps.Copy(header, resp.Header)

	c.Writer.WriteHeader(resp.StatusCode)

	for {
		wpack := <-pendingReqChan
		if wpack == nil {
			break
		}

		for {
			n, err := c.Writer.Write(wpack.Data)
			if err != nil {
				pp.Println("@err/Write", err.Error())
				break
			}
			wpack.Data = wpack.Data[n:]
			if len(wpack.Data) == 0 {
				break
			}
		}

		if wpack.PType == PtypeEndBody {
			break
		}
	}

}
