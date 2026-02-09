package nostrhq

import (
	"context"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/fiatjaf/eventstore"
	"github.com/nbd-wtf/go-nostr"
)

var (
	_ eventstore.Store = (*NoopStore)(nil)
)

type NoopStore struct{}

func (n *NoopStore) Init() error {
	qq.Println("Init@noop")
	return nil
}
func (n *NoopStore) Close() {

}

func (n *NoopStore) QueryEvents(context.Context, nostr.Filter) (chan *nostr.Event, error) {
	qq.Println("QueryEvents@noop")

	ch := make(chan *nostr.Event)
	// no events in noop store: close immediately
	close(ch)
	return ch, nil
}

func (n *NoopStore) DeleteEvent(_ context.Context, _ *nostr.Event) error {
	qq.Println("DeleteEvent@noop")
	return nil
}

func (n *NoopStore) SaveEvent(_ context.Context, _ *nostr.Event) error {
	qq.Println("SaveEvent@noop")
	return nil
}

func (n *NoopStore) ReplaceEvent(_ context.Context, _ *nostr.Event) error {
	qq.Println("ReplaceEvent@noop")
	return nil
}
