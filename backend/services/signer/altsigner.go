package signer

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
)

// donot use for anything too serious, it more as a way obfuscation
// hope you know what you are doing

func (b *Signer) deriveAltKey(salt string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(b.altKey))
	hasher.Write([]byte(salt))
	return hasher.Sum(nil)
}

const (
	NonceSize = 4
)

func (b *Signer) SignAlt(salt string, data string) (string, error) {
	return b.SignAltCore(b.deriveAltKey(salt), data)
}

func (b *Signer) VerifyAlt(salt string, token string) (string, int64, error) {
	return b.VerifyAltCore(b.deriveAltKey(salt), token)
}

func (b *Signer) SignAltBatch(salt string, data []string) ([]string, error) {

	tokens := make([]string, len(data))

	key := b.deriveAltKey(salt)

	for i := range data {
		token, err := b.SignAltCore(key, data[i])
		if err != nil {
			return nil, err
		}
		tokens[i] = token
	}

	return tokens, nil

}

func (b *Signer) VerifyAltBatch(salt string, tokens []string) ([]string, error) {
	data := make([]string, len(tokens))
	key := b.deriveAltKey(salt)

	for i := range tokens {
		d, _, err := b.VerifyAltCore(key, tokens[i])
		if err != nil {
			return nil, err
		}
		data[i] = d
	}

	return data, nil

}

func (b *Signer) SignAltCore(key []byte, data string) (string, error) {

	nonce := make([]byte, NonceSize+4)
	if _, err := rand.Read(nonce[:NonceSize]); err != nil {
		return "", err
	}

	timestamp := time.Now().Unix()
	binary.BigEndian.PutUint32(nonce[NonceSize:], uint32(timestamp))

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("error creating aes block cipher", err)
		return "", err
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, NonceSize+4)
	if err != nil {
		return "", err
	}

	plaintext := []byte(data)
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	token := append(nonce, ciphertext...)

	return base58.Encode(token), nil

}

func (b *Signer) VerifyAltCore(key []byte, token string) (string, int64, error) {

	tokenBytes := base58.Decode(token)

	nonce := tokenBytes[:NonceSize+4]
	ciphertext := tokenBytes[NonceSize+4:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", 0, err
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, NonceSize+4)
	if err != nil {
		return "", 0, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", 0, err
	}

	timeBytes := nonce[NonceSize:]
	timestamp := binary.BigEndian.Uint32(timeBytes)

	return string(plaintext), int64(timestamp), nil
}
