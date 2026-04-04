package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

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

	fmt.Println("Package pushed sucessfully!")

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

	if strings.HasPrefix(token, "ppsec_") {
		return token, nil
	}
	if !strings.HasPrefix(token, "pdsec_") {
		return token, nil
	}

	// pdsec_ is a device token; exchange it for an ephemeral ppsec_ package dev token
	baseURL := strings.TrimSuffix(potatoYaml.Developer.ServerUrl, "/")
	accessToken, err := exchangeDeviceTokenForAccess(baseURL, token)
	if err != nil {
		return "", err
	}

	packageId := potatoYaml.Developer.PackageId
	if packageId == 0 {
		packageId, err = resolvePackageIdBySlug(baseURL, accessToken, potatoYaml.Slug)
		if err != nil {
			return "", err
		}
	} else {
		fmt.Println("Using package id from potato.yaml:", packageId)
	}

	ppsecToken, err := fetchPackageDevToken(baseURL, accessToken, packageId)
	if err != nil {
		return "", err
	}
	return ppsecToken, nil
}

const coreAPI = "/zz/api/core"

func exchangeDeviceTokenForAccess(baseURL, deviceToken string) (string, error) {
	fmt.Println("Exchanging token...")

	body, _ := json.Marshal(map[string]string{"device_token": deviceToken})
	req, err := http.NewRequest("POST", baseURL+coreAPI+"/auth/device-token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("device-token exchange failed: %s %s", resp.Status, string(b))
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("failed to decode device-token exchange response: %w", err)
	}
	if out.AccessToken == "" {
		return "", errors.New("device-token response missing access_token")
	}

	fmt.Println("Token exchange successful")

	return out.AccessToken, nil
}

func resolvePackageIdBySlug(baseURL, accessToken, slug string) (int64, error) {
	fmt.Println("Empty package id, trying to resolve using slug")

	req, err := http.NewRequest("GET", baseURL+coreAPI+"/space/installed", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("list installed failed: %s %s", resp.Status, string(b))
	}
	var out struct {
		Packages []struct {
			InstallId int64  `json:"install_id"`
			Slug      string `json:"slug"`
		} `json:"packages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, err
	}
	var matches []int64
	for _, p := range out.Packages {
		if p.Slug == slug {
			matches = append(matches, p.InstallId)
		}
	}
	switch len(matches) {
	case 0:
		return 0, fmt.Errorf("no installed package found for slug %q; set developer.package_id in potato.yaml or install the package first", slug)
	case 1:
		return matches[0], nil
	default:
		return chooseInstallId(slug, matches)
	}
}

func chooseInstallId(slug string, installIds []int64) (int64, error) {
	fmt.Fprintf(os.Stderr, "Multiple installed packages match slug %q:\n", slug)
	for i, id := range installIds {
		fmt.Fprintf(os.Stderr, "  %d) install_id %d\n", i+1, id)
	}
	fmt.Fprintf(os.Stderr, "Choose 1-%d (or set developer.package_id in potato.yaml): ", len(installIds))
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return 0, errors.New("no input when choosing package")
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	text := strings.TrimSpace(scanner.Text())
	n, err := strconv.Atoi(text)
	if err != nil || n < 1 || n > len(installIds) {
		return 0, fmt.Errorf("invalid choice %q; enter 1-%d", text, len(installIds))
	}
	return installIds[n-1], nil
}

func fetchPackageDevToken(baseURL, accessToken string, packageId int64) (string, error) {
	fmt.Println("Fetching development token...")
	url := fmt.Sprintf("%s%s/package/%d/dev-token?epthermal=true", baseURL, coreAPI, packageId)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "TokenV1 "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("dev-token request failed: %s %s", resp.Status, string(b))
	}
	var out struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("could not decode response: %w", err)
	}
	if out.Token == "" {
		return "", errors.New("dev-token response missing token")
	}
	fmt.Println("Development token fetched successfully")
	return out.Token, nil
}
