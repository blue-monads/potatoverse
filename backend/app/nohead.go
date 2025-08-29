package app

import (
	"crypto/sha256"
	"log/slog"

	"github.com/blue-monads/turnix/backend/app/controller"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type Option struct {
	Database datahub.Database
	Logger   *slog.Logger
	Signer   *signer.Signer
	AppOpts  *xtypes.AppOptions
}

// headless means it has no http server attached to it
type HeadLess struct {
	db      datahub.Database
	signer  *signer.Signer
	logger  *slog.Logger
	ctrl    *controller.Controller
	AppOpts *xtypes.AppOptions
}

func NewHeadLess(opt Option) *HeadLess {
	return &HeadLess{
		db:     opt.Database,
		signer: opt.Signer,
		logger: opt.Logger,
		ctrl: controller.New(controller.Option{
			Database: opt.Database,
			Logger:   opt.Logger,
			Signer:   opt.Signer,
			AppOpts:  opt.AppOpts,
		}),
	}
}

func (h *HeadLess) Init() error {

	h.logger.Info("Initializing HeadLess application")

	return nil
}

func (h *HeadLess) Start() error {
	h.logger.Info("HeadLess application started")

	has, err := h.ctrl.HasFingerprint()
	if err != nil {
		return err
	}

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

func (h *HeadLess) Controller() *controller.Controller {
	return h.ctrl
}

// private

func hashMasterSecret(masterSecret string) string {
	h := sha256.New()
	h.Write([]byte("SALT_FINGERPRINT"))
	h.Write([]byte(masterSecret))

	return string(h.Sum(nil))
}
