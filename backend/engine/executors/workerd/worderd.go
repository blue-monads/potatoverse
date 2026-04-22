package workerd

import (
	"fmt"
	"net"
	"net/http/httputil"
	"os"
	"os/exec"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/remotehub"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/jaevor/go-nanoid"
)

type workerdExecutor struct {
	cmd              *exec.Cmd
	port             int
	proxy            *httputil.ReverseProxy
	execDir          string
	remoteHub        *remotehub.RemoteHub
	packageId        int64
	packageVersionId int64
	spaceId          int64
}

func (e *workerdExecutor) Cleanup() {
	qq.Println("@workerd cleanup", "port", e.port)
	if e.cmd != nil && e.cmd.Process != nil {
		e.cmd.Process.Kill()
		e.cmd.Wait()
	}
	if e.execDir != "" {
		os.RemoveAll(e.execDir)
	}
}

func (e *workerdExecutor) GetDebugData() map[string]any {
	pid := 0
	if e.cmd != nil && e.cmd.Process != nil {
		pid = e.cmd.Process.Pid
	}
	data := map[string]any{
		"executor": "workerd",
		"port":     e.port,
		"pid":      pid,
		"execDir":  e.execDir,
	}
	return data
}

var (
	idgen func() string
)

func init() {
	idgen, _ = nanoid.ASCII(10)
}

func (e *workerdExecutor) HandleHttp(event *xtypes.HttpEvent) error {
	reqId := idgen()

	headers := event.Request.Request.Header

	headers.Set("X-Potato-Request-ID", reqId)
	token, err := e.remoteHub.GetExecToken(e.packageId, e.packageVersionId, e.spaceId, reqId)
	if err != nil {
		qq.Println("error getting exec token:", err)
		return err
	}

	if token != "" {
		headers.Set("X-Exec-Header", token)
	}

	e.proxy.ServeHTTP(event.Request.Writer, event.Request.Request)
	return nil
}

func (e *workerdExecutor) HandleAction(event *xtypes.ActionEvent) error {
	return fmt.Errorf("HandleAction not implemented for workerd")
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
