package lazysyncer

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/buddycdc"
	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/selfcdc"

	"github.com/upper/db/v4"
)

type Options struct {
	DbSession     db.Session
	IsSelfEnabled bool
	Buddies       []string
	BasePath      string
}

type LazySyncer struct {
	cdcSyncer    *selfcdc.SelfCDCSyncer
	buddySyncers map[string]*buddycdc.BuddyCDC
}

func NewLazySyncer(opts Options) *LazySyncer {
	cdcSyncer := selfcdc.NewSelfCDCSyncer(opts.DbSession, opts.IsSelfEnabled)

	buddySyncers := make(map[string]*buddycdc.BuddyCDC)
	for _, buddyId := range opts.Buddies {

		buddyCDC, err := buddycdc.NewBuddyCDC(opts.DbSession, opts.BasePath, buddyId)
		if err != nil {
			return nil
		}

		buddySyncers[buddyId] = buddyCDC
	}

	return &LazySyncer{
		cdcSyncer:    cdcSyncer,
		buddySyncers: buddySyncers,
	}
}

func (l *LazySyncer) Start() error {
	return l.cdcSyncer.Start()
}

func (l *LazySyncer) GetSelfCDCSyncer() *selfcdc.SelfCDCSyncer {
	return l.cdcSyncer
}

func (l *LazySyncer) GetBuddyCDC(buddyId string) *buddycdc.BuddyCDC {
	return l.buddySyncers[buddyId]
}
