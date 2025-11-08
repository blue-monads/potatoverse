package cli

import (
	"fmt"
	"net/http"
	"os"

	"github.com/alecthomas/kong"
)

type PackageCmd struct {
	Build    PackageBuildCmd `cmd:"" help:"Build the package."`
	Push     PackagePushCmd  `cmd:"" help:"Push the package."`
	PushOnly PackagePushOnly `cmd:"" help:"Push the package only."`
}

type PackagePushCmd struct {
	PotatoTomlFile string `name:"potato-toml-file" help:"Path to package directory." type:"path" default:"./potato.toml"`
	OutputZipFile  string `name:"output-zip-file" help:"Output path for built package." type:"path"`
}

type PackageBuildCmd struct {
	PotatoTomlFile string `name:"potato-toml-file" help:"Path to package directory." type:"path" default:"./potato.toml"`
	OutputZipFile  string `name:"output-zip-file" help:"Output path for built package." type:"path"`
}

type PackagePushOnly struct {
	PotatoTomlFile string `name:"potato-toml-file" help:"Path to package directory." type:"path" default:"./potato.toml"`
	OutputZipFile  string `name:"output-zip-file" help:"Output path for built package." type:"path"`
}

func (c *PackagePushOnly) Run(ctx *kong.Context) error {

	return PushPackage(c.PotatoTomlFile, c.OutputZipFile)
}

func PushPackage(potatoTomlFile string, outputZipFile string) error {
	potatoToml, err := readPotatoToml(potatoTomlFile)
	if err != nil {
		return err
	}

	file, err := os.Open(outputZipFile)
	if err != nil {
		return err
	}
	defer file.Close()

	req, err := http.NewRequest("POST", potatoToml.Developer.ServerUrl, file)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", potatoToml.Developer.Token)
	req.Header.Set("Content-Type", "application/zip")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to push package: %s", resp.Status)
	}

	return nil
}
