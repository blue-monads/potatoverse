package server

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// endpoint that accepts all ws and sends +yesyes to all messages
func (s *Server) testWS(ctx *gin.Context) {
	conn, _, _, err := ws.UpgradeHTTP(ctx.Request, ctx.Writer)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		msg, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			break
		}
		reply := append(msg, []byte("+yes_websocket_yes")...)
		if err := wsutil.WriteServerText(conn, reply); err != nil {
			break
		}
	}
}

// endpoint that accepts all POST data and appends +yesyes to the body and returns it
func (s *Server) testPOST(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.String(400, "failed to read body")
		return
	}
	result := append(body, []byte("+yes_post_yes")...)
	ctx.Data(200, "text/plain", result)
}
