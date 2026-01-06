package xutils

import (
	"crypto/sha256"

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
