package nostrutils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

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
