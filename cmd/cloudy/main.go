package main

import (
	"os"
	"strconv"

	"github.com/blue-monads/potatoverse/backend/distro/cloudy"
	turso_libs "github.com/tursodatabase/turso-go-platform-libs"
	turso "turso.tech/database/tursogo"
)

func main() {

	turso.InitLibrary(turso_libs.LoadTursoLibraryConfig{LoadStrategy: "mixed"})

	masterSecret := os.Getenv("CLOUDY_MASTER_SECRET")
	if masterSecret == "" {
		panic("CLOUDY_MASTER_SECRET env variable is required")
	}

	tursoAuthToken := os.Getenv("TURSO_AUTH_TOKEN")
	if tursoAuthToken == "" {
		panic("TURSO_AUTH_TOKEN env variable is required")
	}

	tursoRemoteURL := os.Getenv("TURSO_REMOTE_URL")
	if tursoRemoteURL == "" {
		panic("TURSO_REMOTE_URL env variable is required")
	}

	serveDomain := os.Getenv("CLOUDY_DOMAIN")
	if serveDomain == "" {
		panic("CLOUDY_DOMAIN env variable is required")
	}

	port := os.Getenv("CLOUDY_PORT")
	if port == "" {
		port = "8080"
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		panic("CLOUDY_PORT env variable must be an integer")
	}

	smtpPort := 587
	if p := os.Getenv("SMTP_PORT"); p != "" {
		smtpPort, err = strconv.Atoi(p)
		if err != nil {
			panic("SMTP_PORT env variable must be an integer")
		}
	}

	capp, err := cloudy.New(&cloudy.Config{
		Port:           portInt,
		WorkingDir:     "./cloudy_data",
		MasterSecret:   masterSecret,
		TursoAuthToken: tursoAuthToken,
		TursoRemoteURL: tursoRemoteURL,
		TursoAPIToken:  os.Getenv("TURSO_API_TOKEN"),
		TursoOrg:       os.Getenv("TURSO_ORG"),
		TursoGroup:     os.Getenv("TURSO_GROUP"),
		TursoDBPrefix:  os.Getenv("TURSO_DB_PREFIX"),
		Domain:         serveDomain,
		SMTP: cloudy.SMTPConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     smtpPort,
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
			FromName: os.Getenv("SMTP_FROM_NAME"),
		},
	})
	if err != nil {
		panic(err)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			if err := capp.Migrate(); err != nil {
				panic(err)
			}
			return
		default:
			panic("unknown command: " + os.Args[1])
		}
	}

	if err := capp.Run(); err != nil {
		panic(err)
	}

}
