package app

import (
	"crypto/sha256"
	"log/slog"
	"path"

	"github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/app/server"
	"github.com/blue-monads/turnix/backend/engine"
	"github.com/blue-monads/turnix/backend/utils/qq"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/mailer"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/services/sockd"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type Option struct {
	Database datahub.Database
	Logger   *slog.Logger
	Signer   *signer.Signer
	AppOpts  *xtypes.AppOptions
	Mailer   mailer.Mailer

	WorkingFolderBase string
}

var _ xtypes.App = (*App)(nil)

type App struct {
	db      datahub.Database
	signer  *signer.Signer
	logger  *slog.Logger
	ctrl    *actions.Controller
	AppOpts *xtypes.AppOptions
	engine  *engine.Engine
	sockd   *sockd.Sockd

	server *server.Server
}

func New(opt Option) *App {

	engine := engine.NewEngine(engine.EngineOption{
		DB:            opt.Database,
		WorkingFolder: path.Join(opt.WorkingFolderBase, "engine"),
		Logger:        opt.Logger,
		Repos:         opt.AppOpts.Repos,
		HttpPort:      opt.AppOpts.Port,
	})

	sockd := sockd.NewSockd()

	happ := &App{
		db:     opt.Database,
		signer: opt.Signer,
		logger: opt.Logger,
		ctrl: actions.New(actions.Option{
			Database: opt.Database,
			Logger:   opt.Logger,
			Signer:   opt.Signer,
			AppOpts:  opt.AppOpts,
			Engine:   engine,
			Mailer:   opt.Mailer,
		}),
		engine:  engine,
		sockd:   sockd,
		AppOpts: opt.AppOpts,
	}

	hosts := make([]string, len(happ.AppOpts.Hosts))
	for i, host := range happ.AppOpts.Hosts {
		hosts[i] = host.Name
	}

	server := server.NewServer(server.Option{
		Port:        opt.AppOpts.Port,
		Ctrl:        happ.ctrl,
		Signer:      opt.Signer,
		Engine:      engine,
		Hosts:       hosts,
		LocalSocket: opt.AppOpts.SocketFile,
		SiteName:    opt.AppOpts.Name,
	})

	happ.server = server

	return happ
}

func (h *App) Init() error {

	h.logger.Info("Initializing HeadLess application")

	return nil
}

func (h *App) Start() error {

	err := h.engine.Start(h)
	if err != nil {
		return err
	}

	h.logger.Info("HeadLess application started")

	has, err := h.ctrl.HasFingerprint()
	if err != nil {
		return err
	}

	qq.Println(h.AppOpts)

	// sha256 hash of the master secret
	shash := hashMasterSecret(h.AppOpts.MasterSecret)

	if !has {
		fingerPrint := &actions.AppFingerPrint{
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

	return h.server.Start()

}

// shared methods for App

func (h *App) Database() datahub.Database {
	return h.db
}

func (h *App) Signer() *signer.Signer {
	return h.signer
}

func (h *App) Logger() *slog.Logger {
	return h.logger
}

func (h *App) Controller() any {
	return h.ctrl
}

func (h *App) Engine() any {
	return h.engine
}

func (h *App) Config() any {
	return h.AppOpts
}

func (h *App) Sockd() any {
	return h.sockd
}

// private

func hashMasterSecret(masterSecret string) string {
	h := sha256.New()
	h.Write([]byte("SALT_FINGERPRINT"))
	h.Write([]byte(masterSecret))

	return string(h.Sum(nil))
}
