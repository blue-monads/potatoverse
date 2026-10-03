package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blue-monads/potatoverse/backend/xtypes"
)

func setupMockDevServer(t *testing.T) (workingDir string, baseURL string, cleanup func()) {
	tmpDir := t.TempDir()

	// 1. Setup mock HTTP server
	httpSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if r.URL.Path == "/zz/space/test-space/echo" {
			b, _ := io.ReadAll(r.Body)
			var parsed any
			_ = json.Unmarshal(b, &parsed)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":            true,
				"echo":          parsed,
				"authorization": auth,
			})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/space/installed") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"packages": []map[string]any{
					{"install_id": 1, "slug": "test-app"},
				},
				"spaces": []map[string]any{
					{
						"id":            1,
						"install_id":    1,
						"namespace_key": "test-space",
						"space_type":    "App",
					},
					{
						"id":            2,
						"install_id":    1,
						"namespace_key": "other-space",
						"space_type":    "App",
					},
				},
			})
			return
		}
		if strings.Contains(r.URL.Path, "/dev-token") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"token": "ppsec_mock_token_123",
			})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/package/push") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok": true}`))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/core/space/authorize/") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"token": "mock-authorized-space-token",
			})
			return
		}
		http.NotFound(w, r)
	}))

	port := httpSrv.Listener.Addr().(*net.TCPAddr).Port

	// 2. Setup mock UNIX socket server
	sockPath := filepath.Join(tmpDir, "potatoverse.sock")
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		httpSrv.Close()
		t.Fatalf("listen unix sock: %v", err)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				var req xtypes.UNIXRpcRequest
				if err := json.NewDecoder(c).Decode(&req); err != nil {
					return
				}

				normMethod := strings.TrimPrefix(strings.ToLower(req.Method), "/")
				var resp xtypes.UNIXRpcResponse

				switch normMethod {
				case "info":
					resp = xtypes.UNIXRpcResponse{
						Ok: true,
						Data: map[string]any{
							"port": port,
							"host": []string{"localhost"},
						},
					}
				case "get_admin_token":
					resp = xtypes.UNIXRpcResponse{
						Ok: true,
						Data: map[string]any{
							"token": "admin-mock-token-xyz",
							"user":  map[string]any{"id": 1, "username": "admin"},
						},
					}
				case "get_space_token":
					ns := req.GetStringArg("namespace_key", "namespace")
					if ns == "" {
						ns = "test-space"
					}
					resp = xtypes.UNIXRpcResponse{
						Ok: true,
						Data: map[string]any{
							"token":         "mock-space-token-for-" + ns,
							"space_id":      1,
							"namespace_key": ns,
							"install_id":    1,
						},
					}
				default:
					resp = xtypes.UNIXRpcResponse{Ok: false, Msg: "unknown"}
				}

				_ = json.NewEncoder(c).Encode(resp)
			}(conn)
		}
	}()

	cleanup = func() {
		_ = ln.Close()
		httpSrv.Close()
	}

	return tmpDir, httpSrv.URL, cleanup
}

func TestDevTestsCmd(t *testing.T) {
	workingDir, baseURL, cleanup := setupMockDevServer(t)
	defer cleanup()

	// Create potato.yaml in temp dir
	potatoYamlPath := filepath.Join(workingDir, "potato.yaml")
	potatoYamlContent := `slug: test-app
spaces:
  - namespace: test-space
    is_default: true
`
	if err := os.WriteFile(potatoYamlPath, []byte(potatoYamlContent), 0644); err != nil {
		t.Fatalf("write potato.yaml: %v", err)
	}

	// Create lua test file
	luaFilePath := filepath.Join(workingDir, "api_test.lua")
	luaScript := `
local gluahttp = require("gluahttp")
assert(gluahttp ~= nil, "expected gluahttp module")

local ptest = require("potato-test")
assert(ptest ~= nil, "expected potato-test module")

