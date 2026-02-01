package nostrhq

import (
	"context"

	"github.com/fiatjaf/eventstore"
	"github.com/nbd-wtf/go-nostr"
)

type Relay struct {
}

func (r *Relay) Name() string {
	return "BasicRelay"
}

func (r *Relay) Storage(ctx context.Context) eventstore.Store {
	return &NoopStore{}
}

func (r *Relay) Init() error {

	return nil
}

func (r *Relay) AcceptEvent(ctx context.Context, evt *nostr.Event) (bool, string) {

	if evt.Kind >= 2000 && evt.Kind < 3000 {
		return true, ""
	}

	return false, "event kind not accepted"
}
