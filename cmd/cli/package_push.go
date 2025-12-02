package cli

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/alecthomas/kong"
)

func (c *PackagePushCmd) Run(_ *kong.Context) error {

	zip, err := PackageFiles(c.PotatoTomlFile, c.OutputZipFile)
	if err != nil {
		return err
	}

	return PushPackage(c.PotatoTomlFile, zip)
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

	serverUrl := potatoToml.Developer.ServerUrl
	if serverUrl == "" {
		return errors.New("server url is required")
	}

	token := potatoToml.Developer.Token
	if token == "" {
		if potatoToml.Developer.TokenEnv == "" {
			return errors.New("token is required")
		}

		token = os.Getenv(potatoToml.Developer.TokenEnv)
		if token == "" {
			return errors.New("token is required")
		}
	}

	url := fmt.Sprintf("%s/zz/api/core/package/push", serverUrl)
	req, err := http.NewRequest("POST", url, file)
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
