package xutils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip06"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/tyler-smith/go-bip39"
)

func GenerateKeyPair(masterSecret string) (string, string, error) {

	h := sha256.New()
	h.Write([]byte(masterSecret))
	h.Write([]byte("SALTY_POTATO_NODE_ID"))
	entropy := h.Sum(nil)

	words, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", "", err
	}

	// qq.Println("@words", string(words))

	seed := nip06.SeedFromWords(words)
	pk, err := nip06.PrivateKeyFromSeed(seed)
	if err != nil {
		return "", "", err
	}

	pubkey, err := nostr.GetPublicKey(pk)
	if err != nil {
		return "", "", err
	}

	pubkeyBase64, err := nip19.EncodePublicKey(pubkey)
	if err != nil {
		return "", "", err
	}
	privkeyBase64, err := nip19.EncodePrivateKey(pk)
	if err != nil {
		return "", "", err
	}

	// qq.Println("@pubkey", pubkey)
	// qq.Println("@privkey", pk)

	// qq.Println("@pubkeyBase64", pubkeyBase64)
	// qq.Println("@privkeyBase64", privkeyBase64)

	return pubkeyBase64, privkeyBase64, nil
}

func VerifyNostrAuth(authHeader string) (*nostr.Event, error) {

	eventJson, err := base64.StdEncoding.DecodeString(authHeader)
	if err != nil {
		return nil, fmt.Errorf("Invalid authorization header")
	}

	var event nostr.Event
	err = json.Unmarshal(eventJson, &event)
	if err != nil {
		return nil, fmt.Errorf("Invalid authorization header")
	}

	ok, err := event.CheckSignature()
	if !ok || err != nil {
		return nil, fmt.Errorf("invalid signature")
	}

	if event.Kind != nostr.KindHTTPAuth {
		return nil, fmt.Errorf("wrong event kind")
	}

	return &event, nil
}

func GenerateNostrAuthToken(privkey string, serverURL, method string) (string, error) {

	event := nostr.Event{
		Kind: nostr.KindHTTPAuth,
		Tags: nostr.Tags{{"u", serverURL}, {"m", method}},
	}

	err := event.Sign(privkey)
	if err != nil {
		return "", err
	}

	ejson, err := json.Marshal(event)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(ejson)

	return encoded, nil

}
