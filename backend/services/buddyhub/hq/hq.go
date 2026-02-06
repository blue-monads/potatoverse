package hq

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/blue-monads/potatoverse/backend/utils/nostrutils"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

// hq for buddy discovery

const (
	KindPotato = nostr.KindHTTPAuth + 2
)

var (
	ErrNoBuddyInfo = errors.New("no buddy info found")
)

type HQ struct {
	opts Options

	hexPrivateKey string
	hexPublicKey  string
	logger        *slog.Logger
	relays        []*nostr.Relay
}

type Options struct {
	Servers    []string
	PrivateKey string
	PublicKey  string
	Logger     *slog.Logger
}

func New(opt Options) (*HQ, error) {
	// Decode keys if they're nip19 encoded
	hexPrivKey, err := nostrutils.DecodeKeyToHex(opt.PrivateKey)
	if err != nil {
		return nil, err
	}

	hexPubKey, err := nostrutils.DecodeKeyToHex(opt.PublicKey)
	if err != nil {
		return nil, err
	}

	hq := &HQ{
		opts: opt,

		hexPrivateKey: hexPrivKey,
		hexPublicKey:  hexPubKey,
		logger:        opt.Logger,
		relays:        make([]*nostr.Relay, 0),
	}

	return hq, nil
}

func (hq *HQ) Start() error {

	// Connect to relays
	ctx := context.Background()
	for _, relayURL := range hq.opts.Servers {
		relay, err := nostr.RelayConnect(ctx, relayURL)
		if err != nil {
			hq.opts.Logger.Warn("Failed to connect to relay", "relay", relayURL, "err", err)
			continue
		}
		hq.relays = append(hq.relays, relay)
		hq.opts.Logger.Info("Connected to relay", "relay", relayURL)
	}

	if len(hq.relays) == 0 {
		return errors.New("failed to connect to any relay")
	}

	return nil

}

func (h *HQ) PublishSelfAddress(info *SelfInfo) error {
	content, err := json.Marshal(info)
	if err != nil {
		return err
	}

	ev := nostr.Event{
		PubKey:    h.hexPublicKey,
		Kind:      nostr.KindTextNote,
		Content:   string(content),
		CreatedAt: nostr.Now(),
		Tags: []nostr.Tag{
			{
				"p", h.hexPublicKey,
			},
		},
	}

	err = ev.Sign(h.hexPrivateKey)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Publish to all connected relays
	for _, relay := range h.relays {
		err = relay.Publish(ctx, ev)
		if err != nil {
			h.logger.Warn("Failed to publish to relay", "relay", relay.URL, "err", err)
			continue
		}
	}

	return nil
}

func (h *HQ) FindBuddyInfo(pubkey string) (*SelfInfo, error) {
	// Decode pubkey if it's nip19 encoded
	hexPubKey := pubkey
	_, pubval, err := nip19.Decode(pubkey)
	if err == nil {
		hexPubKey = pubval.(string)
	}

	filter := nostr.Filter{
		Kinds:   []int{nostr.KindTextNote},
		Authors: []string{hexPubKey},
		Limit:   1,
	}

	ctx := context.Background()

	var latestEvent *nostr.Event

	// Try to query from connected relays
	for _, relay := range h.relays {
		events, err := relay.QuerySync(ctx, filter)
		if err != nil {
			h.logger.Warn("Failed to query relay", "relay", relay.URL, "err", err)
			continue
		}

		if len(events) > 0 {
			// Get the most recent event
			latestEvent = events[0]
			for _, ev := range events {
				if ev.CreatedAt > latestEvent.CreatedAt {
					latestEvent = ev
				}
			}
			break
		}
	}

	if latestEvent == nil {
		return nil, ErrNoBuddyInfo
	}

	var info SelfInfo
	err = json.Unmarshal([]byte(latestEvent.Content), &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (h *HQ) Close() {
	for _, relay := range h.relays {
		relay.Close()
	}
	h.relays = nil
}
