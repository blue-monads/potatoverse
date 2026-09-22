package cli

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/blue-monads/potatoverse/backend/xtypes"
	"gopkg.in/yaml.v3"
)

func TestDevMasterSecretDeterministic(t *testing.T) {
	// Ensure env is clean
	origEnv := os.Getenv("POTATO_DEV_MASTER_SECRET")
	defer os.Setenv("POTATO_DEV_MASTER_SECRET", origEnv)
	os.Unsetenv("POTATO_DEV_MASTER_SECRET")

	tempDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to getwd: %v", err)
	}
	defer os.Chdir(origWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Secret should be deterministic and match sha256 of tempDir
	sec1 := resolveDevMasterSecret()
	sec2 := resolveDevMasterSecret()
	if sec1 != sec2 {
		t.Fatalf("expected deterministic secret, got %s != %s", sec1, sec2)
	}

	absDir, _ := filepath.Abs(tempDir)
	expectedHash := fmt.Sprintf("potatosec_%x", sha256.Sum256([]byte(absDir)))
	if sec1 != expectedHash {
		t.Fatalf("expected %s, got %s", expectedHash, sec1)
	}

	// Test with .env.potato
	envPotatoPath := filepath.Join(tempDir, ".env.potato")
	if err := os.WriteFile(envPotatoPath, []byte("POTATO_DEV_MASTER_SECRET=potato-env-secret-123\n"), 0644); err != nil {
		t.Fatalf("failed to write .env.potato: %v", err)
	}

	secFromEnvFile := resolveDevMasterSecret()
	if secFromEnvFile != "potato-env-secret-123" {
		t.Fatalf("expected potato-env-secret-123, got %s", secFromEnvFile)
	}

	// Test with environment variable override
	os.Setenv("POTATO_DEV_MASTER_SECRET", "direct-env-override")
	secFromEnvVar := resolveDevMasterSecret()
	if secFromEnvVar != "direct-env-override" {
		t.Fatalf("expected direct-env-override, got %s", secFromEnvVar)
	}
}

func TestEnsureDevConfig(t *testing.T) {
	origEnv := os.Getenv("POTATO_DEV_MASTER_SECRET")
	defer os.Setenv("POTATO_DEV_MASTER_SECRET", origEnv)
	os.Unsetenv("POTATO_DEV_MASTER_SECRET")

	tempDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to getwd: %v", err)
	}
	defer os.Chdir(origWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	workingDir := filepath.Join(tempDir, ".pdata")
	sockPath := filepath.Join(workingDir, "potatoverse.sock")

	if err := ensureDevConfig(workingDir, sockPath, 8888, "*.localhost"); err != nil {
		t.Fatalf("ensureDevConfig failed: %v", err)
	}

	cfgData, err := os.ReadFile(filepath.Join(workingDir, "config.yaml"))
	if err != nil {
		t.Fatalf("failed to read config.yaml: %v", err)
	}

	var config xtypes.AppOptions
	if err := yaml.Unmarshal(cfgData, &config); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	expectedSecret := resolveDevMasterSecret()
	if config.MasterSecret != expectedSecret {
		t.Fatalf("expected MasterSecret %s, got %s", expectedSecret, config.MasterSecret)
	}
}
