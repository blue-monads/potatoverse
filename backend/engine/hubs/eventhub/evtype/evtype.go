package evtype

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

type TExecution struct {
	Subscription *dbmodels.MQSubscription
	Target       *dbmodels.MQEventTarget
	Event        *dbmodels.MQEvent
	RetryAble    bool
}

type Handler func(ex TExecution) error

type Builder func(app xtypes.App) Handler

type DataHandle interface {
	GetMQSynk() datahub.MQSynk
	GetSpaceOps() datahub.SpaceOps
}
