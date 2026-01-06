package xutils

import (
	"crypto/sha256"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip06"
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

	seed := nip06.SeedFromWords(words)
	pk, err := nip06.PrivateKeyFromSeed(seed)
	if err != nil {
		return "", "", err
	}

	pubkey, err := nostr.GetPublicKey(pk)
	if err != nil {
		return "", "", err
	}

	return pubkey, pk, nil
}
