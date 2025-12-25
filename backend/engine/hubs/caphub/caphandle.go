package caphub

import (
	"encoding/json"
	"errors"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/utils/kosher"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/lazydata"
	"github.com/blue-monads/turnix/backend/xtypes/xcapability"
)

var _ xcapability.XCapabilityHandle = (*CapabilityHandle)(nil)

type CapabilityHandle struct {
	model  *dbmodels.SpaceCapability
	signer *signer.Signer
	app    xtypes.App
}

func NewCapabilityHandle(app xtypes.App, model *dbmodels.SpaceCapability) *CapabilityHandle {
	return &CapabilityHandle{
		model:  model,
		signer: app.Signer(),
		app:    app,
	}
}

func (h *CapabilityHandle) GetModel() *dbmodels.SpaceCapability {
	return h.model
}

func (h *CapabilityHandle) ParseCapToken(token string) (*signer.CapabilityClaim, error) {
	return h.signer.ParseCapability(token)
}

func (h *CapabilityHandle) ValidateCapToken(token string) (*signer.CapabilityClaim, error) {
	claim, err := h.signer.ParseCapability(token)
	if err != nil {
		return nil, err
	}

	if claim.SpaceId != h.model.SpaceID {
		return nil, errors.New("invalid space id")
	}

	if claim.InstallId != h.model.InstallID {
		return nil, errors.New("invalid install id")
	}

	if claim.CapabilityId != h.model.ID {
		return nil, errors.New("invalid capability id")
	}

	return claim, nil
}

func (h *CapabilityHandle) GetOptionsAsLazyData() lazydata.LazyData {
	return lazydata.LazyDataBytes(kosher.Byte(h.model.Options))
}

func (h *CapabilityHandle) GetOptionsAsMap() (map[string]any, error) {
	var result map[string]any
	err := json.Unmarshal(kosher.Byte(h.model.Options), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (h *CapabilityHandle) GetOptions(target any) error {
	return json.Unmarshal(kosher.Byte(h.model.Options), target)
}
