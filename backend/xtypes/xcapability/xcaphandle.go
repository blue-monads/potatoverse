package xcapability

import (
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes/lazydata"
)

type XCapabilityHandle interface {
	GetModel() *dbmodels.SpaceCapability
	ParseCapToken(token string) (*signer.CapabilityClaim, error)
	ValidateCapToken(token string) (*signer.CapabilityClaim, error)
	GetOptionsAsLazyData() lazydata.LazyData
	GetOptionsAsMap() (map[string]any, error)
	GetOptions(target any) error
}
