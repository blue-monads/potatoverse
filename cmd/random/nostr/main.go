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
		qq.Println("@main_error/1", err.Error())
		panic(err)
	}

	bob, err := createNode("bob")
	if err != nil {
		qq.Println("@main_error/2", err.Error())
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
		CreatedAt: nostr.Now(),
	})

	if err != nil {
		qq.Println("@main_error/3", err.Error())
		panic(err)
	}

	qq.Println("@alice wrote event")

	time.Sleep(2 * time.Second)

	rev, err := bob.WriteEventWithResponse(nostr.Event{
		Kind:    nostr.KindHTTPAuth + 2,
		Content: "Hello, world from bob to alice",
		Tags: []nostr.Tag{
			{
				"p", alicePubKey,
				"t", "custom_subtype",
			},
		},
		CreatedAt: nostr.Now(),
	})

	if err != nil {
		qq.Println("@main_error/4", err.Error())
		panic(err)
	}

	if rev != nil {
		fmt.Println("@bob received event", rev.ID, rev.Content)
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

	var nostrRout *nostrout.NostrRout

	nostrRout = nostrout.New(nostrout.Options{
		SelfPubkey:  pubkey,
		SelfPrivkey: privkey,
		Handler: func(ev *nostr.Event) {

			qq.Println("----------@WOWOW@-------------")
			fmt.Println("@event"+key, ev.ID, ev.Content)

			fmt.Println("@tags", ev.Tags, len(ev.Tags))

			currEtags := ev.Tags.Find("e")

			qq.Println("@currEtags", len(currEtags))

			if len(currEtags) > 0 {
				for _, etag := range currEtags {
					qq.Println("@e", etag)
				}

				qq.Println("@don't reply to reply", ev.ID)
				return
			}

			evid := ev.ID

			pIds := ev.Tags.FindLast("p")
			if len(pIds) == 0 {
				qq.Println("@sender not found")
				return
			}

			pId := pIds[1]

			qq.Println("@writing reply to event", evid, pId)

			nostrRout.WriteEventRaw(nostr.Event{
				Kind:    nostr.KindHTTPAuth + 2,
				Content: "reply to event by " + key,
				Tags: []nostr.Tag{
					{
						"p", pId,
					},
					{
						"e", evid,
					},
				},
				CreatedAt: nostr.Now(),
			})

			if err != nil {
				qq.Println("@main_error/5", err.Error())
				panic(err)
			}

		},
		DefaultServers: nostrout.DefaultServers,
	})

	err = nostrRout.Run()
	if err != nil {
		return nil, err
	}

	return nostrRout, nil

}

/*

{
	"kind":27237,
	"id":"f3ab53f80d338f068a3a22347309f8f67b76d782df02fc2281e98374613c473f",
	"pubkey":"20ae37404605417a0cf8d16a52bfff648e898764c1810115ddab846d04e5d21e",
	"created_at":1768463844,
	"tags":[["p","c29eee08ece09b02bccfb5a5b96225361443fc7bb5e8a1ec0c45946ee460e187"]],
	"content":"Hello, world from bob to alice",
	"sig":"1807268ca52134128174c68899f157cef952f813783fdda45afa149c152d12ba4fce00e78edf3a08304b3dd8bfba1e75a5b79d729b93d06d327113ddb64ee5c1"
}

*/
