package signer

import (
	"crypto/sha256"
	"encoding/json"
	"errors"

	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/hako/branca"
	"golang.org/x/crypto/pbkdf2"
)

const (
	TokenTypeAccess             uint16 = 1
	TokenTypeEmailInvite        uint16 = 2
	TokenTypePair               uint16 = 3
	TokenTypeSpace              uint16 = 4
	TokenTypeSpaceAdvisiery     uint16 = 5
	TokenTypeSpaceFilePresigned uint16 = 6
	ToekenPackageDev            uint16 = 7
	TokenTypeCapability         uint16 = 8
	TokenTypeBuddyAuth          uint16 = 9
)

type AccessClaim struct {
	Typeid    uint16         `json:"t,omitempty"`
	UserId    int64          `json:"u,omitempty"`
	Extrameta map[string]any `json:"e,omitempty"`
}

type InviteClaim struct {
	Typeid   uint16 `json:"t,omitempty"`
	InviteId int64  `json:"p,omitempty"`
}

type SpaceClaim struct {
	Typeid    uint16 `json:"t,omitempty"`
	SpaceId   int64  `json:"s,omitempty"`
	UserId    int64  `json:"u,omitempty"`
	SessionId int64  `json:"i,omitempty"`
}

type SpaceAdvisieryClaim struct {
	Typeid       uint16         `json:"t,omitempty"`
	TokenSubType string         `json:"z,omitempty"`
	InstallId    int64          `json:"i,omitempty"`
	SpaceId      int64          `json:"s,omitempty"`
	UserId       int64          `json:"u,omitempty"`
	ResourceId   string         `json:"r,omitempty"`
	Data         map[string]any `json:"d,omitempty"`
}

type SpaceFilePresignedClaim struct {
	Typeid    uint16 `json:"t,omitempty"`
	InstallId int64  `json:"i,omitempty"`
	UserId    int64  `json:"u,omitempty"`
	PathName  string `json:"pn,omitempty"`
	FileName  string `json:"fn,omitempty"`
	Expiry    int64  `json:"e,omitempty"`
}

type PackageDevClaim struct {
	Typeid           uint16 `json:"t,omitempty"`
	InstallPackageId int64  `json:"p,omitempty"`
	UserId           int64  `json:"u,omitempty"`
}

type CapabilityClaim struct {
	Typeid       uint16         `json:"t,omitempty"`
	CapabilityId int64          `json:"c,omitempty"`
	InstallId    int64          `json:"i,omitempty"`
	SpaceId      int64          `json:"s,omitempty"`
	UserId       int64          `json:"u,omitempty"`
	ResourceId   string         `json:"r,omitempty"`
	ExtraMeta    map[string]any `json:"e,omitempty"`
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

func (ts *Signer) ParseSpace(tstr string) (*SpaceClaim, error) {

	claim := &SpaceClaim{}

	err := ts.parse(tstr, claim)
	if err != nil {
		return nil, err
	}

	if claim.Typeid != TokenTypeSpace {
		return nil, ErrInvalidToken
	}

	return claim, nil
}

func (ts *Signer) SignSpace(claim *SpaceClaim) (string, error) {

	claim.Typeid = TokenTypeSpace

	return ts.sign(claim)
}

func (ts *Signer) ParseSpaceAdvisiery(tstr string) (*SpaceAdvisieryClaim, error) {

	claim := &SpaceAdvisieryClaim{}

	err := ts.parse(tstr, claim)
	if err != nil {
		return nil, err
	}

	if claim.Typeid != TokenTypeSpaceAdvisiery {
		return nil, ErrInvalidToken
	}

	return claim, nil
}

func (ts *Signer) SignSpaceAdvisiery(claim *SpaceAdvisieryClaim) (string, error) {

	claim.Typeid = TokenTypeSpaceAdvisiery

	return ts.sign(claim)
}

func (ts *Signer) ParseSpaceFilePresigned(tstr string) (*SpaceFilePresignedClaim, error) {

	claim := &SpaceFilePresignedClaim{}

	err := ts.parse(tstr, claim)
	if err != nil {
		return nil, err
	}

	if claim.Typeid != TokenTypeSpaceFilePresigned {
		return nil, ErrInvalidToken
	}

	return claim, nil
}

func (ts *Signer) SignSpaceFilePresigned(claim *SpaceFilePresignedClaim) (string, error) {

	claim.Typeid = TokenTypeSpaceFilePresigned

	return ts.sign(claim)
}

func (ts *Signer) ParsePackageDev(tstr string) (*PackageDevClaim, error) {

	claim := &PackageDevClaim{}

	err := ts.parse(tstr, claim)
	if err != nil {
		return nil, err
	}

	if claim.Typeid != ToekenPackageDev {
		return nil, ErrInvalidToken
	}

	return claim, nil
}

func (ts *Signer) SignPackageDev(claim *PackageDevClaim) (string, error) {

	claim.Typeid = ToekenPackageDev

	return ts.sign(claim)
}

func (ts *Signer) ParseCapability(tstr string) (*CapabilityClaim, error) {

	claim := &CapabilityClaim{}

	err := ts.parse(tstr, claim)
	if err != nil {
		return nil, err
	}

	if claim.Typeid != TokenTypeCapability {
		qq.Println("claim: ", claim)
		return nil, ErrInvalidToken
	}

	return claim, nil
}

func (ts *Signer) SignCapability(claim *CapabilityClaim) (string, error) {

	claim.Typeid = TokenTypeCapability

	return ts.sign(claim)
}
