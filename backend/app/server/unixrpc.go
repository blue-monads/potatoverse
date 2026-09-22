package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/blue-monads/potatoverse/backend/xtypes"
)

type UNIXRpcRequest = xtypes.UNIXRpcRequest
type UNIXRpcResponse = xtypes.UNIXRpcResponse

func (s *Server) handleUnixRPC(c net.Conn) {
	defer c.Close()

	_ = c.SetDeadline(time.Now().Add(5 * time.Second))

	reader := bufio.NewReader(c)
	line, err := reader.ReadString('\n')
	if err != nil && !(errors.Is(err, io.EOF) && line != "") {
		s.writeUnixRPC(c, UNIXRpcResponse{Ok: false, Msg: "failed to read request"})
		return
	}

	raw := strings.TrimSpace(line)
	if raw == "" {
		s.writeUnixRPC(c, UNIXRpcResponse{Ok: false, Msg: "empty request"})
		return
	}

	var req UNIXRpcRequest
	if strings.HasPrefix(raw, "{") {
		if err := json.Unmarshal([]byte(raw), &req); err != nil {
			s.writeUnixRPC(c, UNIXRpcResponse{Ok: false, Msg: "invalid request json: " + err.Error()})
			return
		}
	} else {
		req = UNIXRpcRequest{Method: raw}
	}

	normMethod := strings.TrimPrefix(strings.ToLower(req.Method), "/")

	switch normMethod {
	case "info":
		s.writeUnixRPC(c, UNIXRpcResponse{
			Ok:  true,
			Msg: "ok",
			Data: map[string]any{
				"port": s.opt.Port,
				"host": s.opt.Hosts,
			},
		})
	case "get_admin_token":
		user, token, err := s.ctrl.GetAdminToken()
		if err != nil {
			s.writeUnixRPC(c, UNIXRpcResponse{Ok: false, Msg: err.Error()})
			return
		}
		s.writeUnixRPC(c, UNIXRpcResponse{
			Ok:  true,
			Msg: "ok",
			Data: map[string]any{
				"user":  user,
				"token": token,
			},
		})
	case "list_admin_user", "list_admin_users":
		users, err := s.ctrl.ListAdminUsers()
		if err != nil {
			s.writeUnixRPC(c, UNIXRpcResponse{Ok: false, Msg: err.Error()})
			return
		}
		data := map[string]any{
			"users": users,
		}
		if len(users) > 0 {
			data["user"] = users[0]
		}
		s.writeUnixRPC(c, UNIXRpcResponse{
			Ok:   true,
			Msg:  "ok",
			Data: data,
		})
	case "list_all_user", "list_all_users":
		users, err := s.ctrl.ListAllUsers()
		if err != nil {
			s.writeUnixRPC(c, UNIXRpcResponse{Ok: false, Msg: err.Error()})
			return
		}
		s.writeUnixRPC(c, UNIXRpcResponse{
			Ok:  true,
			Msg: "ok",
			Data: map[string]any{
				"users": users,
			},
		})
	case "reset_user_pass", "reset_user_password":
		userParam := req.GetStringArg("user", "user_id", "username", "id")
		passParam := req.GetStringArg("password", "pass")
		user, password, err := s.ctrl.ResetUserPass(userParam, passParam)
		if err != nil {
			s.writeUnixRPC(c, UNIXRpcResponse{Ok: false, Msg: err.Error()})
			return
		}
		s.writeUnixRPC(c, UNIXRpcResponse{
			Ok:  true,
			Msg: "ok",
			Data: map[string]any{
				"user":     user,
				"password": password,
			},
		})
	default:
		s.writeUnixRPC(c, UNIXRpcResponse{Ok: false, Msg: "unknown method: " + req.Method})
	}
}

func (s *Server) writeUnixRPC(c net.Conn, resp UNIXRpcResponse) {
	if err := json.NewEncoder(c).Encode(resp); err != nil {
		log.Println("unix socket write error:", err.Error())
	}
}
