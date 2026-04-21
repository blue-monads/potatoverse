package worderd

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
	"github.com/google/uuid"
)

const (
	WorkerdBinary     = "/home/bingo/go/bin/workerd"
	CompatibilityDate = "2023-05-18"
)

func init() {
	registry.RegisterExecutorBuilderFactory("worderd", BuildWorderdExecutorBuilder)
}

func BuildWorderdExecutorBuilder(app xtypes.App) (xtypes.ExecutorBuilder, error) {
	return &WorderdExecutorBuilder{app: app}, nil
}

type WorderdExecutorBuilder struct {
	app xtypes.App
}

func (b *WorderdExecutorBuilder) Name() string { return "worderd" }
func (b *WorderdExecutorBuilder) Icon() string { return "worderd" }

func (b *WorderdExecutorBuilder) Build(opt *xtypes.ExecutorBuilderOption) (xtypes.Executor, error) {
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
		workingDir = filepath.Join(os.TempDir(), "potatoverse-worderd")
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
      compatibilityDate = "%s",
    )),
    (name = "internal_bindings", external = (address = "127.0.0.1:%d")),
  ],
  sockets = [
    ( name = "http",
      address = "127.0.0.1:%d",
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
	for i := 0; i < 20; i++ {
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

	qq.Println("@worderd started", "port", port, "mainPort", mainPort, "space", opt.SpaceId)

	return &WorderdExecutor{
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

type WorderdExecutor struct {
	cmd              *exec.Cmd
	port             int
	proxy            *httputil.ReverseProxy
	execDir          string
	remoteHub        *remotehub.RemoteHub
	packageId        int64
	packageVersionId int64
	spaceId          int64
}

func (e *WorderdExecutor) Cleanup() {
	qq.Println("@worderd cleanup", "port", e.port)
	if e.cmd != nil && e.cmd.Process != nil {
		e.cmd.Process.Kill()
		e.cmd.Wait()
	}
	if e.execDir != "" {
		os.RemoveAll(e.execDir)
	}
}

func (e *WorderdExecutor) GetDebugData() map[string]any {
	pid := 0
	if e.cmd != nil && e.cmd.Process != nil {
		pid = e.cmd.Process.Pid
	}
	data := map[string]any{
		"executor": "worderd",
		"port":     e.port,
		"pid":      pid,
		"execDir":  e.execDir,
	}
	return data
}

func (e *WorderdExecutor) HandleHttp(event *xtypes.HttpEvent) error {
	reqId := uuid.New().String()
	token, err := e.remoteHub.GetExecToken(e.packageId, e.packageVersionId, e.spaceId, reqId)
	if err != nil {
		return err
	}

	req := event.Request.Request.Clone(event.Request.Request.Context())
	req.Header.Set("X-Potato-Request-ID", reqId)
	req.Header.Set("X-Exec-Header", token)

	e.proxy.ServeHTTP(event.Request.Writer, req)
	return nil
}

func (e *WorderdExecutor) HandleAction(event *xtypes.ActionEvent) error {
	return fmt.Errorf("HandleAction not implemented for worderd")
}

func findFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
