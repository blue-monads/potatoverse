package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/blue-monads/potatoverse/backend/utils/luaplus"
	"github.com/blue-monads/potatoverse/cmd/cli/pkgutils"
	"github.com/cjoudrey/gluahttp"
	lua "github.com/yuin/gopher-lua"
	luaJson "layeh.com/gopher-json"
)

type DevTestsCmd struct {
	LuaFile        string `arg:"" name:"lua-file" help:"Path to Lua test file." type:"path"`
	WorkingDir     string `name:"working-dir" help:"Working directory." default:"./.pdata"`
	PotatoYamlFile string `name:"potato-yaml-file" help:"Path to potato manifest file." type:"path" default:"./potato.yaml"`
	UseSpace       string `name:"use-space" help:"Target space namespace key (defaults to root/default space from potato.yaml)." default:""`
}

type DevShellCmd struct {
	WorkingDir     string   `name:"working-dir" help:"Working directory." default:"./.pdata"`
	PotatoYamlFile string   `name:"potato-yaml-file" help:"Path to potato manifest file." type:"path" default:"./potato.yaml"`
	UseSpace       string   `name:"use-space" help:"Target space namespace key (defaults to root/default space from potato.yaml)." default:""`
	Command        []string `arg:"" optional:"" help:"Shell command to execute."`
}

type DevRunAndTestsCmd struct {
	LuaFile         string        `arg:"" name:"lua-file" help:"Path to Lua test file." type:"path"`
	Port            int           `name:"port" short:"p" help:"Server port." default:"7777"`
	Host            string        `name:"host" help:"Server host." default:"*.localhost"`
	WorkingDir      string        `name:"working-dir" help:"Working directory." default:"./.pdata"`
	PotatoYamlFile  string        `name:"potato-yaml-file" help:"Path to potato manifest file." type:"path" default:"./potato.yaml"`
	UseSpace        string        `name:"use-space" help:"Target space namespace key (defaults to root/default space from potato.yaml)." default:""`
	ResetState      bool          `name:"reset-state" help:"Wipe .pdata and start fresh." default:"false"`
	LiveUIServe     bool          `name:"live-ui-serve" help:"Proxy UI from local vite/dev server via POTATO_DEV_SPACES." default:"false"`
	LiveUIServePort int           `name:"live-ui-serve-port" help:"UI dev server port."`
	Delay           time.Duration `name:"delay" help:"Delay after server starts before running tests." default:"10s"`
	KeepServer      bool          `name:"keep-server" help:"Keep server running after tests finish." default:"false"`
}

type DevRunAndShellCmd struct {
	Port            int           `name:"port" short:"p" help:"Server port." default:"7777"`
	Host            string        `name:"host" help:"Server host." default:"*.localhost"`
	WorkingDir      string        `name:"working-dir" help:"Working directory." default:"./.pdata"`
	PotatoYamlFile  string        `name:"potato-yaml-file" help:"Path to potato manifest file." type:"path" default:"./potato.yaml"`
	UseSpace        string        `name:"use-space" help:"Target space namespace key (defaults to root/default space from potato.yaml)." default:""`
	ResetState      bool          `name:"reset-state" help:"Wipe .pdata and start fresh." default:"false"`
	LiveUIServe     bool          `name:"live-ui-serve" help:"Proxy UI from local vite/dev server via POTATO_DEV_SPACES." default:"false"`
	LiveUIServePort int           `name:"live-ui-serve-port" help:"UI dev server port."`
	Delay           time.Duration `name:"delay" help:"Delay after server starts before running shell." default:"10s"`
	KeepServer      bool          `name:"keep-server" help:"Keep server running after command finishes." default:"false"`
	Command         []string      `arg:"" optional:"" help:"Shell command to execute."`
}

type DevEnv struct {
	WorkingDir   string
	SockPath     string
	BaseURL      string
	Port         int
	AdminToken   string
	SpaceToken   string
	NamespaceKey string
	SpaceId      int64
	InstallId    int64
}

