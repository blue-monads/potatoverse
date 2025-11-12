package main

import (
	"crypto/tls"
	"time"

	buddyfs "github.com/blue-monads/turnix/backend/labs/buddyFs"
	"github.com/blue-monads/turnix/backend/utils/qq"
	flib "github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
)

func BetterFtpServer() {
	time.Sleep(2 * time.Second)

	qq.Println("BetterFtpServer")

	bd := buddyfs.NewBuddyFsClient("http://localhost:8666/buddyfs")
	fs := flib.NewFtpServer(&BuddyFsDriver{buddyClient: bd})
	err := fs.ListenAndServe()
	if err != nil {
		qq.Println("Failed to listen and serve: %v", err)
		return
	}

	qq.Println("BetterFtpServer started")

}

type BuddyFsDriver struct {
	buddyClient *buddyfs.BuddyFsClient
}

func (d *BuddyFsDriver) GetSettings() (*flib.Settings, error) {
	return &flib.Settings{
		Listener:    nil,
		ListenAddr:  "127.0.0.1:2121",
		PublicHost:  "127.0.0.1",
		Banner:      "BuddyFS FTP Server",
		TLSRequired: flib.MandatoryEncryption,
	}, nil
}

func (d *BuddyFsDriver) ClientConnected(cc flib.ClientContext) (string, error) {
	return "Welcome to BuddyFS FTP Server", nil
}

func (d *BuddyFsDriver) ClientDisconnected(cc flib.ClientContext) {

}

func (d *BuddyFsDriver) AuthUser(cc flib.ClientContext, user, pass string) (flib.ClientDriver, error) {

	fs := afero.NewBasePathFs(afero.NewOsFs(), "./tmp/buddyfs")

	return fs, nil
}

func (d *BuddyFsDriver) GetTLSConfig() (*tls.Config, error) {

	tc, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{tc},
		InsecureSkipVerify: true,
	}, nil
}
