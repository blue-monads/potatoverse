package main

import (
	"context"
	"log"
	"os"

	"github.com/blue-monads/potatoverse/backend/utils/libx/dbutils"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	turso "turso.tech/database/tursogo"
)

func main() {

	tursoAuthToken := os.Getenv("TURSO_AUTH_TOKEN")
	if tursoAuthToken == "" {
		panic("TURSO_AUTH_TOKEN env variable is required")
	}

	tursoRemoteURL := os.Getenv("TURSO_REMOTE_URL")
	if tursoRemoteURL == "" {
		panic("TURSO_REMOTE_URL env variable is required")
	}

	ctx := context.Background()

	bootstrap := true

	tctx, err := turso.NewTursoSyncDb(ctx, turso.TursoSyncDbConfig{
		Path:             "aa.db",
		RemoteUrl:        tursoRemoteURL,
		AuthToken:        tursoAuthToken,
		BootstrapIfEmpty: &bootstrap,
	})

	if err != nil {
		panic(err)
	}

	pulled, err := tctx.Pull(ctx)
	if err != nil {
		panic(err)
	}

	if pulled {
		log.Println("Pulled latest changes from remote database")
	} else {
		log.Println("No changes to pull from remote database")
	}

	db, err := tctx.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("PRAGMA TABLE_INFO('test');")
	if err != nil {

		qq.Println("@", err)
		log.Fatal(err)
	}
	defer rows.Close()

	data, err := dbutils.GetScan(rows)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Data: %+v\n", data)

	// push changes to remote

	if err := tctx.Push(ctx); err != nil {
		log.Fatal(err)
	}

}