type ResolvedSpace struct {
	SpaceId      int64
	InstallId    int64
	NamespaceKey string
	Token        string
}

func (c *DevShellCmd) Run(_ *kong.Context) error {
	devEnv, err := resolveDevEnv(c.WorkingDir, c.PotatoYamlFile, c.UseSpace)
	if err != nil {
		return err
	}
	return runShellWithEnv(devEnv, c.Command)
}

func (c *DevRunAndShellCmd) Run(_ *kong.Context) error {
	cmd, err := startDevEnvironment(c.WorkingDir, c.PotatoYamlFile, c.Port, c.Host, c.ResetState, c.LiveUIServe, c.LiveUIServePort)
	if err != nil {
		return err
	}
	if cmd != nil {
		defer func() {
			if !c.KeepServer && cmd.Process != nil {
				_ = cmd.Process.Signal(os.Interrupt)
				_ = cmd.Process.Kill()
			}
		}()
	}

	delay := c.Delay
	if delay > 0 {
		fmt.Printf("Dev server ready, waiting %v before running shell...\n", delay)
		time.Sleep(delay)
	}

	devEnv, err := resolveDevEnv(c.WorkingDir, c.PotatoYamlFile, c.UseSpace)
	if err != nil {
		return err
	}

	return runShellWithEnv(devEnv, c.Command)
}

func (c *DevRunAndTestsCmd) Run(kctx *kong.Context) error {
	cmd, err := startDevEnvironment(c.WorkingDir, c.PotatoYamlFile, c.Port, c.Host, c.ResetState, c.LiveUIServe, c.LiveUIServePort)
	if err != nil {
		return err
	}
	if cmd != nil {
		defer func() {
			if !c.KeepServer && cmd.Process != nil {
				_ = cmd.Process.Signal(os.Interrupt)
				_ = cmd.Process.Kill()
			}
		}()
	}

	delay := c.Delay
	if delay > 0 {
		fmt.Printf("Dev server ready, waiting %v before running tests...\n", delay)
		time.Sleep(delay)
	}

	testCmd := &DevTestsCmd{
		LuaFile:        c.LuaFile,
		WorkingDir:     c.WorkingDir,
		PotatoYamlFile: c.PotatoYamlFile,
		UseSpace:       c.UseSpace,
	}
	return testCmd.Run(kctx)
}

func runShellWithEnv(devEnv *DevEnv, command []string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	var execCmd *exec.Cmd
	if len(command) == 0 {
		execCmd = exec.Command(shell)
	} else {
		execCmd = exec.Command(shell, "-c", strings.Join(command, " "))
	}

	env := os.Environ()
	env = setOrReplaceEnv(env, "POTATO_SERVER_URL", devEnv.BaseURL)
	env = setOrReplaceEnv(env, "POTATO_DEV_ADMIN_USER_TOKEN", devEnv.AdminToken)
	env = setOrReplaceEnv(env, "POTATO_DEV_SPACE_TOKEN", devEnv.SpaceToken)
	env = setOrReplaceEnv(env, "POTATO_DEV_NAMESPACE_KEY", devEnv.NamespaceKey)
	env = setOrReplaceEnv(env, "POTATO_DEV_SPACE_ID", fmt.Sprintf("%d", devEnv.SpaceId))

	execCmd.Env = env
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

func startDevEnvironment(workingDirParam, potatoYamlFile string, portOverride int, host string, resetState bool, liveUIServe bool, liveUIServePort int) (*exec.Cmd, error) {
	if potatoYamlFile == "" {
		potatoYamlFile = "./potato.yaml"
	}
	workingDir, err := filepath.Abs(workingDirParam)
	if err != nil {
		return nil, err
	}
	sockPath := filepath.Join(workingDir, "potatoverse.sock")

	running := unixRPCAlive(sockPath)
	if resetState {
		if running {
			return nil, errors.New("cannot reset state while the development server is running")
		}
		if err := os.RemoveAll(workingDir); err != nil {
			return nil, fmt.Errorf("failed to reset state: %w", err)
		}
		fmt.Println("Reset development state")
	}

	if running {
		fmt.Println("Development server is already running, pushing current app...")
		if err := pushCurrentApp(workingDir, potatoYamlFile); err != nil {
			return nil, err
		}
		return nil, nil
	}

	potatoYaml, err := pkgutils.ReadPotatoFile(potatoYamlFile)
	if err != nil {
		return nil, err
	}

	port, err := pickListenPort(portOverride)
	if err != nil {
		return nil, err
	}

	devSpaces := ""
	if liveUIServe {
		devSpaces, err = buildDevSpacesEnv(potatoYaml, liveUIServePort)
		if err != nil {
			return nil, err
		}
	}
	if err := ensureDevConfig(workingDir, sockPath, port, host); err != nil {
		return nil, err
	}

	cmd, err := startDevServer(workingDir, devSpaces)
	if err != nil {
		return nil, err
	}

	if err := waitForUnixRPC(sockPath, 30*time.Second); err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}

	if err := pushCurrentApp(workingDir, potatoYamlFile); err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}

	fmt.Printf("Development server running at http://localhost:%d/zz/pages\n", port)
	return cmd, nil
}

