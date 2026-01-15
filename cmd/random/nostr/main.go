package main

import (
	"fmt"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/buddyhub/nostrout"
	xutils "github.com/blue-monads/potatoverse/backend/utils"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/nbd-wtf/go-nostr"
)

func main() {

	alice, err := createNode("alice")
	if err != nil {
		panic(err)
	}

	bob, err := createNode("bob")
	if err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Second)

	alicePubKey := alice.GetPubkey()
	bobPubKey := bob.GetPubkey()

	err = alice.WriteEventRaw(nostr.Event{
		Kind:    nostr.KindHTTPAuth + 2,
		Content: "Hello, world from alice to bob",
		Tags: []nostr.Tag{
			{
				"p", bobPubKey,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	qq.Println("@alice wrote event")

	time.Sleep(2 * time.Second)

	err = bob.WriteEventRaw(nostr.Event{
		Kind:    nostr.KindHTTPAuth + 2,
		Content: "Hello, world from bob to alice",
		Tags: []nostr.Tag{
			{
				"p", alicePubKey,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	qq.Println("@bob wrote event")

	time.Sleep(10000 * time.Second)

	alice.Close()
	bob.Close()

}

func createNode(key string) (*nostrout.NostrRout, error) {

	pubkey, privkey, err := xutils.GenerateKeyPair(key)
	if err != nil {
		return nil, err
	}

	nostrRout := nostrout.New(nostrout.Options{
		SelfPubkey:  pubkey,
		SelfPrivkey: privkey,
		Handler: func(ev *nostr.Event) {
			fmt.Println("@event"+key, ev)
		},
		DefaultServers: nostrout.DefaultServers,
	})

	err = nostrRout.Run()
	if err != nil {
		return nil, err
	}

	return nostrRout, nil

}
