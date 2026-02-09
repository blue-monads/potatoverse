package main

import (
	"log"

	"github.com/blue-monads/potatoverse/backend"
	"github.com/blue-monads/potatoverse/backend/xtypes"

	_ "github.com/blue-monads/potatoverse/backend/distro"
)

func main() {

	app, err := backend.NewDevApp(&xtypes.AppOptions{
		WorkingDir: "./tmp",
		Port:       7777,
		Hosts: []xtypes.Host{
			{
				Name: "*.localhost",
			},
		},
		Name:         "PotatoVerse",
		MasterSecret: "default-master-secret",
		Debug:        true,
		SocketFile:   "",
		Mailer: xtypes.MailerOptions{
			Type: "stdio",
		},
		BuddyOptions: &xtypes.BuddyHubOptions{
			StaticBuddies: []*xtypes.BuddyInfo{
				{
					Pubkey: "npub1qs3d4qh0w68lg80yyl3ucqsmraufaztt2kyw0zrw042f90cl9h8s38azs5",
				},
			},
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
