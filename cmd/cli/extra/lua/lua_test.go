package luaextra

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckFile(t *testing.T) {
	tempDir := t.TempDir()

	validFile := filepath.Join(tempDir, "valid.lua")
	err := os.WriteFile(validFile, []byte(`
local x = 10
local y = 20
function add(a, b)
    return a + b
end
print(add(x, y))
`), 0644)
	if err != nil {
		t.Fatalf("failed to write valid lua file: %v", err)
	}

	invalidFile := filepath.Join(tempDir, "invalid.lua")
	err = os.WriteFile(invalidFile, []byte(`
function broken(a, b
    return a +
end
`), 0644)
	if err != nil {
		t.Fatalf("failed to write invalid lua file: %v", err)
	}

	t.Run("valid file passes check", func(t *testing.T) {
		err := CheckFile(validFile)
		if err != nil {
			t.Errorf("expected valid file to pass check, got error: %v", err)
		}
	})

	t.Run("invalid file fails check", func(t *testing.T) {
		err := CheckFile(invalidFile)
		if err == nil {
			t.Errorf("expected invalid file to fail check, got nil")
		}
	})

	t.Run("non-existent file fails check", func(t *testing.T) {
		err := CheckFile(filepath.Join(tempDir, "does-not-exist.lua"))
		if err == nil {
			t.Errorf("expected error for non-existent file, got nil")
		}
	})
}

func TestRunCheck(t *testing.T) {
	tempDir := t.TempDir()

	f1 := filepath.Join(tempDir, "f1.lua")
	_ = os.WriteFile(f1, []byte("local a = 1"), 0644)

	f2 := filepath.Join(tempDir, "f2.lua")
	_ = os.WriteFile(f2, []byte("local b = 2"), 0644)

	t.Run("no args returns error", func(t *testing.T) {
		err := RunCheck([]string{})
		if err == nil {
			t.Errorf("expected error when no args provided")
		}
	})

	t.Run("multiple valid files succeed", func(t *testing.T) {
		err := RunCheck([]string{f1, f2})
		if err != nil {
			t.Errorf("expected RunCheck to succeed, got: %v", err)
		}
	})

	t.Run("one invalid file fails check", func(t *testing.T) {
		bad := filepath.Join(tempDir, "bad.lua")
		_ = os.WriteFile(bad, []byte("function ("), 0644)

		err := RunCheck([]string{f1, bad})
		if err == nil {
			t.Errorf("expected RunCheck to fail when a file has syntax errors")
		}
	})
}

func TestShellState_Modules(t *testing.T) {
	L := NewShellState()
	defer L.Close()

	t.Run("json module can be required", func(t *testing.T) {
		code := `
local json = require("json")
local encoded = json.encode({name = "potato", score = 100})
assert(encoded ~= nil)
`
		err := L.DoString(code)
		if err != nil {
			t.Errorf("failed to require/use json module: %v", err)
		}
	})

	t.Run("phttp module can be required", func(t *testing.T) {
		code := `
local http = require("phttp")
assert(http ~= nil)
`
		err := L.DoString(code)
		if err != nil {
			t.Errorf("failed to require phttp module: %v", err)
		}
	})
}

func TestExecuteLine(t *testing.T) {
	L := NewShellState()
	defer L.Close()

	t.Run("evaluate expression", func(t *testing.T) {
		var output string
		executeLine(L, "10 + 25", func(p string) (string, error) {
			return "", nil
		}, func(out string) {
			output = out
		})

		if output != "35" {
			t.Errorf("expected output '35', got %q", output)
		}
	})

	t.Run("evaluate expression with equals prefix", func(t *testing.T) {
		var output string
		executeLine(L, "= 'hello' .. ' world'", func(p string) (string, error) {
			return "", nil
		}, func(out string) {
			output = out
		})

		if output != "hello world" {
			t.Errorf("expected output 'hello world', got %q", output)
		}
	})

	t.Run("evaluate multiline statement", func(t *testing.T) {
		var output string
		continuationLines := []string{
			"    local sum = 0",
			"    for i=1,n do sum = sum + i end",
			"    return sum",
			"end",
		}
		contIdx := 0

		executeLine(L, "function test_sum(n)", func(p string) (string, error) {
			if contIdx < len(continuationLines) {
				line := continuationLines[contIdx]
				contIdx++
				return line, nil
			}
			return "", nil
		}, func(out string) {
			output = out
		})

		// Now call test_sum(5)
		executeLine(L, "test_sum(5)", func(p string) (string, error) {
			return "", nil
		}, func(out string) {
			output = out
		})

		if output != "15" {
			t.Errorf("expected output '15', got %q", output)
		}
	})
}

func TestRunLuaDispatcher(t *testing.T) {
	t.Run("eval flag -e", func(t *testing.T) {
		err := RunLua([]string{"-e", "local z = 1 + 2"})
		if err != nil {
			t.Errorf("expected -e to succeed, got: %v", err)
		}
	})

	t.Run("eval flag without code returns error", func(t *testing.T) {
		err := RunLua([]string{"-e"})
		if err == nil {
			t.Errorf("expected error when -e has no code argument")
		}
	})

	t.Run("help subcommand", func(t *testing.T) {
		err := RunLua([]string{"help"})
		if err != nil {
			t.Errorf("expected help to succeed, got: %v", err)
		}
	})

	t.Run("file execution", func(t *testing.T) {
		tempDir := t.TempDir()
		script := filepath.Join(tempDir, "script.lua")
		_ = os.WriteFile(script, []byte("local x = 5\nassert(x == 5)"), 0644)

		err := RunLua([]string{script})
		if err != nil {
			t.Errorf("expected running script to succeed, got: %v", err)
		}
	})

	t.Run("unknown command returns error", func(t *testing.T) {
		err := RunLua([]string{"nonexistent-command-xyz"})
		if err == nil || !strings.Contains(err.Error(), "unknown lua command") {
			t.Errorf("expected unknown command error, got: %v", err)
		}
	})
}
