package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

func main() {
	fmt.Println("Hello, World!")

	fistPart := "potato_asbasu66612jhagshasg___&aksa"
	secondPart := "verse_asbasu66612jhagshasg___&aksa"

	// calculate sha1 hash of "potato"
	hash := sha1.New()
	hash.Write([]byte(fistPart + secondPart))
	fmt.Println(hex.EncodeToString(hash.Sum(nil)))

	// calculate sha1 hash of "potatoverse"
	hash = sha1.New()
	hash.Write([]byte(fistPart))
	hash.Write([]byte(secondPart))
	fmt.Println(hex.EncodeToString(hash.Sum(nil)))

}
