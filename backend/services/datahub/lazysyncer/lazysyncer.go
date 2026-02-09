package lazysyncer

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/buddycdc"
	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/selfcdc"

	"github.com/upper/db/v4"
)

type Options struct {
	DbSession     db.Session
	IsSelfEnabled bool
	Buddies       []string
	BasePath      string
	Transport     datahub.BuddyTransport
}

type LazySyncer struct {
	cdcSyncer    *selfcdc.SelfCDCSyncer
	buddySyncers map[string]*buddycdc.BuddyCDC
	transport    datahub.BuddyTransport
}

func New(opts Options) *LazySyncer {
	cdcSyncer := selfcdc.NewSelfCDCSyncer(opts.DbSession, opts.IsSelfEnabled)

	ls := &LazySyncer{
		cdcSyncer: cdcSyncer,
	}

	buddySyncers := make(map[string]*buddycdc.BuddyCDC)
	for _, buddyId := range opts.Buddies {

		buddyCDC, err := buddycdc.NewBuddyCDC(buddycdc.Options{
			MainDb:      opts.DbSession,
			BasePath:    opts.BasePath,
			BuddyPubKey: buddyId,
			Transport:   NewBuddyAdapter(ls, buddyId),
		})
		if err != nil {
			return nil
		}

		buddySyncers[buddyId] = buddyCDC
	}

	ls.buddySyncers = buddySyncers

	return ls
}

func NewTest(opts Options) *LazySyncer {
	cdcSyncer := selfcdc.NewSelfCDCSyncer(opts.DbSession, opts.IsSelfEnabled)

	ls := &LazySyncer{
		cdcSyncer: cdcSyncer,
	}

	buddySyncers := make(map[string]*buddycdc.BuddyCDC)
	for _, buddyId := range opts.Buddies {

		buddyCDC, err := buddycdc.NewBuddyCDC(buddycdc.Options{
			MainDb:      opts.DbSession,
			BasePath:    opts.BasePath,
			BuddyPubKey: buddyId,
			Transport:   cdcSyncer,
		})
		if err != nil {
			return nil
		}

		buddySyncers[buddyId] = buddyCDC
	}

	ls.buddySyncers = buddySyncers

	return ls
}

func (l *LazySyncer) Start(transport datahub.BuddyTransport) error {

	l.transport = transport

	err := l.cdcSyncer.Start()
	if err != nil {
		return err
	}

	for _, buddyCDC := range l.buddySyncers {
		buddyCDC.Start()
	}

	return nil
}

func (l *LazySyncer) GetSelfCDCSyncer() *selfcdc.SelfCDCSyncer {
	return l.cdcSyncer
}

func (l *LazySyncer) GetBuddyCDC(buddyId string) *buddycdc.BuddyCDC {
	return l.buddySyncers[buddyId]
}
