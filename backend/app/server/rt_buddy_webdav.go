package server

import (
	"sync"

	"github.com/blue-monads/turnix/backend/app/server/webdav"
	"github.com/gin-gonic/gin"
)

const (
	BuddyWebdavPrefix = "/zz/buddy/webdav"
)

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
		server := getWebdavServer(ctx.Request.Host)

		server.Handle(ctx)
	}

}
