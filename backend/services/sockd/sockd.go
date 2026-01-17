package sockd

import (
	"github.com/blue-monads/potatoverse/backend/services/sockd/broadcast"
	"github.com/blue-monads/potatoverse/backend/services/sockd/notifier"
	"github.com/blue-monads/potatoverse/backend/services/sockd/pubsub"
)

type Sockd struct {
	broadcast broadcast.BroadcastSockd
	pubsub    pubsub.PubSubSockd
	notifier  notifier.Notifier
}

func NewSockd() *Sockd {
	s := &Sockd{
		broadcast: broadcast.New(),
		pubsub:    pubsub.New(),
		notifier:  notifier.New(),
	}

	go s.notifier.Run()

	return s
}

func (s *Sockd) GetBroadcast() *broadcast.BroadcastSockd {
	return &s.broadcast
}

func (s *Sockd) GetPubSub() *pubsub.PubSubSockd {
	return &s.pubsub
}

func (s *Sockd) GetNotifier() *notifier.Notifier {
	return &s.notifier
}
