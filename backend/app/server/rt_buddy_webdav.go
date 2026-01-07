package server

import (
	"errors"
	"sync"

	"github.com/blue-monads/turnix/backend/app/server/webdav"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
)

const (
	BuddyWebdavPrefix = "/zz/buddy/webdav"
)

func (s *Server) webdavAuth(ctx *gin.Context) (string, string, error) {
	user, pass, ok := ctx.Request.BasicAuth()
	if !ok {
		return "", "", errors.New("Unauthorized")
	}

	return user, pass, nil
}

func (s *Server) handleBuddyWebdav() func(ctx *gin.Context) {

	servers := make(map[string]*webdav.WebdavServer)
	rwLock := sync.RWMutex{}

	getWebdavServer := func(skey string) *webdav.WebdavServer {
		rwLock.RLock()
		server, exists := servers[skey]
		rwLock.RUnlock()
		if !exists {

			buddyDir, err := s.buddyhub.GetBuddyDir(skey)
			if err != nil {
				return nil
			}

			server = webdav.New(buddyDir, BuddyWebdavPrefix)
			server.Build()
			rwLock.Lock()
			servers[skey] = server
			rwLock.Unlock()
		}
		return server
	}

	return func(ctx *gin.Context) {

		buddyPubkey, authKey, err := s.webdavAuth(ctx)
		if err != nil {
			httpx.WriteAuthErr(ctx, err)
			return
		}

		ev, err := verifyNostrAuth(authKey)
		if err != nil {
			httpx.WriteAuthErr(ctx, err)
			return
		}

		if ev.PubKey == buddyPubkey {
			httpx.WriteAuthErr(ctx, errors.New("Wrong buddy pubkey"))
			return
		}

		server := getWebdavServer(buddyPubkey)

		server.Handle(ctx)
	}

}
