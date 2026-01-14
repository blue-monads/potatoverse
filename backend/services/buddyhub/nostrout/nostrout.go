package nostrout

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
)

var (
	NoStrServerList = []string{
		"wss://relay.damus.io",
		"wss://relay.nostr.band",
		"wss://cache1.primal.net",
		"wss://relay.bitcoiner.social",
		"wss://nostr.mutinywallet.com",
		"wss://relay.current.fyi",
		"wss://relay.nos.social",
		"wss://relay.nostr.inosta.cc",
		"wss://relay.nostr.pub",
		"wss://nostr.rocks",
		"wss://relay.nostr.info",
		"wss://relay.nostrich.de",
		"wss://relay.snort.social",
		"wss://relay.wellorder.net",
		"wss://wot.nostr.party",
	}

	DefaultServers = []string{
		NoStrServerList[9],
		NoStrServerList[10],
		NoStrServerList[4],
	}
)

const (
	KindPotato = nostr.KindHTTPAuth + 2
)

type OutControl struct {
	selfPubkey string
	writeChan  chan *nostr.Event
	relays     []*nostr.Relay
	handler    func(ev *nostr.Event)
}

func NewOutControl(selfPubkey string, handler func(ev *nostr.Event)) *OutControl {
	return &OutControl{
		selfPubkey: selfPubkey,
		writeChan:  make(chan *nostr.Event),
		relays:     make([]*nostr.Relay, 0, len(DefaultServers)),
		handler:    handler,
	}
}

func (o *OutControl) Run() error {
	return o.runLoop()
}

func (o *OutControl) runLoop() error {

	ctx := context.Background()

	filters := nostr.Filters{{
		Kinds:   []int{nostr.KindTextNote},
		Authors: []string{o.selfPubkey},
	}}

	relays := make([]*nostr.Relay, 0, len(DefaultServers))

	for _, server := range DefaultServers {
		relay, err := nostr.RelayConnect(ctx, server)
		if err != nil {
			return err
		}
		sub, err := relay.Subscribe(ctx, filters)
		if err != nil {
			return err
		}
		go o.handleEvent(sub)
		relays = append(relays, relay)
	}

	o.relays = relays

	return nil

}

func (o *OutControl) handleEvent(sub *nostr.Subscription) {
	defer sub.Close()
	for ev := range sub.Events {
		o.handler(ev)
	}
}

func (o *OutControl) WriteEventRaw(ev nostr.Event) error {

	err := ev.Sign(o.selfPubkey)
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, relay := range o.relays {
		err := relay.Publish(ctx, ev)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *OutControl) Close() {
	for _, relay := range o.relays {
		relay.Close()
	}
}
