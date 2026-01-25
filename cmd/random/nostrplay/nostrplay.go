package main

import (
	"context"
	"fmt"
	"time"

	"github.com/blue-monads/potatoverse/backend/utils/nostrutils"
	"github.com/nbd-wtf/go-nostr"
)

func main() {

	pubkey, privkey, err := nostrutils.GenerateKeyPair("alice")
	if err != nil {
		panic(err)
	}

	hexPrivKey, err := nostrutils.DecodeKeyToHex(privkey)
	if err != nil {
		panic(err)
	}

	hexPubKey, err := nostrutils.DecodeKeyToHex(pubkey)
	if err != nil {
		panic(err)
	}

	fmt.Println("@hexPrivKey", hexPrivKey)
	fmt.Println("@hexPubKey", hexPubKey)
	fmt.Println("@pubkey", pubkey)
	fmt.Println("@privkey", privkey)

	ctx := context.Background()
	relayURL := "wss://nos.lol"

	relay, err := nostr.RelayConnect(ctx, relayURL)
	if err != nil {
		panic(err)
	}

	ev := nostr.Event{
		PubKey:    hexPubKey,
		Kind:      nostr.KindTextNote,
		Content:   "Hello, world",
		CreatedAt: nostr.Now(),
		Tags: []nostr.Tag{
			{
				"p", hexPubKey,
			},
		},
	}

	err = ev.Sign(hexPrivKey)
	if err != nil {
		panic(err)
	}

	// publish event
	err = relay.Publish(ctx, ev)
	if err != nil {
		panic(err)
	}

	fmt.Println("@published event", ev.ID)

	time.Sleep(2 * time.Second)

	sub, err := relay.QuerySync(ctx, nostr.Filter{
		Kinds:   []int{nostr.KindTextNote},
		Authors: []string{hexPubKey},
	})
	if err != nil {
		panic(err)
	}

	for _, ev := range sub {
		fmt.Println("@event", ev.ID, ev.Content)
	}

}
