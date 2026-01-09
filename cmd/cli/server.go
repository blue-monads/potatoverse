package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/blue-monads/potatoverse/backend"
	xutils "github.com/blue-monads/potatoverse/backend/utils"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/pelletier/go-toml/v2"
)

// server

type ServerCmd struct {
	Init  ServerInitCmd  `cmd:"" help:"Initialize the server with default options."`
	Start ServerStartCmd `cmd:"" help:"Start the server."`
	Stop  ServerStopCmd  `cmd:"" help:"Stop the server."`
}

type ServerInitCmd struct {
	DBFile          string `name:"db" help:"Path to database file." default:"data.db"`
	Port            int    `name:"port" short:"p" help:"Server port." default:"7777"`
	Host            string `name:"host" help:"Server host." default:"*.localhost"`
	Name            string `name:"name" help:"Name of node." default:"PotatoVerse"`
	SocketFile      string `name:"socket-file" help:"Socket file of node."`
	MasterSecret    string `name:"master-secret" help:"Master secret of node."`
	MasterSecretEnv string `name:"master-secret-env" help:"Master secret environment variable of node."`
	Debug           bool   `name:"debug" help:"Debug mode of node." default:"false"`
	WorkingDir      string `name:"working-dir" help:"Working dir of node."`
}

func (c *ServerInitCmd) Run(ctx *kong.Context) error {

	config := xtypes.AppOptions{}

	config.Name = c.Name
	config.Port = c.Port
	config.Hosts = []xtypes.Host{
		{
			Name: c.Host,
		},
	}
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

type ServerStartCmd struct {
	Config   string `name:"config" short:"c" help:"Path to configuration file." type:"path" default:"./config.toml"`
	AutoSeed bool   `name:"auto-seed" short:"s" help:"Auto seed the server." default:"false"`
}

func (c *ServerStartCmd) Run(ctx *kong.Context) error {

	binary, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(binary, "server", "actual-start", "--config", c.Config, "--auto-seed", fmt.Sprintf("%t", c.AutoSeed))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

type ServerActualStartCmd struct {
	Config   string `name:"config" short:"c" help:"Path to configuration file." type:"path" default:"./config.toml"`
	AutoSeed bool   `name:"auto-seed" short:"s" help:"Auto seed the server." default:"false"`
}

func (c *ServerActualStartCmd) Run(ctx *kong.Context) error {

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