func (c *DevTestsCmd) Run(_ *kong.Context) error {
	luaFile, err := filepath.Abs(c.LuaFile)
	if err != nil {
		return err
	}
	if _, err := os.Stat(luaFile); err != nil {
		return fmt.Errorf("test file not found: %w", err)
	}

	devEnv, err := resolveDevEnv(c.WorkingDir, c.PotatoYamlFile, c.UseSpace)
	if err != nil {
		return err
	}

	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()

	// Preload extra modules
	L.PreloadModule("gluahttp", gluahttp.NewHttpModule(http.DefaultClient).Loader)
	L.PreloadModule("phttp", gluahttp.NewHttpModule(http.DefaultClient).Loader)
	L.PreloadModule("http", gluahttp.NewHttpModule(http.DefaultClient).Loader)
	L.PreloadModule("json", luaJson.Loader)

	ctxTable := buildTestContextTable(L, devEnv)

	// Preload potato-test module
	L.PreloadModule("potato-test", func(ls *lua.LState) int {
		mod := ls.NewTable()
		mod.RawSetString("ctx", ctxTable)
		mod.RawSetString("root_space", ctxTable.RawGetString("root_space"))
		mod.RawSetString("create_space_http", ctxTable.RawGetString("create_space_http"))
		mod.RawSetString("get_space_token", ctxTable.RawGetString("get_space_token"))
		mod.RawSetString("token", lua.LString(devEnv.SpaceToken))
		mod.RawSetString("admin_token", lua.LString(devEnv.AdminToken))
		mod.RawSetString("server_url", lua.LString(devEnv.BaseURL))
		mod.RawSetString("namespace_key", lua.LString(devEnv.NamespaceKey))
		ls.Push(mod)
		return 1
	})

	// Also make ctx available globally
	L.SetGlobal("ctx", ctxTable)

	if err := L.DoFile(luaFile); err != nil {
		return fmt.Errorf("test execution failed in %s: %w", filepath.Base(luaFile), err)
	}

	onTestRun := L.GetGlobal("on_test_run")
	if onTestRun.Type() == lua.LTFunction {
		err := L.CallByParam(lua.P{
			Fn:      onTestRun,
			NRet:    1,
			Protect: true,
		}, ctxTable)
		if err != nil {
			return fmt.Errorf("test failed in %s (on_test_run): %w", filepath.Base(luaFile), err)
		}
		if L.GetTop() > 0 {
			ret := L.Get(-1)
			if ret == lua.LFalse {
				return fmt.Errorf("test failed in %s: on_test_run returned false", filepath.Base(luaFile))
			}
		}
		fmt.Printf("✓ %s: on_test_run passed\n", filepath.Base(luaFile))
	} else {
		fmt.Printf("✓ %s executed successfully\n", filepath.Base(luaFile))
	}

	return nil
}

