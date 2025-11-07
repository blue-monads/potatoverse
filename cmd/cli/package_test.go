package cli

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/blue-monads/turnix/backend/xtypes"
)

func TestPackageFilesV2(t *testing.T) {
	testdataPath := filepath.Join("testdata")
	absTestdataPath, err := filepath.Abs(testdataPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	tests := []struct {
		name           string
		opts           *xtypes.PackagingOptions
		expectedFiles  []string
		expectedCount  int
		expectError    bool
		errorSubstring string
	}{
		{
			name: "include all files with **/*",
			opts: &xtypes.PackagingOptions{
				IncludeFiles: []string{"**/*"},
			},
			expectedFiles: []string{
				"aa/zz/xx.txt",
				"bb/abcisxyz.md",
				"cc/eeeeeeeee.gg",
				"cc/nnn/ok.ok",
				"ee/bullhorn.txt",
				"ee/sans.txt",
			},
			expectedCount: 6,
		},
		{
			name: "include only txt files",
			opts: &xtypes.PackagingOptions{
				IncludeFiles: []string{"**/*.txt"},
			},
			expectedFiles: []string{
				"aa/zz/xx.txt",
				"ee/bullhorn.txt",
				"ee/sans.txt",
			},
			expectedCount: 3,
		},
		{
			name: "include specific directory with **",
			opts: &xtypes.PackagingOptions{
				IncludeFiles: []string{"cc/**/*"},
			},
			expectedFiles: []string{
				"cc/eeeeeeeee.gg",
				"cc/nnn/ok.ok",
			},
			expectedCount: 2,
		},
		{
			name: "include multiple patterns",
			opts: &xtypes.PackagingOptions{
				IncludeFiles: []string{"aa/**/*", "bb/**/*"},
			},
			expectedFiles: []string{
				"aa/zz/xx.txt",
				"bb/abcisxyz.md",
			},
			expectedCount: 2,
		},
		{
			name: "exclude txt files",
			opts: &xtypes.PackagingOptions{
				IncludeFiles: []string{"**/*"},
				ExcludeFiles: []string{"**/*.txt"},
			},
			expectedFiles: []string{
				"bb/abcisxyz.md",
				"cc/eeeeeeeee.gg",
				"cc/nnn/ok.ok",
			},
			expectedCount: 3,
		},
		{
			name: "exclude specific directory",
			opts: &xtypes.PackagingOptions{
				IncludeFiles: []string{"**/*"},
				ExcludeFiles: []string{"cc/**/*"},
			},
			expectedFiles: []string{
				"aa/zz/xx.txt",
				"bb/abcisxyz.md",
				"ee/bullhorn.txt",
				"ee/sans.txt",
			},
			expectedCount: 4,
		},
		{
			name: "exclude nested directory",
			opts: &xtypes.PackagingOptions{
				IncludeFiles: []string{"**/*"},
				ExcludeFiles: []string{"cc/nnn/**/*"},
			},
			expectedFiles: []string{
				"aa/zz/xx.txt",
				"bb/abcisxyz.md",
				"cc/eeeeeeeee.gg",
				"ee/bullhorn.txt",
				"ee/sans.txt",
			},
			expectedCount: 5,
		},
		{
			name: "include with single asterisk",
			opts: &xtypes.PackagingOptions{
				IncludeFiles: []string{"ee/*.txt"},
			},
			expectedFiles: []string{
				"ee/bullhorn.txt",
				"ee/sans.txt",
			},
			expectedCount: 2,
		},
		{
			name: "include specific file",
			opts: &xtypes.PackagingOptions{
				IncludeFiles: []string{"bb/abcisxyz.md"},
			},
			expectedFiles: []string{
				"bb/abcisxyz.md",
			},
			expectedCount: 1,
		},
		{
			name: "no include patterns - should include everything",
			opts: &xtypes.PackagingOptions{
				IncludeFiles: []string{},
			},
			expectedFiles: []string{
				"aa/zz/xx.txt",
				"bb/abcisxyz.md",
				"cc/eeeeeeeee.gg",
				"cc/nnn/ok.ok",
				"ee/bullhorn.txt",
				"ee/sans.txt",
			},
			expectedCount: 6,
		},
		{
			name:          "nil opts - should return nil",
			opts:          nil,
			expectedCount: 0,
		},
		{
			name: "complex include and exclude",
			opts: &xtypes.PackagingOptions{
				IncludeFiles: []string{"**/*.txt", "**/*.md"},
				ExcludeFiles: []string{"aa/**/*"},
			},
			expectedFiles: []string{
				"bb/abcisxyz.md",
				"ee/bullhorn.txt",
				"ee/sans.txt",
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary zip file
			tmpFile, err := os.CreateTemp("", "test-package-*.zip")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			tmpFile.Close()

			// Create zip writer
			zipFile, err := os.Create(tmpFile.Name())
			if err != nil {
				t.Fatalf("Failed to create zip file: %v", err)
			}
			defer zipFile.Close()

			zipWriter := zip.NewWriter(zipFile)
			defer zipWriter.Close()

			// Run the function
			err = packageFilesV2(absTestdataPath, tt.opts, zipWriter)
			if err != nil {
				zipWriter.Close()
				zipFile.Close()
				if tt.expectError {
					if tt.errorSubstring != "" && !strings.Contains(err.Error(), tt.errorSubstring) {
						t.Errorf("Expected error containing %q, got %v", tt.errorSubstring, err)
					}
					return
				}
				t.Fatalf("packageFilesV2() error = %v", err)
			}

			// Close zip writer to finalize
			err = zipWriter.Close()
			if err != nil {
				t.Fatalf("Failed to close zip writer: %v", err)
			}
			zipFile.Close()

			// Read the zip file and verify contents
			zipReader, err := zip.OpenReader(tmpFile.Name())
			if err != nil {
				t.Fatalf("Failed to open zip file: %v", err)
			}
			defer zipReader.Close()

			// Collect file names from zip
			var zipFiles []string
			for _, f := range zipReader.File {
				zipFiles = append(zipFiles, f.Name)
			}

			// Sort for comparison
			sort.Strings(zipFiles)
			sort.Strings(tt.expectedFiles)

			// Check count
			if len(zipFiles) != tt.expectedCount {
				t.Errorf("Expected %d files, got %d. Files: %v", tt.expectedCount, len(zipFiles), zipFiles)
			}

			// Check specific files if expected
			if len(tt.expectedFiles) > 0 {
				if !reflect.DeepEqual(zipFiles, tt.expectedFiles) {
					t.Errorf("File lists don't match.\nExpected: %v\nGot:      %v", tt.expectedFiles, zipFiles)
				}
			}
		})
	}
}

func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		path        string
		shouldMatch bool
		expectError bool
	}{
		{
			name:        "simple wildcard",
			pattern:     "*.txt",
			path:        "file.txt",
			shouldMatch: true,
		},
		{
			name:        "simple wildcard no match",
			pattern:     "*.txt",
			path:        "file.md",
			shouldMatch: false,
		},
		{
			name:        "double asterisk recursive",
			pattern:     "**/*.txt",
			path:        "aa/bb/file.txt",
			shouldMatch: true,
		},
		{
			name:        "double asterisk at root",
			pattern:     "**/*.txt",
			path:        "file.txt",
			shouldMatch: true,
		},
		{
			name:        "double asterisk nested",
			pattern:     "aa/**/*",
			path:        "aa/bb/cc/file.txt",
			shouldMatch: true,
		},
		{
			name:        "question mark single char",
			pattern:     "file?.txt",
			path:        "file1.txt",
			shouldMatch: true,
		},
		{
			name:        "question mark no match",
			pattern:     "file?.txt",
			path:        "file12.txt",
			shouldMatch: false,
		},
		{
			name:        "exact path match",
			pattern:     "aa/zz/xx.txt",
			path:        "aa/zz/xx.txt",
			shouldMatch: true,
		},
		{
			name:        "exact path no match",
			pattern:     "aa/zz/xx.txt",
			path:        "aa/zz/yy.txt",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := globToRegex(tt.pattern)
			if err != nil {
				if !tt.expectError {
					t.Errorf("globToRegex() error = %v", err)
				}
				return
			}
			if tt.expectError {
				t.Errorf("Expected error but got none")
				return
			}

			matched := regex.MatchString(tt.path)
			if matched != tt.shouldMatch {
				t.Errorf("Pattern %q matching %q: expected %v, got %v", tt.pattern, tt.path, tt.shouldMatch, matched)
			}
		})
	}
}
