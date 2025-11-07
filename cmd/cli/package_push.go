package cli

import (
	"errors"
	"net/http"
	"os"

	"github.com/alecthomas/kong"
)

func (c *PackagePushCmd) Run(_ *kong.Context) error {
	potatoToml, err := readPotatoToml(c.PotatoTomlFile)
	if err != nil {
		return err
	}

	zip, err := PackageFiles(c.PotatoTomlFile, c.OutputZipFile)
	if err != nil {
		return err
	}

	token := potatoToml.Developer.Token
	if token == "" {
		return errors.New("token is required")
	}

	serverUrl := potatoToml.Developer.ServerUrl
	if serverUrl == "" {
		return errors.New("server url is required")
	}

	file, err := os.Open(zip)
	if err != nil {
		return err
	}
	defer file.Close()

	req, err := http.NewRequest("POST", serverUrl, file)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/zip")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to push package")
	}

	return nil
}