function on_test_run(ctx)
    assert(ctx ~= nil, "expected ctx argument")
    assert(ctx.token == "mock-space-token-for-test-space", "token mismatch: " .. tostring(ctx.token))
    assert(ctx.admin_token == "admin-mock-token-xyz", "admin token mismatch")
    assert(ctx.namespace_key == "test-space", "namespace_key mismatch")

    local root = ctx.rootSpace()
    assert(root ~= nil, "expected rootSpace")
    assert(root.token == ctx.token, "rootSpace token mismatch")

    local res, err = root.post("/zz/space/test-space/echo", {
        payload = { message = "hello potatoverse", num = 123 }
    })

    assert(err == nil, "http post error: " .. tostring(err))
    assert(res ~= nil, "expected response")
    assert(res.status_code == 200, "expected status 200, got: " .. tostring(res.status_code))
    assert(res.ok == true, "expected res.ok == true")

    local data, jerr = res.json()
    assert(jerr == nil, "json decode error: " .. tostring(jerr))
    assert(data.echo.message == "hello potatoverse", "payload echo message mismatch")
    assert(data.echo.num == 123, "payload echo num mismatch")
    assert(data.authorization == "Bearer " .. ctx.token, "auth header mismatch: " .. tostring(data.authorization))

    -- Test createSpaceHttp
    local althttp, aerr = ctx.createSpaceHttp({
        namespace_key = "other-space"
    })
    assert(aerr == nil, "createSpaceHttp error: " .. tostring(aerr))
    assert(althttp ~= nil, "expected althttp client")
    assert(althttp.namespace_key == "other-space", "expected other-space namespace")
    assert(althttp.token == "mock-space-token-for-other-space", "expected other space token")
end
`
	if err := os.WriteFile(luaFilePath, []byte(luaScript), 0644); err != nil {
		t.Fatalf("write lua script: %v", err)
	}

	cmd := &DevTestsCmd{
		LuaFile:        luaFilePath,
		WorkingDir:     workingDir,
		PotatoYamlFile: potatoYamlPath,
	}

	if err := cmd.Run(nil); err != nil {
		t.Fatalf("DevTestsCmd.Run failed: %v", err)
	}

	_ = baseURL
}

func TestDevShellCmd(t *testing.T) {
	workingDir, _, cleanup := setupMockDevServer(t)
	defer cleanup()

	potatoYamlPath := filepath.Join(workingDir, "potato.yaml")
	potatoYamlContent := `slug: test-app
spaces:
  - namespace: test-space
    is_default: true
developer:
  output_zip_file: test-app.spk.zip
`
	if err := os.WriteFile(potatoYamlPath, []byte(potatoYamlContent), 0644); err != nil {
		t.Fatalf("write potato.yaml: %v", err)
	}

	outFilePath := filepath.Join(workingDir, "env_out.txt")

	cmd := &DevShellCmd{
		WorkingDir:     workingDir,
		PotatoYamlFile: potatoYamlPath,
		Command: []string{
			fmt.Sprintf("echo $POTATO_SERVER_URL $POTATO_DEV_ADMIN_USER_TOKEN $POTATO_DEV_SPACE_TOKEN > %s", outFilePath),
		},
	}

	if err := cmd.Run(nil); err != nil {
		t.Fatalf("DevShellCmd.Run failed: %v", err)
	}

	outBytes, err := os.ReadFile(outFilePath)
	if err != nil {
		t.Fatalf("read env_out.txt: %v", err)
	}

	outStr := strings.TrimSpace(string(outBytes))
	parts := strings.Split(outStr, " ")
	if len(parts) < 3 {
		t.Fatalf("expected at least 3 parts, got %q", outStr)
	}

	if !strings.HasPrefix(parts[0], "http://localhost:") && !strings.HasPrefix(parts[0], "http://127.0.0.1:") {
		t.Errorf("expected POTATO_SERVER_URL, got %s", parts[0])
	}
	if parts[1] != "admin-mock-token-xyz" {
		t.Errorf("expected POTATO_DEV_ADMIN_USER_TOKEN admin-mock-token-xyz, got %s", parts[1])
	}
	if parts[2] != "mock-space-token-for-test-space" {
		t.Errorf("expected POTATO_DEV_SPACE_TOKEN mock-space-token-for-test-space, got %s", parts[2])
	}
}

func TestDevRunAndShellCmd(t *testing.T) {
	workingDir, _, cleanup := setupMockDevServer(t)
	defer cleanup()

	potatoYamlPath := filepath.Join(workingDir, "potato.yaml")
	zipPath := filepath.Join(workingDir, "test-app.spk.zip")
	_ = os.WriteFile(zipPath, []byte("mock-zip-content"), 0644)

	potatoYamlContent := fmt.Sprintf(`slug: test-app
