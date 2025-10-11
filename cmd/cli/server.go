package cli

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/blue-monads/turnix/backend"
	xutils "github.com/blue-monads/turnix/backend/utils"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/pelletier/go-toml/v2"
)

func (c *ServerInitCmd) Run(ctx *kong.Context) error {

	config := xtypes.AppOptions{}

	config.Name = c.Name
	config.Port = c.Port
	config.Host = c.Host
	config.SocketFile = c.SocketFile
	config.MasterSecret = c.MasterSecret
	config.Debug = c.Debug
	config.WorkingDir = c.WorkingDir
	config.Mailer = xtypes.MailerOptions{
		Type: "stdio",
	}

	if config.MasterSecret == "" && c.MasterSecretEnv != "" {
		config.MasterSecret = fmt.Sprintf("$%s", c.MasterSecretEnv)
	}

	if config.MasterSecret == "" {
		random, err := xutils.GenerateRandomString(32)
		if err != nil {
			return err
		}

		config.MasterSecret = fmt.Sprintf("potatosec_%s", random)
	}

	if config.WorkingDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		config.WorkingDir = path.Join(cwd, ".pdata")
	}

	if config.SocketFile == "" {
		config.SocketFile = path.Join(config.WorkingDir, "potatoverse.sock")
	}

	cfgData, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	os.MkdirAll(config.WorkingDir, 0755)

	err = os.WriteFile("./config.toml", cfgData, 0644)
	if err != nil {
		return err
	}

	return nil

}

func (c *ServerStartCmd) Run(ctx *kong.Context) error {

	cfgData, err := os.ReadFile(c.Config)
	if err != nil {
		return err
	}

	config := xtypes.AppOptions{}
	err = toml.Unmarshal(cfgData, &config)
	if err != nil {
		return err
	}

	if after, ok := strings.CutPrefix(config.MasterSecret, "$"); ok {
		config.MasterSecret = os.Getenv(after)
	}

	app, err := backend.NewProdApp(&config, c.AutoSeed)
	if err != nil {
		return err
	}

	err = app.Start()
	if err != nil {
		return err
	}

	ch := make(chan struct{})
	<-ch // block forever

	return nil
}

type ServerStopCmd struct {
	Force bool `name:"force" short:"f" help:"Force stop the server."`
}

func (c *ServerStopCmd) Run(ctx *kong.Context) error {
	panic("Server Stop - Not implemented yet")

}
