package signer

import (
	"crypto/sha256"
	"encoding/json"
	"errors"

	"github.com/hako/branca"
	"golang.org/x/crypto/pbkdf2"
)

const (
	TokenTypeAccess      uint8 = 1
	TokenTypeEmailInvite uint8 = 2
	TokenTypePair        uint8 = 3
	TokenTypeAdvisiery   uint8 = 4
)

type AccessClaim struct {
	XID             string         `json:"x,omitempty"`
	Typeid          uint8          `json:"t,omitempty"`
	UserId          int64          `json:"u,omitempty"`
	PinnedProjectId int64          `json:"p,omitempty"`
	Extrameta       map[string]any `json:"e,omitempty"`
}

type InviteClaim struct {
	XID            string `json:"x,omitempty"`
	Typeid         uint8  `json:"t,omitempty"`
	InviteFromUser int64  `json:"u,omitempty"`
	InviteToUser   int64  `json:"z,omitempty"`
	InviteEmail    string `json:"e,omitempty"`
	ProjectId      int64  `json:"p,omitempty"`
}

type AdvisieryClaimE struct {
	XID       string `json:"x,omitempty"`
	Typeid    uint8  `json:"t,omitempty"`
	ProjectId int64  `json:"p,omitempty"`
	Data      any    `json:"d,omitempty"`
}

type AdvisieryClaimD struct {
	XID       string          `json:"x,omitempty"`
	Typeid    uint8           `json:"t,omitempty"`
	ProjectId int64           `json:"p,omitempty"`
	Data      json.RawMessage `json:"d,omitempty"`
}

// fixme => add expiry

var ErrInvalidToken = errors.New("INVALID TOKEN")

type Signer struct {
	signer *branca.Branca
}

func New(key []byte) *Signer {
	masterKey := pbkdf2.Key(key, []byte("SALTY_SALMON"), 2048, 32, sha256.New)

	return &Signer{
		signer: branca.NewBranca(string(masterKey)),
	}
}

func (t *Signer) parse(token string, dest any) error {
	str, err := t.signer.DecodeToString(token)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(str), dest)
}

func (t *Signer) sign(o any) (string, error) {
	out, err := json.Marshal(o)
	if err != nil {
		return "", nil
	}

	return t.signer.EncodeToString(string(out))
}

func (ts *Signer) ParseAccess(tstr string) (*AccessClaim, error) {

	claim := &AccessClaim{}

	err := ts.parse(tstr, claim)
	if err != nil {
		return nil, err
	}

	if claim.Typeid != TokenTypeAccess {
		return nil, ErrInvalidToken
	}

	return claim, nil
}

func (ts *Signer) SignAccess(claim *AccessClaim) (string, error) {

	claim.Typeid = TokenTypeAccess

	return ts.sign(claim)
}

func (ts *Signer) ParseInvite(tstr string) (*InviteClaim, error) {

	claim := &InviteClaim{}

	err := ts.parse(tstr, claim)
	if err != nil {
		return nil, err
	}

	if claim.Typeid != TokenTypeEmailInvite {
		return nil, ErrInvalidToken
	}

	return claim, nil
}

func (ts *Signer) SignInvite(claim *InviteClaim) (string, error) {

	claim.Typeid = TokenTypeEmailInvite

	return ts.sign(claim)
}

func (ts *Signer) ParseProjectAdvisiery(pid int64, tstr string) ([]byte, error) {

	claim := &AdvisieryClaimD{}

	err := ts.parse(tstr, claim)
	if err != nil {
		return nil, err
	}

	if claim.ProjectId != pid {
		return nil, ErrInvalidToken
	}

	if claim.Typeid != TokenTypeAdvisiery {
		return nil, ErrInvalidToken
	}

	return claim.Data, nil
}

func (ts *Signer) SignProjectAdvisiery(pid int64, data any) (string, error) {
	claim := &AdvisieryClaimE{
		ProjectId: pid,
		Data:      data,
		Typeid:    TokenTypeAdvisiery,
	}

	return ts.sign(claim)
}