spaces:
  - namespace: test-space
    is_default: true
developer:
  output_zip_file: %s
`, zipPath)
	if err := os.WriteFile(potatoYamlPath, []byte(potatoYamlContent), 0644); err != nil {
		t.Fatalf("write potato.yaml: %v", err)
	}

	outFilePath := filepath.Join(workingDir, "env_out_run.txt")

	cmd := &DevRunAndShellCmd{
		WorkingDir:     workingDir,
		PotatoYamlFile: potatoYamlPath,
		Delay:          0,
		KeepServer:     true,
		Command: []string{
			fmt.Sprintf("echo $POTATO_SERVER_URL $POTATO_DEV_ADMIN_USER_TOKEN $POTATO_DEV_SPACE_TOKEN > %s", outFilePath),
		},
	}

	if err := cmd.Run(nil); err != nil {
		t.Fatalf("DevRunAndShellCmd.Run failed: %v", err)
	}

	outBytes, err := os.ReadFile(outFilePath)
	if err != nil {
		t.Fatalf("read env_out_run.txt: %v", err)
	}

	outStr := strings.TrimSpace(string(outBytes))
	parts := strings.Split(outStr, " ")
	if len(parts) < 3 {
		t.Fatalf("expected at least 3 parts, got %q", outStr)
	}
}

func TestDevRunAndTestsCmd(t *testing.T) {
	workingDir, _, cleanup := setupMockDevServer(t)
	defer cleanup()

	potatoYamlPath := filepath.Join(workingDir, "potato.yaml")
	zipPath := filepath.Join(workingDir, "test-app.spk.zip")
	_ = os.WriteFile(zipPath, []byte("mock-zip-content"), 0644)

	potatoYamlContent := fmt.Sprintf(`slug: test-app
spaces:
  - namespace: test-space
    is_default: true
developer:
  output_zip_file: %s
`, zipPath)
	if err := os.WriteFile(potatoYamlPath, []byte(potatoYamlContent), 0644); err != nil {
		t.Fatalf("write potato.yaml: %v", err)
	}

	luaFilePath := filepath.Join(workingDir, "run_test.lua")
	luaScript := `
function on_test_run(ctx)
    assert(ctx.token ~= nil and ctx.token ~= "", "expected token")
end
`
	if err := os.WriteFile(luaFilePath, []byte(luaScript), 0644); err != nil {
		t.Fatalf("write lua: %v", err)
	}

	cmd := &DevRunAndTestsCmd{
		LuaFile:        luaFilePath,
		WorkingDir:     workingDir,
		PotatoYamlFile: potatoYamlPath,
		Delay:          0,
		KeepServer:     true,
	}

	if err := cmd.Run(nil); err != nil {
		t.Fatalf("DevRunAndTestsCmd.Run failed: %v", err)
	}
}

func TestDevShellCmd_UseSpaceOverride(t *testing.T) {
	workingDir, _, cleanup := setupMockDevServer(t)
	defer cleanup()

	outFilePath := filepath.Join(workingDir, "env_out_custom.txt")

	cmd := &DevShellCmd{
		WorkingDir: workingDir,
		UseSpace:   "custom-space",
		Command: []string{
			fmt.Sprintf("echo $POTATO_DEV_SPACE_TOKEN > %s", outFilePath),
		},
	}

	if err := cmd.Run(nil); err != nil {
		t.Fatalf("DevShellCmd.Run failed: %v", err)
	}

	outBytes, err := os.ReadFile(outFilePath)
	if err != nil {
		t.Fatalf("read env_out_custom.txt: %v", err)
	}

	outStr := strings.TrimSpace(string(outBytes))
	if outStr != "mock-space-token-for-custom-space" {
		t.Fatalf("expected token mock-space-token-for-custom-space, got %q", outStr)
	}
}
