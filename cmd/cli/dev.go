package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub"
	xutils "github.com/blue-monads/potatoverse/backend/utils"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/models"
	"github.com/blue-monads/potatoverse/cmd/cli/pkgutils"
	"gopkg.in/yaml.v3"
)

type DevCmd struct {
	Run  DevRunCmd  `cmd:"" help:"Start a local development server for the current potato app."`
	Push DevPushCmd `cmd:"" help:"Push the current potato app to a running local development server."`
}

type DevRunCmd struct {
	Port            int    `name:"port" short:"p" help:"Server port." default:"7777"`
	Host            string `name:"host" help:"Server host." default:"*.localhost"`
	WorkingDir      string `name:"working-dir" help:"Working directory." default:"./.pdata"`
	ResetState      bool   `name:"reset-state" help:"Wipe .pdata and start fresh." default:"false"`
	LiveUIServe     bool   `name:"live-ui-serve" help:"Proxy UI from local vite/dev server via POTATO_DEV_SPACES." default:"false"`
	LiveUIServePort int    `name:"live-ui-serve-port" help:"UI dev server port. Overrides potato.yaml spaces.dev_serve_port."`
}

type DevPushCmd struct {
	WorkingDir     string `name:"working-dir" help:"Working directory." default:"./.pdata"`
	PotatoYamlFile string `name:"potato-yaml-file" help:"Path to potato manifest file." type:"path" default:"./potato.yaml"`
}

func (c *DevRunCmd) Run(_ *kong.Context) error {
	workingDir, err := filepath.Abs(c.WorkingDir)
	if err != nil {
		return err
	}
	sockPath := filepath.Join(workingDir, "potatoverse.sock")

	running := unixRPCAlive(sockPath)
	if c.ResetState {
		if running {
			return errors.New("cannot reset state while the development server is running")
		}
		if err := os.RemoveAll(workingDir); err != nil {
			return fmt.Errorf("failed to reset state: %w", err)
		}
		fmt.Println("Reset development state")
	}

	if running {
		return errors.New("development server is already running")
	}

	pdataExists := dirExists(workingDir)
	if pdataExists && !c.ResetState {
		fmt.Println("Existing .pdata found, starting with saved state")
	}

	potatoYaml, err := pkgutils.ReadPotatoFile("./potato.yaml")
	if err != nil {
		return err
	}

	port, err := pickListenPort(c.Port)
	if err != nil {
		return err
	}

	devSpaces := ""
	if c.LiveUIServe {
		devSpaces, err = buildDevSpacesEnv(potatoYaml, c.LiveUIServePort)
		if err != nil {
			return err
		}
	}
	if err := ensureDevConfig(workingDir, sockPath, port, c.Host); err != nil {
		return err
	}

	cmd, err := startDevServer(workingDir, devSpaces)
	if err != nil {
		return err
	}

	if err := waitForUnixRPC(sockPath, 30*time.Second); err != nil {
		_ = cmd.Process.Kill()
		return err
	}

	if err := pushCurrentApp(workingDir, "./potato.yaml"); err != nil {
		_ = cmd.Process.Kill()
		return err
	}

	fmt.Printf("Development server running at http://localhost:%d/zz/pages\n", port)
	return waitForProcess(cmd)
}

func (c *DevPushCmd) Run(_ *kong.Context) error {
	workingDir, err := filepath.Abs(c.WorkingDir)
	if err != nil {
		return err
	}
	return pushCurrentApp(workingDir, c.PotatoYamlFile)
}

func pushCurrentApp(workingDir, potatoYamlFile string) error {
	sockPath := filepath.Join(workingDir, "potatoverse.sock")
	if _, err := os.Stat(sockPath); err != nil {
		return fmt.Errorf("development server socket not found at %s", sockPath)
	}

	potatoYaml, err := pkgutils.ReadPotatoFile(potatoYamlFile)
	if err != nil {
		return err
	}

	zipFile, err := resolveOutputZip(potatoYaml)
	if err != nil {
		return err
	}
	if _, err := os.Stat(zipFile); err != nil {
		fmt.Println("Package zip not found, building...")
		if err := RunBuildCommand(potatoYamlFile); err != nil {
			return err
		}
		if _, err := PackageFiles(potatoYamlFile, zipFile); err != nil {
			return err
		}
	}

	info, err := unixRPCCall(sockPath, "/info")
	if err != nil {
		return err
	}
	port, ok := jsonInt(info["port"])
	if !ok || port == 0 {
		return errors.New("unix rpc /info did not return a port")
	}
	baseURL := fmt.Sprintf("http://localhost:%d", port)

	tokenData, err := unixRPCCall(sockPath, "/get_admin_token")
	if err != nil {
		return err
	}
	adminToken, _ := tokenData["token"].(string)
	if adminToken == "" {
		return errors.New("unix rpc /get_admin_token did not return a token")
	}

	packageId, err := resolveOrInstallPackage(baseURL, adminToken, potatoYaml.Slug, zipFile)
	if err != nil {
		return err
	}

	ppsec, err := fetchPackageDevToken(baseURL, adminToken, packageId)
	if err != nil {
		return err
	}

	return pushZip(baseURL, zipFile, ppsec)
}

func resolveOrInstallPackage(baseURL, adminToken, slug, zipFile string) (int64, error) {
	packageId, err := resolvePackageIdBySlug(baseURL, adminToken, slug)
	if err == nil {
		return packageId, nil
	}
	if !strings.Contains(err.Error(), "no installed package found") {
		return 0, err
	}

	fmt.Println("Package not installed yet, installing...")
	installedId, err := installPackageZip(baseURL, adminToken, zipFile)
	if err != nil {
		return 0, err
	}
	if installedId == 0 {
		installedId = 1
	}
	return installedId, nil
}

