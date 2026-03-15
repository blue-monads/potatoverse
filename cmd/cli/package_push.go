package cli

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/alecthomas/kong"
	"github.com/blue-monads/potatoverse/backend/xtypes/models"
	"github.com/blue-monads/potatoverse/cmd/cli/pkgutils"
)

func (c *PackagePushCmd) Run(_ *kong.Context) error {

	potatoYaml, err := pkgutils.ReadPotatoFile(c.PotatoYamlFile)
	if err != nil {
		return err
	}

	outputZipFile := c.OutputZipFile

	if outputZipFile == "" {
		outputZipFile = potatoYaml.Developer.OutputZipFile
		if outputZipFile == "" {
			outputZipFile = fmt.Sprintf("%s.zip", potatoYaml.Slug)
		}
	}

	return PushPackage(c.PotatoYamlFile, outputZipFile)
}

func PushPackage(potatoYamlFile string, outputZipFile string) error {
	potatoYaml, err := pkgutils.ReadPotatoFile(potatoYamlFile)
	if err != nil {
		return err
	}

	file, err := os.Open(outputZipFile)
	if err != nil {
		return err
	}
	defer file.Close()

	serverUrl := potatoYaml.Developer.ServerUrl
	if serverUrl == "" {
		return errors.New("server url is required")
	}

	url := fmt.Sprintf("%s/zz/api/core/package/push", serverUrl)
	req, err := http.NewRequest("POST", url, file)
	if err != nil {
		return err
	}

	token, err := deriveDevToken(potatoYaml)
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
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to push package: %s %s", resp.Status, string(body))

	}

	return nil
}

func deriveDevToken(potatoYaml *models.PotatoPackage) (string, error) {

	token := potatoYaml.Developer.Token
	if token == "" {
		if potatoYaml.Developer.TokenEnv == "" {
			return "", errors.New("token is required/1")
		}

		token = os.Getenv(potatoYaml.Developer.TokenEnv)
	}

	if token == "" {
		return "", errors.New("token is required/2")
	}

	return "", nil
}
