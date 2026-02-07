package xtypes

import (
	"net/http"
)

type BuddyTransport interface {
	SendBuddy(buddyPubkey string, req *http.Request) (*http.Response, error)
}
