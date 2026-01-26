package lazysyncer

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/selfcdc"
	"github.com/upper/db/v4"
)

type LazySyncer struct {
	cdcSyncer *selfcdc.SelfCDCSyncer
}

func NewLazySyncer(db db.Session, isEnabled bool) *LazySyncer {
	cdcSyncer := selfcdc.NewSelfCDCSyncer(db, isEnabled)

	return &LazySyncer{
		cdcSyncer: cdcSyncer,
	}
}

func (l *LazySyncer) Start() error {
	return l.cdcSyncer.Start()
}

func (l *LazySyncer) GetSelfCDCSyncer() *selfcdc.SelfCDCSyncer {
	return l.cdcSyncer
}
