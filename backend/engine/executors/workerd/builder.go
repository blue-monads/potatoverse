package workerd

import (
	"fmt"
	"net"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/blue-monads/potatoverse/backend/engine"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/remotehub"
	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

const (
	WorkerdBinary     = "/home/bingo/go/bin/workerd"
	CompatibilityDate = "2023-05-18"
)

func init() {
	registry.RegisterExecutorBuilderFactory("workerd", BuildworkerdExecutorBuilder)
}

func BuildworkerdExecutorBuilder(app xtypes.App) (xtypes.ExecutorBuilder, error) {
	return &workerdExecutorBuilder{app: app}, nil
}

type workerdExecutorBuilder struct {
	app xtypes.App
}

func (b *workerdExecutorBuilder) Name() string { return "workerd" }
func (b *workerdExecutorBuilder) Icon() string { return "workerd" }

func (b *workerdExecutorBuilder) Build(opt *xtypes.ExecutorBuilderOption) (xtypes.Executor, error) {
	code := ""
	if opt.CodeLoader != nil {
		var err error
		code, err = opt.CodeLoader()
		if err != nil {
			return nil, fmt.Errorf("could not load source code: %w", err)
		}
	} else {
		sOps := b.app.Database().GetSpaceOps()
		s, err := sOps.GetSpace(opt.SpaceId)
		if err == nil {
			if s.ServerFile == "" {
				s.ServerFile = "server.js"
			}

			pfops := b.app.Database().GetPackageFileOps()
			packageFile, err := pfops.GetFileContentByPath(opt.PackageVersionId, "", s.ServerFile)
			if err == nil {
				code = string(packageFile)
			}
		}
	}

	if code == "" {
		code = "// No code provided\nexport default { fetch() { return new Response('No code'); } };"
	}

	port, err := findFreePort()
	if err != nil {
		return nil, fmt.Errorf("could not find free port: %w", err)
	}

	workingDir := opt.WorkingFolder
	if workingDir == "" {
		workingDir = filepath.Join(os.TempDir(), "potatoverse-workerd")
	}

	execDir := filepath.Join(workingDir, fmt.Sprintf("exec-%d-%d", opt.SpaceId, time.Now().UnixNano()))
	if err := os.MkdirAll(execDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create working directory: %w", err)
	}

	workerPath := filepath.Join(execDir, "worker.js")
	if err := os.WriteFile(workerPath, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("could not write worker.js: %w", err)
	}

	// Use RemoteHub bindings
	eng := b.app.Engine().(*engine.Engine)
	rhub := eng.GetRemoteHub()
	mainPort := eng.HttpPort

	// Generate potato.js and config.capnp
	if err := os.WriteFile(filepath.Join(execDir, "potato.js"), []byte(remotehub.PotatoJs), 0644); err != nil {
		return nil, err
	}

	configContent := fmt.Sprintf(`
using Workerd = import "/workerd/workerd.capnp";

const config :Workerd.Config = (
  services = [
    (name = "main", worker = (
      modules = [
        (name = "worker", esModule = embed "worker.js"),
        (name = "potato", esModule = embed "potato.js")
      ],
      compatibilityDate = "%[1]s",
      bindings = [
        (name = "internal_bindings", service = "internal_bindings")
      ],
      globalOutbound = "internet"
    )),
    (name = "internal_bindings", worker = (
      modules = [
        (name = "forwarder.js", esModule = "export default { async fetch(req, env) { const url = new URL(req.url); url.port = env.PORT; return await env.net.fetch(url, req); } }")
      ],
      compatibilityDate = "%[1]s",
      bindings = [
        (name = "net", service = "internet"),
        (name = "PORT", text = "%[2]d")
      ]
    )),
    (name = "internet", network = (allow = ["public", "private", "local", "network", "127.0.0.0/8"])),
  ],
  sockets = [
    ( name = "http",
      address = "127.0.0.1:%[3]d",
      http = (),
      service = "main"
    ),
  ]
);
`, CompatibilityDate, mainPort, port)

	if err := os.WriteFile(filepath.Join(execDir, "config.capnp"), []byte(configContent), 0644); err != nil {
		return nil, err
	}

	cmd := exec.Command(WorkerdBinary, "serve", "config.capnp")
	cmd.Dir = execDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start workerd: %w", err)
	}

	started := false
	for range 20 {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 50*time.Millisecond)
		if err == nil {
			conn.Close()
			started = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if !started {
		cmd.Process.Kill()
		return nil, fmt.Errorf("workerd failed to start on port %d within timeout", port)
	}

	targetURL, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	qq.Println("@workerd started", "port", port, "mainPort", mainPort, "space", opt.SpaceId)

	return &workerdExecutor{
		cmd:              cmd,
		port:             port,
		proxy:            proxy,
		execDir:          execDir,
		remoteHub:        rhub,
		packageId:        opt.InstalledId,
		packageVersionId: opt.PackageVersionId,
		spaceId:          opt.SpaceId,
	}, nil
}
