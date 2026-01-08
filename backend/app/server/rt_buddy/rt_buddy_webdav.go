package rtbuddy

import (
	"errors"

	"github.com/blue-monads/turnix/backend/app/server/rt_buddy/webdav"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
)

const (
	BuddyWebdavPrefix = "/zz/buddy/webdav"
)

func (s *BuddyRouteServer) webdavAuth(ctx *gin.Context) (string, string, error) {
	user, pass, ok := ctx.Request.BasicAuth()
	if !ok {
		return "", "", errors.New("Unauthorized")
	}

	return user, pass, nil
}

func (s *BuddyRouteServer) handleBuddyWebdav(ctx *gin.Context) {

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

	server := s.getWebdavServer(buddyPubkey)

	server.Handle(ctx)

}

func (s *BuddyRouteServer) getWebdavServer(buddyPubkey string) *webdav.WebdavServer {
	s.webdavLock.RLock()
	server, exists := s.webdavServers[buddyPubkey]
	s.webdavLock.RUnlock()
	if !exists {

		buddyDir, err := s.buddyhub.GetBuddyDir(buddyPubkey)
		if err != nil {
			return nil
		}

		server = webdav.New(buddyDir, BuddyWebdavPrefix)
		server.Build()
		s.webdavLock.Lock()
		s.webdavServers[buddyPubkey] = server
		s.webdavLock.Unlock()
	}
	return server

}
