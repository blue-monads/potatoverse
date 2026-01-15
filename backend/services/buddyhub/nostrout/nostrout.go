package nostrout

import (
	"context"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

/*

Nostr Resources:
	https://nostr-nips.com
	https://github.com/nostr-protocol/nips
	https://github.com/nbd-wtf/go-nostr
	https://nostrbook.dev/

*/

var (
	NoStrServerList = []string{
		"wss://relay.damus.io",
		"wss://relay.nostr.band",
		"wss://cache1.primal.net",
		"wss://relay.bitcoiner.social",
		"wss://relay.current.fyi",
		"wss://relay.nos.social",
		"wss://relay.nostr.inosta.cc",
		"wss://relay.nostr.pub",
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

type NostrRout struct {
	opt Options

	hexPrivateKey string
	hexPublicKey  string

	writeChan chan *nostr.Event
	relays    []*nostr.Relay
}

type Options struct {
	SelfPubkey     string
	SelfPrivkey    string
	Handler        func(ev *nostr.Event)
	DefaultServers []string
}

func New(opt Options) *NostrRout {

	_, privval, err := nip19.Decode(opt.SelfPrivkey)
	if err != nil {
		panic(err)
	}

	privkey := privval.(string)

	_, pubval, err := nip19.Decode(opt.SelfPubkey)
	if err != nil {
		panic(err)
	}

	pubkey := pubval.(string)

	return &NostrRout{
		opt:           opt,
		writeChan:     make(chan *nostr.Event),
		relays:        make([]*nostr.Relay, 0, len(opt.DefaultServers)),
		hexPrivateKey: privkey,
		hexPublicKey:  pubkey,
	}
}

func (o *NostrRout) Run() error {
	return o.runLoop()
}

func (o *NostrRout) runLoop() error {

	ctx := context.Background()

	_, pubval, err := nip19.Decode(o.opt.SelfPubkey)
	if err != nil {
		return err
	}

	selfPubkey := pubval.(string)

	filters := nostr.Filters{{
		Kinds:   []int{nostr.KindTextNote},
		Authors: []string{selfPubkey},
	}}

	relays := make([]*nostr.Relay, 0, len(DefaultServers))

	for _, server := range DefaultServers {
		relay, err := nostr.RelayConnect(ctx, server)
		if err != nil {

			qq.Println("@error", err)
			continue

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

func (o *NostrRout) handleEvent(sub *nostr.Subscription) {
	defer sub.Close()
	for ev := range sub.Events {
		o.opt.Handler(ev)
	}
}

func (o *NostrRout) WriteEventRaw(ev nostr.Event) error {

	err := ev.Sign(o.hexPrivateKey)
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

func (o *NostrRout) Close() {
	for _, relay := range o.relays {
		relay.Close()
	}
}

func (o *NostrRout) GetPubkey() string {
	return o.hexPublicKey
}

func (o *NostrRout) GetPrivkey() string {
	return o.hexPrivateKey
}

func (o *NostrRout) GetOptions() *Options {
	return &o.opt
}
