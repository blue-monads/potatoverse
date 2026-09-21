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

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
)

type unixRPCResponse struct {
	Ok   bool           `json:"ok"`
	Msg  string         `json:"msg,omitempty"`
	Data map[string]any `json:"data,omitempty"`
}

func (s *Server) handleUnixRPC(c net.Conn) {
	defer c.Close()

	_ = c.SetDeadline(time.Now().Add(5 * time.Second))

	reader := bufio.NewReader(c)
	line, err := reader.ReadString('\n')
	if err != nil && !(errors.Is(err, io.EOF) && line != "") {
		s.writeUnixRPC(c, unixRPCResponse{Ok: false, Msg: "failed to read request"})
		return
	}

	path := strings.TrimSpace(line)
	if path == "" {
		s.writeUnixRPC(c, unixRPCResponse{Ok: false, Msg: "empty request"})
		return
	}

	switch path {
	case "/info":
		s.writeUnixRPC(c, unixRPCResponse{
			Ok:  true,
			Msg: "ok",
			Data: map[string]any{
				"port": s.opt.Port,
				"host": s.opt.Hosts,
			},
		})
	case "/get_admin_token":
		user, token, err := s.issueAdminAuthToken()
		if err != nil {
			s.writeUnixRPC(c, unixRPCResponse{Ok: false, Msg: err.Error()})
			return
		}
		s.writeUnixRPC(c, unixRPCResponse{
			Ok:  true,
			Msg: "ok",
			Data: map[string]any{
				"user":  user,
				"token": token,
			},
		})
	default:
		s.writeUnixRPC(c, unixRPCResponse{Ok: false, Msg: "unknown method: " + path})
	}
}

func (s *Server) writeUnixRPC(c net.Conn, resp unixRPCResponse) {
	if err := json.NewEncoder(c).Encode(resp); err != nil {
		log.Println("unix socket write error:", err.Error())
	}
}

func (s *Server) issueAdminAuthToken() (*dbmodels.User, string, error) {
	users, err := s.ctrl.ListUsers(0, 200)
	if err != nil {
		return nil, "", err
	}

	var admin *dbmodels.User
	for i := range users {
		u := &users[i]
		if u.Ugroup == "admin" && !u.Disabled && !u.IsDeleted {
			admin = u
			break
		}
	}
	if admin == nil {
		return nil, "", errors.New("no admin user found")
	}

	token, err := s.signer.SignAccess(&signer.AccessClaim{
		UserId: admin.ID,
	})
	if err != nil {
		return nil, "", err
	}

	admin.Password = ""
	admin.ExtraMeta = ""

	return admin, token, nil
}
