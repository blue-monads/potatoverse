package signer

import (
	"testing"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
)

func TestSigner(t *testing.T) {

	signer := New([]byte("1234567890"))

	token, err := signer.SignAlt("salt", "fidexample")
	if err != nil {
		panic(err)
	}

	qq.Println("token: ", token)

	stoken, timestamp, err := signer.VerifyAlt("salt", token)
	if err != nil {
		panic(err)
	}

	qq.Println("stoken: ", stoken)
	qq.Println("timestamp: ", timestamp)

}
