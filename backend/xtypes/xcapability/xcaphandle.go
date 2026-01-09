package xcapability

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
)

type XCapabilityHandle interface {
	GetModel() *dbmodels.SpaceCapability
	ParseCapToken(token string) (*signer.CapabilityClaim, error)
	ValidateCapToken(token string) (*signer.CapabilityClaim, error)
	GetOptions(target any) error
	GetOptionsAsLazyData() lazydata.LazyData
}
