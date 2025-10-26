package main

import (
	"log"

	"github.com/blue-monads/turnix/backend"
	"github.com/blue-monads/turnix/backend/xtypes"
)

func main() {

	app, err := backend.NewDevApp(&xtypes.AppOptions{
		WorkingDir:   "./tmp",
		Port:         7777,
		Hosts:        []string{"*.localhost"},
		Name:         "PotatoVerse",
		MasterSecret: "default-master-secret",
		Debug:        true,
		SocketFile:   "",
		Mailer: xtypes.MailerOptions{
			Type: "stdio",
		},
	}, true)

	if err != nil {
		log.Fatalf("Failed to create HeadLess app: %v", err)
	}

	err = app.Start()
	if err != nil {
		log.Fatalf("Failed to start HeadLess app: %v", err)
	}

	ch := make(chan struct{})
	<-ch // block forever

}
