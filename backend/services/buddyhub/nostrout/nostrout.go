package nostrout

import (
	"context"
	"math/rand"
	"slices"
	"sync"

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
		"wss://nos.lol",
		// "wss://relay.nos.social", kind not allowed
		"wss://relay.nostr.inosta.cc",
		"wss://relay.nostr.pub",
		"wss://relay.nostr.info",
		"wss://relay.nostrich.de",
		"wss://relay.snort.social",
		"wss://relay.wellorder.net",
		"wss://nos.lol",
		"wss://nostr.bitcoiner.social",
		"wss://no.str.cr",
		"wss://nostr-dev.wellorder.net",
		"wss://nostr.einundzwanzig.space",
		"wss://nostr.middling.mydns.jp",
		"wss://nostr.mom",
		"wss://nostr.noones.com",
		"wss://nostr.oxtr.dev",
		"wss://nostr.slothy.win",
		"wss://nostr-verified.wellorder.net",
		"wss://nostr-verif.slothy.win",
		"wss://nostr.vulpem.com",
		"wss://relay.damus.io",
		"wss://relay.minds.com/nostr/v1/ws",
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

	filters := nostr.Filters{{
		Kinds:   []int{nostr.KindTextNote},
		Authors: []string{o.hexPublicKey},
	}}

	selectedServers := make([]string, 0, 7)
	selectedServers = append(selectedServers, DefaultServers...)

	for len(selectedServers) < 7 {
		server := NoStrServerList[rand.Intn(len(NoStrServerList))]
		if !slices.Contains(selectedServers, server) {
			selectedServers = append(selectedServers, server)
		}
	}

	relays := make([]*nostr.Relay, 0, len(selectedServers))

	relayChan := make(chan *nostr.Relay, len(selectedServers))
	wg := sync.WaitGroup{}

	for _, server := range selectedServers {

		wg.Add(1)

		go func() {
			defer wg.Done()

			relay, err := o.connect(ctx, server, filters)
			if err != nil {
				qq.Println("@error", err.Error())
				return
			}

			relayChan <- relay

		}()

	}

	wg.Wait()
	close(relayChan)

	for relay := range relayChan {
		relays = append(relays, relay)
	}

	o.relays = relays

	qq.Println("@connected to relays", len(relays))

	return nil

}

func (o *NostrRout) connect(ctx context.Context, relayServer string, filters nostr.Filters) (*nostr.Relay, error) {
	relay, err := nostr.RelayConnect(ctx, relayServer)
	if err != nil {
		return nil, err
	}

	sub, err := relay.Subscribe(ctx, filters)
	if err != nil {
		return nil, err
	}

	go o.handleEvent(sub)

	return relay, nil
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
			qq.Println("@error/relay", relay.URL, err.Error())
			continue
		}

		qq.Println("@success/relay", relay.URL)
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
