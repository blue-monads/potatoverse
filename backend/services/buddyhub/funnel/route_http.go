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
	"strconv"

	"github.com/blue-monads/potatoverse/backend/services/buddyhub/packetwire"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

func DebugLog(a ...interface{}) (n int, err error) {

	return 0, nil
}

// Route routes an HTTP request to the specified server and writes the response back to gin.Context
func (f *Funnel) routeHttp(nodeId string, c *gin.Context) {
	DebugLog("@routeHttp/1", nodeId)

	// Get server connection
	f.scLock.RLock()
	serverConn, exists := f.serverConnections[nodeId]
	f.scLock.RUnlock()

	DebugLog("@routeHttp/2")

	if !exists {
		DebugLog("@routeHttp/2{SERVER_NOT_CONNECTED}")
		c.Abort()
		f.dumpConnIds()

		return
	}

	DebugLog("@routeHttp/2.1")

	// Generate request ID
	reqId := packetwire.GetRequestId()

	DebugLog("@routeHttp/2.2")

	pendingReqChan := make(chan *packetwire.Packet)
	f.pendingReqLock.Lock()
	f.pendingReq[reqId] = pendingReqChan
	f.pendingReqLock.Unlock()

	defer func() {
		DebugLog("@cleanup/1{REQ_ID}", reqId)
		f.pendingReqLock.Lock()
		delete(f.pendingReq, reqId)
		f.pendingReqLock.Unlock()
	}()

	DebugLog("@routeHttp/3")

	// Dump request
	req := c.Request
	out, err := httputil.DumpRequest(req, false)
	if err != nil {
		c.Error(err)
		return
	}

	DebugLog("@routeHttp/4")

	// Write request header packet

	serverConn.writeChan <- &ServerWrite{
		packet: &packetwire.Packet{
			PType:  packetwire.PTypeSendHeader,
			Offset: 0,
			Total:  int32(req.ContentLength),
			Data:   out,
		},
		reqId: reqId,
	}

	DebugLog("@routeHttp/6")

	if req.ContentLength > 0 {

		DebugLog("@routeHttp/7")

		offset := int32(0)

		for {
			fbuf := make([]byte, packetwire.FragmentSize)

			DebugLog("@routeHttp/8")

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

			if n == 0 && last {
				// Send EndBody packet for EOF with no data
				serverConn.writeChan <- &ServerWrite{
					packet: &packetwire.Packet{
						PType:  packetwire.PtypeEndBody,
						Offset: int32(offset),
						Total:  int32(req.ContentLength),
						Data:   []byte{},
					},
					reqId: reqId,
				}
				break
			}

			if n == 0 {
				// No data read, skip this iteration
				continue
			}

			ptype := packetwire.PtypeSendBody
			if last {
				ptype = packetwire.PtypeEndBody
			}

			toSend := fbuf[:n]

			serverConn.writeChan <- &ServerWrite{
				packet: &packetwire.Packet{
					PType:  ptype,
					Offset: int32(offset),
					Total:  int32(req.ContentLength),
					Data:   toSend,
				},
				reqId: reqId,
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
	if wpack.PType != packetwire.PTypeSendHeader {
		c.Error(errors.New("invalid packet type"))
		return
	}

	reader := bytes.NewReader(wpack.Data)
	resp, err := http.ReadResponse(bufio.NewReader(reader), c.Request)
	if err != nil {
		log.Println("@err/ReadResponse", err.Error())
		panic(err)
	}

	DebugLog("@routeHttp/parseResponse/1{STATUS}", resp.StatusCode, "CONTENT_LENGTH", resp.ContentLength)

	header := c.Writer.Header()
	maps.Copy(header, resp.Header)

	// Ensure Content-Length is set correctly if it was in the response
	// (maps.Copy should have already copied it, but we ensure it's correct)
	if resp.ContentLength > -1 {
		header.Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	}

	DebugLog("@routeHttp/parseResponse/2{HEADERS_COPIED}")

	c.Writer.WriteHeader(resp.StatusCode)

	for {
		wpack := <-pendingReqChan
		if wpack == nil {
			break
		}

		DebugLog("@routeHttp/writeBody/1{PACKET_TYPE}", wpack.PType, "DATA_LEN", len(wpack.Data))

		for {
			n, err := c.Writer.Write(wpack.Data)
			if err != nil {
				pp.Println("@err/Write", err.Error())
				break
			}
			DebugLog("@routeHttp/writeBody/2{WRITTEN}", n, "REMAINING", len(wpack.Data)-n)
			wpack.Data = wpack.Data[n:]
			if len(wpack.Data) == 0 {
				break
			}
		}

		if wpack.PType == packetwire.PtypeEndBody {
			DebugLog("@routeHttp/writeBody/3{END_BODY}")
			break
		}
	}

}

func (f *Funnel) dumpConnIds() {
	f.scLock.RLock()
	defer f.scLock.RUnlock()

	keys := make([]string, 0, len(f.serverConnections))
	for k := range f.serverConnections {
		keys = append(keys, k)
	}

	DebugLog("@keys", keys)

}
