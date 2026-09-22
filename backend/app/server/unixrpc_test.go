package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/blue-monads/potatoverse/backend/app/actions"
	"github.com/blue-monads/potatoverse/backend/services/buddyhub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

func setupTestServer(t *testing.T) (*Server, string, func()) {
	tmpDir, err := os.MkdirTemp("", "potatoverse_unixrpc_test_*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	opts := &xtypes.AppOptions{
		Name:         "PotatoTest",
		Port:         9999,
		Hosts:        []xtypes.Host{{Name: "localhost"}},
		MasterSecret: "test_master_secret_1234567890",
		WorkingDir:   tmpDir,
		SocketFile:   filepath.Join(tmpDir, "potatoverse.sock"),
	}

	logger := slog.Default()
	bhub := buddyhub.NewBuddyHub(opts, logger)
	dbFile := filepath.Join(tmpDir, "test.sqlite")
	db, err := database.NewDB(dbFile, logger)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("new db: %v", err)
	}
	if err := db.Init(bhub); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("init db: %v", err)
	}

	sig := signer.New([]byte(opts.MasterSecret))
	ctrl := actions.New(actions.Option{
		Database: db,
		Logger:   logger,
		Signer:   sig,
		AppOpts:  opts,
	})

	if err := ctrl.AddUserGroup("admin", "Admin group"); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("add admin group: %v", err)
	}
	if _, err := ctrl.AddAdminUserDirect("demo", "demopass123", "demo@example.com"); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("add demo admin: %v", err)
	}
	if _, err := ctrl.AddAdminUserDirect("batman", "batmanpass123", "batman@example.com"); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("add batman admin: %v", err)
	}

	sockPath := filepath.Join(tmpDir, "test.sock")
	l, err := net.Listen("unix", sockPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("listen unix socket: %v", err)
	}

	srv := NewServer(Option{
		Port:   opts.Port,
		Ctrl:   ctrl,
		Signer: sig,
		Hosts:  []string{"localhost"},
	})

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go srv.handleUnixRPC(c)
		}
	}()

	cleanup := func() {
		l.Close()
		os.RemoveAll(tmpDir)
	}

	return srv, sockPath, cleanup
}

func callRPCRequest(sockPath string, req xtypes.UNIXRpcRequest) (UNIXRpcResponse, error) {
	conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
	if err != nil {
		return UNIXRpcResponse{}, err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return UNIXRpcResponse{}, err
	}

	var resp UNIXRpcResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return UNIXRpcResponse{}, err
	}
	return resp, nil
}

func callRPC(sockPath, request string) (UNIXRpcResponse, error) {
	conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
	if err != nil {
		return UNIXRpcResponse{}, err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := fmt.Fprintf(conn, "%s\n", request); err != nil {
		return UNIXRpcResponse{}, err
	}

	var resp UNIXRpcResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return UNIXRpcResponse{}, err
	}
	return resp, nil
}

func TestUnixRPC(t *testing.T) {
	_, sockPath, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("info command", func(t *testing.T) {
		resp, err := callRPC(sockPath, "/info")
		if err != nil {
			t.Fatalf("rpc call: %v", err)
		}
		if !resp.Ok {
			t.Fatalf("expected ok, got: %s", resp.Msg)
		}
		if resp.Data["port"].(float64) != 9999 {
			t.Fatalf("expected port 9999, got %v", resp.Data["port"])
		}
	})

	t.Run("get_admin_token command via UNIXRpcRequest", func(t *testing.T) {
		resp, err := callRPCRequest(sockPath, xtypes.UNIXRpcRequest{
			Method: "/get_admin_token",
		})
		if err != nil {
			t.Fatalf("rpc call: %v", err)
		}
		if !resp.Ok {
			t.Fatalf("expected ok, got: %s", resp.Msg)
		}
		token, ok := resp.Data["token"].(string)
		if !ok || token == "" {
			t.Fatalf("expected non-empty token")
		}
	})

	t.Run("list_admin_user command via UNIXRpcRequest", func(t *testing.T) {
		resp, err := callRPCRequest(sockPath, xtypes.UNIXRpcRequest{
			Method: "list_admin_user",
		})
		if err != nil {
			t.Fatalf("rpc call: %v", err)
		}
		if !resp.Ok {
			t.Fatalf("expected ok, got: %s", resp.Msg)
		}
		users, ok := resp.Data["users"].([]any)
		if !ok || len(users) == 0 {
			t.Fatalf("expected admin users list, got %v", resp.Data["users"])
		}
	})

	t.Run("list_all_user command via UNIXRpcRequest", func(t *testing.T) {
		resp, err := callRPCRequest(sockPath, xtypes.UNIXRpcRequest{
			Method: "/list_all_user",
		})
		if err != nil {
			t.Fatalf("rpc call: %v", err)
		}
		if !resp.Ok {
			t.Fatalf("expected ok, got: %s", resp.Msg)
		}
		users, ok := resp.Data["users"].([]any)
		if !ok || len(users) == 0 {
			t.Fatalf("expected all users list, got %v", resp.Data["users"])
		}
	})

	t.Run("reset_user_pass default admin", func(t *testing.T) {
		resp, err := callRPCRequest(sockPath, xtypes.UNIXRpcRequest{
			Method: "reset_user_pass",
		})
		if err != nil {
			t.Fatalf("rpc call: %v", err)
		}
		if !resp.Ok {
			t.Fatalf("expected ok, got: %s", resp.Msg)
		}
		pass, ok := resp.Data["password"].(string)
		if !ok || pass == "" {
			t.Fatalf("expected new password in data")
		}
	})

	t.Run("reset_user_pass with UNIXRpcRequest Args", func(t *testing.T) {
		resp, err := callRPCRequest(sockPath, xtypes.UNIXRpcRequest{
			Method: "reset_user_pass",
			Args: map[string]any{
				"username": "batman",
				"password": "struct_pass_123",
			},
		})
		if err != nil {
			t.Fatalf("rpc call: %v", err)
		}
		if !resp.Ok {
			t.Fatalf("expected ok, got: %s", resp.Msg)
		}
		if resp.Data["password"] != "struct_pass_123" {
			t.Fatalf("expected struct_pass_123, got %v", resp.Data["password"])
		}
	})

	t.Run("reset_user_pass with user id in Args", func(t *testing.T) {
		resp, err := callRPCRequest(sockPath, xtypes.UNIXRpcRequest{
			Method: "/reset_user_pass",
			Args: map[string]any{
				"user_id":  1,
				"password": "id_pass_456",
			},
		})
		if err != nil {
			t.Fatalf("rpc call: %v", err)
		}
		if !resp.Ok {
			t.Fatalf("expected ok, got: %s", resp.Msg)
		}
		if resp.Data["password"] != "id_pass_456" {
			t.Fatalf("expected id_pass_456, got %v", resp.Data["password"])
		}
	})
}
