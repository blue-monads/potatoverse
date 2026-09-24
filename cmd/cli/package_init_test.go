package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsParentGitRepo(t *testing.T) {
	// Inside current git repository
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	targetInside := filepath.Join(cwd, "test-nonexistent-project")
	if !isParentGitRepo(targetInside) {
		t.Errorf("expected isParentGitRepo(%q) to be true inside repo", targetInside)
	}

	// Inside /tmp (not in git repository)
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	targetOutside := filepath.Join(tmpDir, "test-nonexistent-project")
	if isParentGitRepo(targetOutside) {
		t.Errorf("expected isParentGitRepo(%q) to be false outside repo", targetOutside)
	}
}

func TestResetGitState_InsideGitRepo(t *testing.T) {
	// Create temporary directory inside current git repo
	dir, err := os.MkdirTemp(".", "test-resetgit-inside-*")
	if err != nil {
		t.Fatalf("failed to create local temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Simulate template .git folder
	fakeGit := filepath.Join(dir, ".git")
	if err := os.Mkdir(fakeGit, 0755); err != nil {
		t.Fatalf("failed to create fake .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fakeGit, "config"), []byte("template"), 0644); err != nil {
		t.Fatalf("failed to write fake git config: %v", err)
	}

	if err := resetGitState(dir); err != nil {
		t.Fatalf("resetGitState failed: %v", err)
	}

	// Since parent is a git repo, .git should be removed and git init should NOT be run
	if _, err := os.Stat(fakeGit); !os.IsNotExist(err) {
		t.Errorf("expected .git to NOT exist in %s because parent is a git repo", dir)
	}
}

func TestResetGitState_OutsideGitRepo(t *testing.T) {
	// Create temporary directory in /tmp (outside any git repo)
	tmpDir, err := os.MkdirTemp("", "test-resetgit-outside-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	projDir := filepath.Join(tmpDir, "my-app")
	if err := os.Mkdir(projDir, 0755); err != nil {
		t.Fatalf("failed to create projDir: %v", err)
	}

	// Simulate template .git folder
	fakeGit := filepath.Join(projDir, ".git")
	if err := os.Mkdir(fakeGit, 0755); err != nil {
		t.Fatalf("failed to create fake .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fakeGit, "config"), []byte("template"), 0644); err != nil {
		t.Fatalf("failed to write fake git config: %v", err)
	}

	if err := resetGitState(projDir); err != nil {
		t.Fatalf("resetGitState failed: %v", err)
	}

	// Outside git repo, git init SHOULD have run, creating a new .git
	if _, err := os.Stat(fakeGit); err != nil {
		t.Errorf("expected .git to exist in %s because parent is not a git repo, err: %v", projDir, err)
	}
}

func TestReplaceTemplateSlug(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-slug-replace-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldSlug := "my-old-slug"
	newSlug := "my-new-slug"

	// Create directory containing old slug
	subDir := filepath.Join(tmpDir, oldSlug+"-sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subDir: %v", err)
	}

	// Create file containing old slug in name and content
	filePath := filepath.Join(subDir, oldSlug+".txt")
	content := "Hello " + oldSlug + " world"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	if err := replaceTemplateSlug(tmpDir, oldSlug, newSlug); err != nil {
		t.Fatalf("replaceTemplateSlug failed: %v", err)
	}

	expectedFilePath := filepath.Join(tmpDir, newSlug+"-sub", newSlug+".txt")
	data, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatalf("failed to read renamed file: %v", err)
	}

	if !strings.Contains(string(data), "Hello "+newSlug+" world") {
		t.Errorf("expected file content to contain replaced slug, got: %s", string(data))
	}
}
