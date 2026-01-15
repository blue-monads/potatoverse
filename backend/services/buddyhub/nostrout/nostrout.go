package nostrout

import (
	"context"
	"errors"
	"math/rand"
	"slices"
	"sync"
	"time"

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
		"wss://nostr.middling.mydns.jp",
		"wss://nostr.mom",
		"wss://nostr.noones.com",
		"wss://nostr.oxtr.dev",
		"wss://nostr.slothy.win",
		"wss://nostr-verified.wellorder.net",
		"wss://nostr.vulpem.com",
		"wss://relay.damus.io",
		"wss://relay.minds.com/nostr/v1/ws",
	}

	DefaultServers = []string{
		//		"wss://proxy.nostr-relay.app/ac0805e2c2d5ad533d76967da021440d3f9da5308692c8ab78b5f90995740305",
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

	processedEvents      map[string]bool
	processedEventsMutex sync.RWMutex
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

		processedEvents:      make(map[string]bool),
		processedEventsMutex: sync.RWMutex{},
	}
}

func (o *NostrRout) Run() error {
	return o.runLoop()
}

func (o *NostrRout) runLoop() error {

	ctx := context.Background()

	filters := nostr.Filters{{
		Kinds: []int{KindPotato},
		Tags: map[string][]string{
			"p": {o.hexPublicKey},
		},
	}}

	selectedServers := make([]string, 0, 7)
	selectedServers = append(selectedServers, DefaultServers...)

	for len(selectedServers) < 5 {
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
	defer func() {
		sub.Unsub()
		sub.Close()
		qq.Println("@closed/subscription")
	}()

	for ev := range sub.Events {

		eTags := ev.Tags.Find("e")
		if len(eTags) > 0 {
			qq.Println("@skipping_event_subscribed/event_reply", ev.ID)
			continue
		}

		o.processedEventsMutex.RLock()
		if o.processedEvents[ev.ID] {
			o.processedEventsMutex.RUnlock()
			qq.Println("@skipping_event_subscribed/event_already_processed", ev.ID)
			continue
		}
		o.processedEventsMutex.RUnlock()

		// double check
		o.processedEventsMutex.Lock()
		recheck := o.processedEvents[ev.ID]
		if recheck {
			o.processedEventsMutex.Unlock()
			qq.Println("@skipping_event_subscribed/event_already_processed", ev.ID)
			continue
		}

		o.processedEvents[ev.ID] = true
		o.processedEventsMutex.Unlock()

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

var (
	ErrNoResponse = errors.New("no response")
)

func (o *NostrRout) WriteEventWithResponse(ev nostr.Event) (*nostr.Event, error) {

	err := ev.Sign(o.hexPrivateKey)
	if err != nil {
		return nil, err
	}

	filters := nostr.Filters{{
		Kinds: []int{KindPotato},
		Tags: map[string][]string{
			"e": {ev.ID},
		},
	}}

	if len(o.relays) == 0 {
		return nil, ErrNoResponse
	}

	batchSize := 2
	timeoutSecs := 5

	// Try relays in batches of 2, looping until all are exhausted
	for i := 0; i < len(o.relays); i += batchSize {
		end := min(i+batchSize, len(o.relays))

		batch := o.relays[i:end]
		qq.Println("@trying_batch", i, end, "relays:", len(batch))

		resp, err := o.tryRelaysForResponse(ev, filters, batch, timeoutSecs)
		if err == nil && resp != nil {
			return resp, nil
		}

		qq.Println("@no_response_from_batch", i, end)
	}

	return nil, ErrNoResponse
}

func (o *NostrRout) tryRelaysForResponse(ev nostr.Event, filters nostr.Filters, relays []*nostr.Relay, timeoutSecs int) (*nostr.Event, error) {
	responseChan := make(chan *nostr.Event, 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	for _, relay := range relays {
		go func(r *nostr.Relay) {

			sub, err := r.Subscribe(ctx, filters)
			if err != nil {
				qq.Println("@error/relay/subscribe", r.URL, err.Error())
				return
			}
			defer sub.Unsub()

			err = r.Publish(ctx, ev)
			if err != nil {
				qq.Println("@error/relay/publish", r.URL, err.Error())
				return
			}

			qq.Println("@published/relay", r.URL)

			for {
				select {
				case respEv := <-sub.Events:
					select {
					case responseChan <- respEv:
					default:
					}
					return
				case <-ctx.Done():
					return
				}
			}
		}(relay)
	}

	select {
	case resp := <-responseChan:
		return resp, nil
	case <-ctx.Done():
		return nil, ErrNoResponse
	}
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
