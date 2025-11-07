package cli

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/models"
	"github.com/pelletier/go-toml/v2"
)

func (c *PackageBuildCmd) Run(_ *kong.Context) error {
	potatoTomlFile, err := os.ReadFile(c.PotatoTomlFile)
	if err != nil {
		return err
	}

	potatoToml := models.PotatoPackage{}
	err = toml.Unmarshal(potatoTomlFile, &potatoToml)
	if err != nil {
		return err
	}

	if c.OutputZipFile == "" {
		c.OutputZipFile = fmt.Sprintf("%s.zip", potatoToml.Slug)
	}

	zipFile, err := os.Create(c.OutputZipFile)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	potatoFileDir := path.Dir(c.PotatoTomlFile)

	err = includeSubFolder(potatoFileDir, potatoFileDir, potatoToml.FilesDir, zipWriter)
	if err != nil {
		return err
	}

	potatoToml.FilesDir = ""
	potatoToml.DevToken = ""

	potatoJson, err := json.Marshal(potatoToml)
	if err != nil {
		return err
	}

	pfile, err := zipWriter.Create("potato.json")
	if err != nil {
		return err
	}
	_, err = pfile.Write(potatoJson)
	if err != nil {
		return err
	}

	err = zipWriter.Close()
	if err != nil {
		return err
	}

	fmt.Printf("Package built successfully: %s\n", c.OutputZipFile)

	return nil
}

func includeSubFolder(basePath, folder, name string, zipWriter *zip.Writer) error {

	fullPath := path.Join(folder, name)

	files, err := os.ReadDir(fullPath)
	if err != nil {
		return err
	}

	for _, file := range files {

		if file.IsDir() {
			err = includeSubFolder(basePath, fullPath, file.Name(), zipWriter)
			if err != nil {
				return err
			}
			continue
		}

		// Create the relative path for this file within the zip
		filePath := path.Join(fullPath, file.Name())
		targetPath := strings.TrimPrefix(filePath, basePath)
		targetPath = strings.TrimPrefix(targetPath, "/")

		zfile, err := zipWriter.Create(targetPath)
		if err != nil {
			return err
		}

		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		_, err = zfile.Write(fileData)
		if err != nil {
			return err
		}

	}

	return nil
}

func packageFilesV2(basePath string, opts *xtypes.PackagingOptions, zipWriter *zip.Writer) error {
	if opts == nil {
		return nil
	}

	// Normalize basePath to absolute path
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return err
	}

	// Convert glob patterns to regex patterns for matching
	includePatterns := make([]*regexp.Regexp, 0, len(opts.IncludeFiles))
	for _, pattern := range opts.IncludeFiles {
		regex, err := globToRegex(pattern)
		if err != nil {
			return fmt.Errorf("invalid include pattern %q: %w", pattern, err)
		}
		includePatterns = append(includePatterns, regex)
	}

	excludePatterns := make([]*regexp.Regexp, 0, len(opts.ExcludeFiles))
	for _, pattern := range opts.ExcludeFiles {
		regex, err := globToRegex(pattern)
		if err != nil {
			return fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
		}
		excludePatterns = append(excludePatterns, regex)
	}

	// Walk the directory tree
	return filepath.Walk(absBasePath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories themselves, only process files
		if info.IsDir() {
			return nil
		}

		// Get relative path from basePath
		relPath, err := filepath.Rel(absBasePath, filePath)
		if err != nil {
			return err
		}

		// Normalize path separators to forward slashes for pattern matching
		normalizedPath := filepath.ToSlash(relPath)

		// Check if file matches any include pattern
		matchesInclude := false
		if len(includePatterns) == 0 {
			// If no include patterns, include everything
			matchesInclude = true
		} else {
			for _, pattern := range includePatterns {
				if pattern.MatchString(normalizedPath) {
					matchesInclude = true
					break
				}
			}
		}

		if !matchesInclude {
			return nil
		}

		// Check if file matches any exclude pattern
		for _, pattern := range excludePatterns {
			if pattern.MatchString(normalizedPath) {
				// Excluded, skip this file
				return nil
			}
		}

		// File should be included, add it to zip
		// Use forward slashes in zip paths (zip standard)
		zipPath := filepath.ToSlash(relPath)

		zfile, err := zipWriter.Create(zipPath)
		if err != nil {
			return err
		}

		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		_, err = zfile.Write(fileData)
		return err
	})
}

// globToRegex converts a glob pattern (supporting *, ?, and **) to a regex pattern
func globToRegex(pattern string) (*regexp.Regexp, error) {
	// Normalize path separators
	pattern = filepath.ToSlash(pattern)

	// Escape special regex characters except *, ?, and \
	escaped := ""
	for i := 0; i < len(pattern); i++ {
		char := pattern[i]
		switch char {
		case '*':
			// Check if it's ** (double asterisk)
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				// Check what comes after **
				if i+2 < len(pattern) && pattern[i+2] == '/' {
					// **/ means zero or more directory segments followed by /
					// This should match files at any depth, including directly in the directory
					escaped += `(.*/)?`
					i += 2 // Skip ** and /
				} else if i+2 < len(pattern) && pattern[i+2] == '*' {
					// *** is invalid, but handle as **
					escaped += `.*`
					i++ // Skip one more *
				} else {
					// ** at end or followed by something else - matches zero or more directories
					escaped += `.*`
					i++ // Skip the next *
				}
			} else {
				// Single * matches any sequence of non-separator characters
				escaped += `[^/]*`
			}
		case '?':
			// ? matches any single non-separator character
			escaped += `[^/]`
		case '.', '+', '(', ')', '[', ']', '{', '}', '^', '$', '|':
			// Escape regex special characters
			escaped += `\` + string(char)
		case '\\':
			// Handle backslash - if followed by special char, keep as is, otherwise escape
			if i+1 < len(pattern) {
				next := pattern[i+1]
				if next == '*' || next == '?' || next == '\\' {
					escaped += string(char) + string(next)
					i++ // Skip the next character
				} else {
					escaped += `\\`
				}
			} else {
				escaped += `\\`
			}
		default:
			escaped += string(char)
		}
	}

	// Anchor to start and end for exact matching
	regexPattern := "^" + escaped + "$"
	return regexp.Compile(regexPattern)
}