func resolveDevEnv(workingDirParam, potatoYamlParam, useSpace string) (*DevEnv, error) {
	workingDir, err := filepath.Abs(workingDirParam)
	if err != nil {
		return nil, err
	}
	sockPath := filepath.Join(workingDir, "potatoverse.sock")
	if !unixRPCAlive(sockPath) {
		return nil, fmt.Errorf("development server is not running (socket %s not found or not responding)", sockPath)
	}

	info, err := unixRPCCall(sockPath, "/info")
	if err != nil {
		return nil, fmt.Errorf("unix rpc /info failed: %w", err)
	}
	port, ok := jsonInt(info["port"])
	if !ok || port == 0 {
		return nil, errors.New("unix rpc /info did not return a valid port")
	}
	baseURL := fmt.Sprintf("http://localhost:%d", port)

	tokenData, err := unixRPCCall(sockPath, "/get_admin_token")
	if err != nil {
		return nil, fmt.Errorf("unix rpc /get_admin_token failed: %w", err)
	}
	adminToken, _ := tokenData["token"].(string)
	if adminToken == "" {
		return nil, errors.New("unix rpc /get_admin_token did not return a token")
	}

	targetNamespace := useSpace
	if targetNamespace == "" && potatoYamlParam != "" {
		if _, err := os.Stat(potatoYamlParam); err == nil {
			if pkg, err := pkgutils.ReadPotatoFile(potatoYamlParam); err == nil {
				for _, s := range pkg.Spaces {
					if s.IsDefault && s.Namespace != "" {
						targetNamespace = s.Namespace
						break
					}
				}
				if targetNamespace == "" && len(pkg.Spaces) > 0 {
					targetNamespace = pkg.Spaces[0].Namespace
				}
				if targetNamespace == "" && pkg.Slug != "" {
					targetNamespace = pkg.Slug
				}
			}
		}
	}

	resolved, err := resolveSpaceToken(sockPath, baseURL, adminToken, targetNamespace, 0)
	if err != nil {
		return nil, err
	}

	return &DevEnv{
		WorkingDir:   workingDir,
		SockPath:     sockPath,
		BaseURL:      baseURL,
		Port:         port,
		AdminToken:   adminToken,
		SpaceToken:   resolved.Token,
		NamespaceKey: resolved.NamespaceKey,
		SpaceId:      resolved.SpaceId,
		InstallId:    resolved.InstallId,
	}, nil
}