func installPackageZip(baseURL, accessToken, zipFile string) (int64, error) {
	file, err := os.Open(zipFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	req, err := http.NewRequest("POST", baseURL+coreAPI+"/package/install/zip", file)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "TokenV1 "+accessToken)
	req.Header.Set("Content-Type", "application/zip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("install zip failed: %s %s", resp.Status, string(b))
	}

	var out struct {
		InstalledId int64 `json:"installed_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, fmt.Errorf("failed to decode install response: %w", err)
	}
	fmt.Println("Package installed, id:", out.InstalledId)
	return out.InstalledId, nil
}

func pushZip(baseURL, zipFile, ppsecToken string) error {
	file, err := os.Open(zipFile)
	if err != nil {
		return err
	}
	defer file.Close()

	req, err := http.NewRequest("POST", baseURL+coreAPI+"/package/push", file)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", ppsecToken)
	req.Header.Set("Content-Type", "application/zip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to push package: %s %s", resp.Status, string(body))
	}

	fmt.Println("Package pushed sucessfully!")
	return nil
}

func resolveOutputZip(potatoYaml *models.PotatoPackage) (string, error) {
	if potatoYaml.Developer != nil && potatoYaml.Developer.OutputZipFile != "" {
		return potatoYaml.Developer.OutputZipFile, nil
	}
	if potatoYaml.Slug == "" {
		return "", errors.New("potato.yaml is missing slug")
	}
	return fmt.Sprintf("%s.spk.zip", potatoYaml.Slug), nil
}

func buildDevSpacesEnv(pkg *models.PotatoPackage, portOverride int) (string, error) {
	parts := make([]string, 0, len(pkg.Spaces))
	for _, space := range pkg.Spaces {
		if space.Namespace == "" {
			continue
		}
		port := space.DevServePort
		if portOverride > 0 {
			port = portOverride
		}
		if port <= 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s:%d", space.Namespace, port))
	}
	if len(parts) == 0 {
		return "", errors.New("live-ui-serve is set but no space port found; set spaces.dev_serve_port in potato.yaml or pass --live-ui-serve-port")
	}
	return strings.Join(parts, ","), nil
}

func pickListenPort(preferred int) (int, error) {
	if preferred > 0 && portFree(preferred) {
		return preferred, nil
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("failed to pick a free port: %w", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	if preferred > 0 {
		fmt.Printf("port %d is in use, using %d instead\n", preferred, port)
	}
	return port, nil
}

func portFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

func ensureDevConfig(workingDir, sockPath string, port int, host string) error {
	if err := os.MkdirAll(workingDir, 0755); err != nil {
		return err
	}

	cfgPath := filepath.Join(workingDir, "config.yaml")
	config := xtypes.AppOptions{}
	if cfgData, err := os.ReadFile(cfgPath); err == nil {
		if err := yaml.Unmarshal(cfgData, &config); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	if config.MasterSecret == "" {
		secret, err := xutils.GenerateRandomString(32)
		if err != nil {
			return err
		}
		config.MasterSecret = fmt.Sprintf("potatosec_%s", secret)
	}
	if config.Name == "" {
		config.Name = "PotatoVerse Dev"
	}
	if len(config.Hosts) == 0 {
		config.Hosts = []xtypes.Host{{Name: host}}
	}
	if len(config.Repos) == 0 {
		config.Repos = repohub.Default
	}

	config.Port = port
	config.Debug = true
	config.WorkingDir = workingDir
	config.SocketFile = sockPath
	config.Mailer = xtypes.MailerOptions{Type: "stdio"}

	cfgData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(cfgPath, cfgData, 0644)
}

func startDevServer(workingDir, devSpaces string) (*exec.Cmd, error) {
	binary, err := os.Executable()
	if err != nil {
		return nil, err
	}

	cfgPath := filepath.Join(workingDir, "config.yaml")
	cmd := exec.Command(binary, "server", "actual-start", "--config", cfgPath, "--auto-seed")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = filterEnv(os.Environ(), "POTATO_DEV_SPACES")
	if devSpaces != "" {
		cmd.Env = append(cmd.Env, "POTATO_DEV_SPACES="+devSpaces)
		fmt.Println("POTATO_DEV_SPACES=", devSpaces)
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func waitForProcess(cmd *exec.Cmd) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Wait()
	}()

	select {
	case sig := <-sigCh:
		_ = cmd.Process.Signal(sig)
		return <-errCh
	case err := <-errCh:
		return err
	}
}

type unixRPCResp struct {
	Ok   bool           `json:"ok"`
	Msg  string         `json:"msg"`
	Data map[string]any `json:"data"`
}

func unixRPCCall(sockPath, method string) (map[string]any, error) {
	conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("unix rpc connect: %w", err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := fmt.Fprintf(conn, "%s\n", method); err != nil {
		return nil, fmt.Errorf("unix rpc write: %w", err)
	}

	var resp unixRPCResp
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, fmt.Errorf("unix rpc decode: %w", err)
	}
	if !resp.Ok {
		msg := resp.Msg
		if msg == "" {
			msg = "unix rpc failed"
		}
		return nil, errors.New(msg)
	}
	return resp.Data, nil
}

func unixRPCAlive(sockPath string) bool {
	_, err := unixRPCCall(sockPath, "/info")
	return err == nil
}

func waitForUnixRPC(sockPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if unixRPCAlive(sockPath) {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return errors.New("timed out waiting for development server socket")
}

func filterEnv(env []string, key string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env))
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func jsonInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		return int(i), err == nil
	default:
		return 0, false
	}
}
