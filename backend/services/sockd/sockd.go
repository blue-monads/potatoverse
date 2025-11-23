package sockd

import (
	"github.com/blue-monads/turnix/backend/services/sockd/broadcast"
	"github.com/blue-monads/turnix/backend/services/sockd/higher"
	"github.com/blue-monads/turnix/backend/services/sockd/pubsub"
)

type Sockd struct {
	broadcast broadcast.BroadcastSockd
	pubsub    pubsub.PubSubSockd
	higher    higher.HigherSockd
}

func NewSockd() *Sockd {
	return &Sockd{
		broadcast: broadcast.New(),
		pubsub:    pubsub.New(),
		higher:    higher.New(),
	}
}

func (s *Sockd) GetBroadcast() *broadcast.BroadcastSockd {
	return &s.broadcast
}

func (s *Sockd) GetPubSub() *pubsub.PubSubSockd {
	return &s.pubsub
}

func (s *Sockd) GetHigher() *higher.HigherSockd {
	return &s.higher
}
