package funnel

import (
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/gin-gonic/gin"
)

// Route routes an HTTP request to the specified server and writes the response back to gin.Context
func (f *Funnel) routeHttp(serverId string, c *gin.Context) {
	// Get server connection
	f.scLock.RLock()
	serverConn, exists := f.serverConnections[serverId]
	f.scLock.RUnlock()

	if !exists {
		c.JSON(http.StatusBadGateway, gin.H{"error": "server not connected"})
		return
	}

	// Generate request ID
	reqId := GetRequestId()
	reqIdBytes := []byte(reqId)

	// Create pending request
	pending := &PendingRequest{
		ResponseChan: make(chan *http.Response, 1),
		ErrorChan:    make(chan error, 1),
	}

	f.prLock.Lock()
	f.pendingRequests[reqId] = pending
	f.prLock.Unlock()

	// Clean up pending request when done
	defer func() {
		f.prLock.Lock()
		delete(f.pendingRequests, reqId)
		f.prLock.Unlock()
	}()

	// Dump request
	req := c.Request
	out, err := httputil.DumpRequest(req, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Write request ID
	_, err = serverConn.Conn.Write(reqIdBytes)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to write request id"})
		return
	}

	// Write request header packet
	err = WritePacket(serverConn.Conn, &Packet{
		PType:  PTypeSendHeader,
		Offset: 0,
		Total:  int32(req.ContentLength),
		Data:   out,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to write request header"})
		return
	}

	// If request has body, send it
	if req.Body != nil && req.ContentLength > 0 {
		offset := int32(0)
		fbuf := make([]byte, FragmentSize)
		sentEndBody := false

		for {
			n, err := req.Body.Read(fbuf)
			if err != nil && err != io.EOF {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if n == 0 {
				// Send EndBody if we haven't already
				if !sentEndBody {
					err = WritePacket(serverConn.Conn, &Packet{
						PType:  PtypeEndBody,
						Offset: offset,
						Total:  int32(req.ContentLength),
						Data:   []byte{},
					})
					if err != nil {
						c.JSON(http.StatusBadGateway, gin.H{"error": "failed to write request body end"})
						return
					}
					sentEndBody = true
				}
				break
			}

			ptype := PtypeSendBody
			if err == io.EOF {
				ptype = PtypeEndBody
				sentEndBody = true
			}

			toSend := fbuf[:n]
			err = WritePacket(serverConn.Conn, &Packet{
				PType:  ptype,
				Offset: offset,
				Total:  int32(req.ContentLength),
				Data:   toSend,
			})

			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": "failed to write request body"})
				return
			}

			offset += int32(n)

			if err == io.EOF {
				break
			}
		}
	} else if req.Body != nil {
		// Body exists but ContentLength is unknown or 0, read until EOF
		offset := int32(0)
		fbuf := make([]byte, FragmentSize)

		for {
			n, err := req.Body.Read(fbuf)
			if err != nil && err != io.EOF {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if n == 0 {
				// Send EndBody
				err = WritePacket(serverConn.Conn, &Packet{
					PType:  PtypeEndBody,
					Offset: offset,
					Total:  -1, // Unknown total
					Data:   []byte{},
				})
				if err != nil {
					c.JSON(http.StatusBadGateway, gin.H{"error": "failed to write request body end"})
					return
				}
				break
			}

			ptype := PtypeSendBody
			if err == io.EOF {
				ptype = PtypeEndBody
			}

			toSend := fbuf[:n]
			err = WritePacket(serverConn.Conn, &Packet{
				PType:  ptype,
				Offset: offset,
				Total:  -1, // Unknown total
				Data:   toSend,
			})

			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": "failed to write request body"})
				return
			}

			offset += int32(n)

			if err == io.EOF {
				break
			}
		}
	}

	// Wait for response
	select {
	case resp := <-pending.ResponseChan:
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Writer.Header().Add(key, value)
			}
		}
		c.Writer.WriteHeader(resp.StatusCode)

		// Copy response body
		_, err := io.Copy(c.Writer, resp.Body)
		if err != nil {
			// Response already started, can't send error
			return
		}

	case err := <-pending.ErrorChan:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
}
