package app

import (
	"crypto/sha256"
	"log/slog"
	"path"

	controller "github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/engine"
	"github.com/k0kubun/pp"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type Option struct {
	Database          datahub.Database
	Logger            *slog.Logger
	Signer            *signer.Signer
	AppOpts           *xtypes.AppOptions
	WorkingFolderBase string
}

var _ xtypes.App = (*HeadLess)(nil)

// headless means it has no http server attached to it
type HeadLess struct {
	db      datahub.Database
	signer  *signer.Signer
	logger  *slog.Logger
	ctrl    *controller.Controller
	AppOpts *xtypes.AppOptions
	engine  *engine.Engine
}

func NewHeadLess(opt Option) *HeadLess {

	engine := engine.NewEngine(opt.Database, path.Join(opt.WorkingFolderBase, "engine"))

	happ := &HeadLess{
		db:     opt.Database,
		signer: opt.Signer,
		logger: opt.Logger,
		ctrl: controller.New(controller.Option{
			Database: opt.Database,
			Logger:   opt.Logger,
			Signer:   opt.Signer,
			AppOpts:  opt.AppOpts,
			Engine:   engine,
		}),
		engine:  engine,
		AppOpts: opt.AppOpts,
	}

	return happ
}

func (h *HeadLess) Init() error {

	h.logger.Info("Initializing HeadLess application")

	return nil
}

func (h *HeadLess) Start() error {

	err := h.engine.Start(h)
	if err != nil {
		return err
	}

	h.logger.Info("HeadLess application started")

	has, err := h.ctrl.HasFingerprint()
	if err != nil {
		return err
	}

	pp.Println(h.AppOpts)

	// sha256 hash of the master secret
	shash := hashMasterSecret(h.AppOpts.MasterSecret)

	if !has {
		fingerPrint := &controller.AppFingerPrint{
			Version:          "0.1.0",
			Commit:           "unknown",
			BuildAt:          "unknown",
			MasterSecretHash: shash,
		}

		err = h.ctrl.SetAppFingerPrint(fingerPrint)
		if err != nil {
			h.logger.Error("Failed to set app fingerprint", "err", err)
			return err
		}

		h.logger.Info("App fingerprint set", "fingerprint", fingerPrint)

	}

	oldFingerPrint, err := h.ctrl.GetAppFingerPrint()
	if err != nil {
		h.logger.Error("Failed to get app fingerprint", "err", err)
		return err
	}

	if oldFingerPrint.MasterSecretHash != shash {
		h.logger.Warn("Master secret hash has changed, updating fingerprint")
	}

	return nil
}

// shared methods for HeadLess

func (h *HeadLess) Database() datahub.Database {
	return h.db
}

func (h *HeadLess) Signer() *signer.Signer {
	return h.signer
}

func (h *HeadLess) Logger() *slog.Logger {
	return h.logger
}

func (h *HeadLess) Controller() any {
	return h.ctrl
}

func (h *HeadLess) Engine() *engine.Engine {
	return h.engine
}

// private

func hashMasterSecret(masterSecret string) string {
	h := sha256.New()
	h.Write([]byte("SALT_FINGERPRINT"))
	h.Write([]byte(masterSecret))

	return string(h.Sum(nil))
}
