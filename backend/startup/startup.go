package startup

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/blue-monads/potatoverse/backend/app"
	"github.com/blue-monads/potatoverse/backend/app/actions"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub"
	"github.com/blue-monads/potatoverse/backend/services/buddyhub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/mailer/stdio"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

func BuildApp(options *xtypes.AppOptions, seedDB bool) (*app.App, error) {

	logger := slog.Default()

	maindbDir := filepath.Join(options.WorkingDir, "datadb")
	dbFile := filepath.Join(maindbDir, "main.sqlite")

	os.MkdirAll(maindbDir, 0755)

	bhub := buddyhub.NewBuddyHub(options, logger)

	db, err := database.NewDB(dbFile, logger)
	if err != nil {
		logger.Error("Failed to initialize database", "err", err)
		return nil, err
	}

	if err := db.Init(bhub); err != nil {
		logger.Error("Failed to initialize database", "err", err)
		return nil, err
	}

	m := stdio.NewMailer(logger.With("module", "mailer"))

	if options.Name == "" {
		options.Name = "PotatoVerse"
	}

	// randNumber := rand.Intn(10000000)
	// randNumer2 := rand.Intn(10000000)

	// if randNumber == 11 && randNumer2 == 11 {
	// 	database.StartLitestream(dbFile)
	// }

	if len(options.Repos) == 0 {
		options.Repos = repohub.Default
	}

	happ := app.New(app.Option{
		Database: db,
		Logger:   logger,
		Signer:   signer.New([]byte(options.MasterSecret)),
		AppOpts: &xtypes.AppOptions{
			Port:         options.Port,
			Hosts:        options.Hosts,
			MasterSecret: options.MasterSecret,
			Debug:        options.Debug,
			WorkingDir:   options.WorkingDir,
			Name:         options.Name,
			SocketFile:   options.SocketFile,
			Mailer:       options.Mailer,
			Repos:        options.Repos,
		},
		Mailer:            m,
		WorkingFolderBase: options.WorkingDir,
		BuddyHub:          bhub,
	})

	if seedDB {
		ctrl := happ.Controller().(*actions.Controller)

		ugroups, err := ctrl.ListUserGroups()
		if err != nil {
			return nil, err
		}

		if len(ugroups) == 0 {
			err = ctrl.AddUserGroup("admin", "Admin group")
			if err != nil {
				return nil, err
			}

			err = ctrl.AddUserGroup("normal", "Normal group")
			if err != nil {
				return nil, err
			}

			_, err = ctrl.AddAdminUserDirect("demo", "demogodTheGreat_123", "demo@example.com")
			if err != nil {
				return nil, err
			}

			_, err = ctrl.AddAdminUserDirect("batman", "ilikebats_123", "batman@example.com")
			if err != nil {
				return nil, err
			}

			uops := db.GetUserOps()

			_, err = uops.AddUserMessage(&dbmodels.UserMessage{
				Title:         "Welcome to PotatoVerse",
				Contents:      "Welcome to PotatoVerse",
				ToUser:        1,
				IsRead:        false,
				FromUserId:    0,
				FromSpaceId:   0,
				CallbackToken: "",
				WarnLevel:     0,
			})
			if err != nil {
				return nil, err
			}

		}

	}

	return happ, nil
}
