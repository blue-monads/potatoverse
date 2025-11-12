package main

import (
	"log"

	"goftp.io/server/v2"
	"goftp.io/server/v2/driver/file"
)

type SimpleAuth struct {
	Username string
	Password string
}

func (a *SimpleAuth) CheckPasswd(ctx *server.Context, name, pass string) (bool, error) {
	return name == a.Username && pass == a.Password, nil
}

func RunFtpServer() {

	driver, err := file.NewDriver("./tmp/buddyfs")
	if err != nil {
		log.Fatalf("Failed to create driver: %v", err)
	}

	auth := &SimpleAuth{
		Username: "admin",
		Password: "admin",
	}

	fserver, err := server.NewServer(&server.Options{
		Driver:       driver,
		Port:         2121,
		Hostname:     "localhost",
		Auth:         auth,
		Perm:         server.NewSimplePerm("admin", "admin"),
		RateLimit:    1024 * 1024,
		Commands:     server.DefaultCommands(),
		Name:         "BuddyFS FTP Server",
		PublicIP:     "127.0.0.1",
		PassivePorts: "1024-65535",
		TLS:          true,
		CertFile:     "cert.pem",
		KeyFile:      "key.pem",
		ExplicitFTPS: true,
	})
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
		return
	}

	err = fserver.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to listen and serve: %v", err)
		return
	}

}
