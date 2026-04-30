package xcapability

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
)

type XCapabilityHandle interface {
	GetSpaceId() (int64, error)
	GetModel() *dbmodels.SpaceCapability
	ParseCapToken(token string) (*signer.CapabilityClaim, error)
	ValidateCapToken(token string) (*signer.CapabilityClaim, error)
	GetOptions(target any) error
	GetOptionsAsLazyData() lazydata.LazyData
}
