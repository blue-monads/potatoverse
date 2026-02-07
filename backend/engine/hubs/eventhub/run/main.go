package main

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/eslayer"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/evtype"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
)

func main() {

	tmpFolder := os.TempDir()
	qq.Println("tmpFolder", tmpFolder)

	defer os.RemoveAll(tmpFolder)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := database.NewDB(tmpFolder, logger)
	if err != nil {
		qq.Println("Failed to create database", "err", err)
		return
	}

	failCount := 0

	handlers := map[string]evtype.Handler{
		"test": func(ex *evtype.TExecution) error {

			if ex.Subscription.EventKey == "fail-2-times-then-success" {
				if failCount == 2 {

					return nil
				}

				ex.RetryAble = true
				return errors.New("Error occured ")
			}

			qq.Println("test", ex)
			return nil
		},
	}

	eslayer := eslayer.NewESLayer(db, handlers)

	err = eslayer.Start()
	if err != nil {
		qq.Println("Failed to start eslayer", "err", err)
		return
	}

	time.Sleep(10 * time.Second)

	db.GetMQSynk().AddEvent(1, "test-good", []byte("test"))
	db.GetMQSynk().AddEvent(1, "fail-2-times-then-success", []byte("test"))

	eslayer.Stop()

}
