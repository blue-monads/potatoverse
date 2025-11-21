package funnel

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

// Route routes an HTTP request to the specified server and writes the response back to gin.Context
func (f *Funnel) routeHttp(serverId string, c *gin.Context) {
	qq.Println("@Funnel/routeHttp/1{SERVER_ID}", serverId)

	// Get server connection
	f.scLock.RLock()
	serverConn, exists := f.serverConnections[serverId]
	f.scLock.RUnlock()

	qq.Println("@Funnel/routeHttp/2{SERVER_CONN}")

	if !exists {
		c.Error(errors.New("server not connected"))
		return
	}

	// Generate request ID
	reqId := GetRequestId()
	reqIdBytes := []byte(reqId)

	pendingReqChan := make(chan *Packet)
	f.pendingReqLock.Lock()
	f.pendingReq[reqId] = pendingReqChan
	f.pendingReqLock.Unlock()

	defer func() {
		f.pendingReqLock.Lock()
		delete(f.pendingReq, reqId)
		f.pendingReqLock.Unlock()
	}()

	// Dump request
	req := c.Request
	out, err := httputil.DumpRequest(req, false)
	if err != nil {
		c.Error(err)
		return
	}

	// Write request ID
	_, err = serverConn.Write(reqIdBytes)
	if err != nil {
		c.Error(err)
		return
	}

	// Write request header packet
	err = WritePacket(serverConn, &Packet{
		PType:  PTypeSendHeader,
		Offset: 0,
		Total:  int32(req.ContentLength),
		Data:   out,
	})
	if err != nil {
		c.Error(err)
		return
	}

	if req.ContentLength > 0 {

		fbuf := make([]byte, FragmentSize)
		offset := int32(0)

		for {

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

			err = WritePacket(serverConn, &Packet{
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
	for k, v := range resp.Header {
		header[k] = v
	}

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
