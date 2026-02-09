package lazysyncer

import (
	"log/slog"

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
	Logger        *slog.Logger
}

type LazySyncer struct {
	cdcSyncer    *selfcdc.SelfCDCSyncer
	buddySyncers map[string]*buddycdc.BuddyCDC
	transport    datahub.BuddyTransport
	logger       *slog.Logger
}

func New(opts Options) *LazySyncer {
	selfLogger := opts.Logger.With("module", "selfsyncer")
	cdcSyncer := selfcdc.NewSelfCDCSyncer(opts.DbSession, selfLogger, opts.IsSelfEnabled)

	ls := &LazySyncer{
		cdcSyncer: cdcSyncer,
		logger:    opts.Logger,
	}

	buddySyncers := make(map[string]*buddycdc.BuddyCDC)
	for _, buddyId := range opts.Buddies {

		buddyCDC, err := buddycdc.NewBuddyCDC(buddycdc.Options{
			MainDb:      opts.DbSession,
			BasePath:    opts.BasePath,
			BuddyPubKey: buddyId,
			Transport:   NewBuddyAdapter(ls, buddyId),
			Logger:      opts.Logger.With("module", "buddysyncer"),
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

	selfLogger := opts.Logger.With("module", "selfsyncer")
	cdcSyncer := selfcdc.NewSelfCDCSyncer(opts.DbSession, selfLogger, opts.IsSelfEnabled)

	ls := &LazySyncer{
		cdcSyncer: cdcSyncer,
		logger:    opts.Logger,
	}

	buddySyncers := make(map[string]*buddycdc.BuddyCDC)
	for _, buddyId := range opts.Buddies {

		buddyCDC, err := buddycdc.NewBuddyCDC(buddycdc.Options{
			MainDb:      opts.DbSession,
			BasePath:    opts.BasePath,
			BuddyPubKey: buddyId,
			Transport:   cdcSyncer,
			Logger:      opts.Logger.With("module", "buddysyncer"),
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
		err = buddyCDC.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *LazySyncer) GetSelfCDCSyncer() *selfcdc.SelfCDCSyncer {
	return l.cdcSyncer
}

func (l *LazySyncer) GetBuddyCDC(buddyId string) *buddycdc.BuddyCDC {
	return l.buddySyncers[buddyId]
}