func resolveSpaceToken(sockPath, baseURL, adminToken, targetNamespace string, targetSpaceId int64) (*ResolvedSpace, error) {
	// First attempt via UNIX RPC
	args := map[string]any{}
	if targetNamespace != "" {
		args["namespace_key"] = targetNamespace
	}
	if targetSpaceId > 0 {
		args["space_id"] = targetSpaceId
	}
	data, err := unixRPCCallWithArgs(sockPath, "/get_space_token", args)
	if err == nil && data != nil {
		tok, _ := data["token"].(string)
		if tok != "" {
			sid, _ := jsonInt(data["space_id"])
			iid, _ := jsonInt(data["install_id"])
			ns, _ := data["namespace_key"].(string)
			return &ResolvedSpace{
				SpaceId:      int64(sid),
				InstallId:    int64(iid),
				NamespaceKey: ns,
				Token:        tok,
			}, nil
		}
	}

	// Fallback via HTTP API using admin token
	req, err := http.NewRequest("GET", baseURL+coreAPI+"/space/installed", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "TokenV1 "+adminToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list installed spaces failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list installed spaces returned status %s: %s", resp.Status, string(b))
	}

	var out struct {
		Spaces []struct {
			ID           int64  `json:"id"`
			InstallID    int64  `json:"install_id"`
			NamespaceKey string `json:"namespace_key"`
			SpaceType    string `json:"space_type"`
		} `json:"spaces"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode installed spaces: %w", err)
	}

	var matchedSpace *struct {
		ID           int64  `json:"id"`
		InstallID    int64  `json:"install_id"`
		NamespaceKey string `json:"namespace_key"`
		SpaceType    string `json:"space_type"`
	}

	for i := range out.Spaces {
		sp := &out.Spaces[i]
		if targetSpaceId > 0 && sp.ID == targetSpaceId {
			matchedSpace = sp
			break
		}
		if targetNamespace != "" && sp.NamespaceKey == targetNamespace {
			matchedSpace = sp
			break
		}
	}

	if matchedSpace == nil && targetSpaceId == 0 && targetNamespace == "" {
		for i := range out.Spaces {
			sp := &out.Spaces[i]
			if sp.SpaceType != "AppPlugin" {
				matchedSpace = sp
				break
			}
		}
		if matchedSpace == nil && len(out.Spaces) > 0 {
			matchedSpace = &out.Spaces[0]
		}
	}

	if matchedSpace == nil {
		if targetNamespace != "" {
			return nil, fmt.Errorf("space with namespace %q not found", targetNamespace)
		}
		return nil, errors.New("no installed spaces found")
	}

	authReqBody, _ := json.Marshal(map[string]any{"space_id": matchedSpace.ID})
	authReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s/space/authorize/%s", baseURL, coreAPI, matchedSpace.NamespaceKey), bytes.NewReader(authReqBody))
	if err != nil {
		return nil, err
	}
	authReq.Header.Set("Authorization", "TokenV1 "+adminToken)
	authReq.Header.Set("Content-Type", "application/json")

	authResp, err := http.DefaultClient.Do(authReq)
	if err != nil {
		return nil, fmt.Errorf("authorize space failed: %w", err)
	}
	defer authResp.Body.Close()

	if authResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(authResp.Body)
		return nil, fmt.Errorf("authorize space failed with status %s: %s", authResp.Status, string(b))
	}

	var authOut struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(authResp.Body).Decode(&authOut); err != nil {
		return nil, fmt.Errorf("failed to decode authorize space response: %w", err)
	}
	if authOut.Token == "" {
		return nil, errors.New("authorize space did not return a token")
	}

	return &ResolvedSpace{
		SpaceId:      matchedSpace.ID,
		InstallId:    matchedSpace.InstallID,
		NamespaceKey: matchedSpace.NamespaceKey,
		Token:        authOut.Token,
	}, nil
}

func buildTestContextTable(L *lua.LState, devEnv *DevEnv) *lua.LTable {
	ctxTable := L.NewTable()

	ctxTable.RawSetString("token", lua.LString(devEnv.SpaceToken))
	ctxTable.RawSetString("space_token", lua.LString(devEnv.SpaceToken))
	ctxTable.RawSetString("root_space_token", lua.LString(devEnv.SpaceToken))
	ctxTable.RawSetString("admin_token", lua.LString(devEnv.AdminToken))
	ctxTable.RawSetString("server_url", lua.LString(devEnv.BaseURL))
	ctxTable.RawSetString("namespace_key", lua.LString(devEnv.NamespaceKey))
	ctxTable.RawSetString("space_id", lua.LNumber(devEnv.SpaceId))

	rootHttpClient := newSpaceHttpClientTable(L, devEnv, devEnv.SpaceToken, devEnv.NamespaceKey, devEnv.SpaceId)

	ctxTable.RawSetString("root_space", L.NewFunction(func(ls *lua.LState) int {
		ls.Push(rootHttpClient)
		return 1
	}))

	createSpaceHttpFn := L.NewFunction(func(ls *lua.LState) int {
		argOffset := 1
		if ls.GetTop() >= 2 && ls.Get(1) == ctxTable {
			argOffset = 2
		}
		var targetNamespace string
		var targetSpaceId int64

		if ls.GetTop() >= argOffset {
			argVal := ls.Get(argOffset)
			if argVal.Type() == lua.LTString {
				targetNamespace = argVal.String()
			} else if tbl, ok := argVal.(*lua.LTable); ok {
				if ns := tbl.RawGetString("namespace_key"); ns != lua.LNil {
					targetNamespace = ns.String()
				} else if ns := tbl.RawGetString("namespace"); ns != lua.LNil {
					targetNamespace = ns.String()
				}
				if sid := tbl.RawGetString("space_id"); sid != lua.LNil {
					if n, ok := sid.(lua.LNumber); ok {
						targetSpaceId = int64(n)
					}
				}
			}
		}

		resolved, err := resolveSpaceToken(devEnv.SockPath, devEnv.BaseURL, devEnv.AdminToken, targetNamespace, targetSpaceId)
		if err != nil {
			ls.Push(lua.LNil)
			ls.Push(lua.LString(err.Error()))
			return 2
		}

		altClient := newSpaceHttpClientTable(ls, devEnv, resolved.Token, resolved.NamespaceKey, resolved.SpaceId)
		ls.Push(altClient)
		ls.Push(lua.LNil)
		return 2
	})
	ctxTable.RawSetString("create_space_http", createSpaceHttpFn)

	getSpaceTokenFn := L.NewFunction(func(ls *lua.LState) int {
		argOffset := 1
		if ls.GetTop() >= 2 && ls.Get(1) == ctxTable {
			argOffset = 2
		}
		var targetNamespace string
		var targetSpaceId int64

		if ls.GetTop() >= argOffset {
			argVal := ls.Get(argOffset)
			if argVal.Type() == lua.LTString {
				targetNamespace = argVal.String()
			} else if tbl, ok := argVal.(*lua.LTable); ok {
				if ns := tbl.RawGetString("namespace_key"); ns != lua.LNil {
					targetNamespace = ns.String()
				} else if ns := tbl.RawGetString("namespace"); ns != lua.LNil {
					targetNamespace = ns.String()
				}
				if sid := tbl.RawGetString("space_id"); sid != lua.LNil {
					if n, ok := sid.(lua.LNumber); ok {
						targetSpaceId = int64(n)
					}
				}
			}
		}

		resolved, err := resolveSpaceToken(devEnv.SockPath, devEnv.BaseURL, devEnv.AdminToken, targetNamespace, targetSpaceId)
		if err != nil {
			ls.Push(lua.LNil)
			ls.Push(lua.LString(err.Error()))
			return 2
		}

		ls.Push(lua.LString(resolved.Token))
		ls.Push(lua.LNil)
		return 2
	})
	ctxTable.RawSetString("get_space_token", getSpaceTokenFn)

	return ctxTable
}

func newSpaceHttpClientTable(L *lua.LState, devEnv *DevEnv, spaceToken, namespaceKey string, spaceId int64) *lua.LTable {
	clientTable := L.NewTable()

	clientTable.RawSetString("token", lua.LString(spaceToken))
	clientTable.RawSetString("namespace_key", lua.LString(namespaceKey))
	clientTable.RawSetString("namespace", lua.LString(namespaceKey))
	clientTable.RawSetString("space_id", lua.LNumber(spaceId))
	clientTable.RawSetString("server_url", lua.LString(devEnv.BaseURL))
	clientTable.RawSetString("base_url", lua.LString(devEnv.BaseURL))

	getTokenFn := L.NewFunction(func(ls *lua.LState) int {
		ls.Push(lua.LString(spaceToken))
		return 1
	})
	clientTable.RawSetString("get_token", getTokenFn)

	methods := []string{"get", "post", "put", "delete", "patch", "head", "options"}
	for _, m := range methods {
		httpMethod := strings.ToUpper(m)
		fn := L.NewFunction(func(ls *lua.LState) int {
			return doSpaceHttpRequest(ls, devEnv, spaceToken, namespaceKey, httpMethod)
		})
		clientTable.RawSetString(m, fn)
		clientTable.RawSetString(httpMethod, fn)
	}

	reqFn := L.NewFunction(func(ls *lua.LState) int {
		return doSpaceHttpRequest(ls, devEnv, spaceToken, namespaceKey, "")
	})
	clientTable.RawSetString("request", reqFn)
	clientTable.RawSetString("REQUEST", reqFn)

	return clientTable
}

func doSpaceHttpRequest(L *lua.LState, devEnv *DevEnv, spaceToken, defaultNamespaceKey, defaultMethod string) int {
	argOffset := 1
	if L.GetTop() >= 1 {
		first := L.Get(1)
		if first.Type() == lua.LTTable || first.Type() == lua.LTUserData {
			argOffset = 2
		}
	}

	method := defaultMethod
	var urlStr string
	var optsTable *lua.LTable

	if defaultMethod == "" {
		if L.GetTop() < argOffset+1 {
			L.Push(lua.LNil)
			L.Push(lua.LString("request requires method and url"))
			return 2
		}
		method = strings.ToUpper(L.CheckString(argOffset))
		urlStr = L.CheckString(argOffset + 1)
		if L.GetTop() >= argOffset+2 {
			optsTable = L.OptTable(argOffset+2, nil)
		}
	} else {
		if L.GetTop() < argOffset {
			L.Push(lua.LNil)
			L.Push(lua.LString("url is required"))
			return 2
		}
		urlStr = L.CheckString(argOffset)
		if L.GetTop() >= argOffset+1 {
			optsTable = L.OptTable(argOffset+1, nil)
		}
	}

	var reqURL string
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		reqURL = urlStr
	} else if strings.HasPrefix(urlStr, "/") {
		reqURL = devEnv.BaseURL + urlStr
	} else {
		reqURL = fmt.Sprintf("%s/zz/space/%s/%s", devEnv.BaseURL, defaultNamespaceKey, urlStr)
	}

	headers := make(http.Header)
	var bodyReader io.Reader
	timeout := 30 * time.Second

	if optsTable != nil {
		if hVal := optsTable.RawGetString("headers"); hVal.Type() == lua.LTTable {
			hTbl := hVal.(*lua.LTable)
			hTbl.ForEach(func(k, v lua.LValue) {
				headers.Set(k.String(), v.String())
			})
		}

		if tVal := optsTable.RawGetString("timeout"); tVal != lua.LNil {
			if num, ok := tVal.(lua.LNumber); ok {
				timeout = time.Duration(float64(num) * float64(time.Second))
			} else if str, ok := tVal.(lua.LString); ok {
				if d, err := time.ParseDuration(string(str)); err == nil {
					timeout = d
				}
			}
		}

		if qVal := optsTable.RawGetString("query"); qVal != lua.LNil {
			if qTbl, ok := qVal.(*lua.LTable); ok {
				qVals := url.Values{}
				qTbl.ForEach(func(k, v lua.LValue) {
					qVals.Add(k.String(), v.String())
				})
				if strings.Contains(reqURL, "?") {
					reqURL += "&" + qVals.Encode()
				} else {
					reqURL += "?" + qVals.Encode()
				}
			} else if qStr, ok := qVal.(lua.LString); ok {
				if strings.Contains(reqURL, "?") {
					reqURL += "&" + string(qStr)
				} else {
					reqURL += "?" + string(qStr)
				}
			}
		}

		payloadVal := optsTable.RawGetString("payload")
		if payloadVal == lua.LNil {
			payloadVal = optsTable.RawGetString("json")
		}
		bodyVal := optsTable.RawGetString("body")
		formVal := optsTable.RawGetString("form")

		if payloadVal != lua.LNil {
			if pTbl, ok := payloadVal.(*lua.LTable); ok {
				anyVal := luaplus.LuaToAny(L, pTbl)
				b, err := json.Marshal(anyVal)
				if err != nil {
					L.Push(lua.LNil)
					L.Push(lua.LString("failed to marshal payload: " + err.Error()))
					return 2
				}
				bodyReader = bytes.NewReader(b)
				if headers.Get("Content-Type") == "" {
					headers.Set("Content-Type", "application/json")
				}
			} else if pStr, ok := payloadVal.(lua.LString); ok {
				bodyReader = strings.NewReader(string(pStr))
				if headers.Get("Content-Type") == "" {
					headers.Set("Content-Type", "application/json")
				}
			} else {
				anyVal := luaplus.LuaTypeToGoType(L, payloadVal)
				b, err := json.Marshal(anyVal)
				if err == nil {
					bodyReader = bytes.NewReader(b)
					if headers.Get("Content-Type") == "" {
						headers.Set("Content-Type", "application/json")
					}
				}
			}
		} else if bodyVal != lua.LNil {
			if bStr, ok := bodyVal.(lua.LString); ok {
				bodyReader = strings.NewReader(string(bStr))
			} else if bTbl, ok := bodyVal.(*lua.LTable); ok {
				anyVal := luaplus.LuaToAny(L, bTbl)
				b, _ := json.Marshal(anyVal)
				bodyReader = bytes.NewReader(b)
				if headers.Get("Content-Type") == "" {
					headers.Set("Content-Type", "application/json")
				}
			}
		} else if formVal != lua.LNil {
			if fTbl, ok := formVal.(*lua.LTable); ok {
				formVals := url.Values{}
				fTbl.ForEach(func(k, v lua.LValue) {
					formVals.Add(k.String(), v.String())
				})
				bodyReader = strings.NewReader(formVals.Encode())
				if headers.Get("Content-Type") == "" {
					headers.Set("Content-Type", "application/x-www-form-urlencoded")
				}
			}
		}
	}

	if headers.Get("Authorization") == "" && spaceToken != "" {
		headers.Set("Authorization", "Bearer "+spaceToken)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("failed to create http request: " + err.Error()))
		return 2
	}
	httpReq.Header = headers

	if optsTable != nil {
		if cVal := optsTable.RawGetString("cookies"); cVal.Type() == lua.LTTable {
			cTbl := cVal.(*lua.LTable)
			cTbl.ForEach(func(k, v lua.LValue) {
				httpReq.AddCookie(&http.Cookie{Name: k.String(), Value: v.String()})
			})
		}
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("http request failed: " + err.Error()))
		return 2
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("failed to read response body: " + err.Error()))
		return 2
	}

	resTable := L.NewTable()
	resTable.RawSetString("status_code", lua.LNumber(resp.StatusCode))
	resTable.RawSetString("status", lua.LString(resp.Status))
	resTable.RawSetString("body", lua.LString(string(respBytes)))
	resTable.RawSetString("ok", lua.LBool(resp.StatusCode >= 200 && resp.StatusCode < 300))

	headersTbl := L.NewTable()
	for k := range resp.Header {
		val := resp.Header.Get(k)
		headersTbl.RawSetString(k, lua.LString(val))
		headersTbl.RawSetString(strings.ToLower(k), lua.LString(val))
	}
	resTable.RawSetString("headers", headersTbl)

	cookiesTbl := L.NewTable()
	for _, c := range resp.Cookies() {
		cookiesTbl.RawSetString(c.Name, lua.LString(c.Value))
	}
	resTable.RawSetString("cookies", cookiesTbl)

	jsonFn := L.NewFunction(func(ls *lua.LState) int {
		var target any
		if err := json.Unmarshal(respBytes, &target); err != nil {
			ls.Push(lua.LNil)
			ls.Push(lua.LString("failed to decode json: " + err.Error()))
			return 2
		}
		ls.Push(luaplus.GoTypeToLuaType(ls, target))
		ls.Push(lua.LNil)
		return 2
	})
	resTable.RawSetString("json", jsonFn)

	var preParsed any
	if json.Unmarshal(respBytes, &preParsed) == nil {
		resTable.RawSetString("data", luaplus.GoTypeToLuaType(L, preParsed))
	}

	L.Push(resTable)
	L.Push(lua.LNil)
	return 2
}

func setOrReplaceEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}
