package sockd

import (
	"github.com/blue-monads/turnix/backend/services/sockd/broadcast"
	"github.com/blue-monads/turnix/backend/services/sockd/pubsub"
)

type Sockd struct {
	broadcast broadcast.BroadcastSockd
	pubsub    pubsub.PubSubSockd
}

func NewSockd() *Sockd {
	return &Sockd{
		broadcast: broadcast.NewSockd(),
		pubsub:    pubsub.NewSockd(),
	}
}

func (s *Sockd) GetBroadcast() *broadcast.BroadcastSockd {
	return &s.broadcast
}

func (s *Sockd) GetPubSub() *pubsub.PubSubSockd {
	return &s.pubsub
}
