package cli

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blue-monads/turnix/backend/xtypes/models"
	"github.com/pelletier/go-toml/v2"
)

func readPotatoToml(potatoTomlFile string) (*models.PotatoPackage, error) {
	potatoTomlFileData, err := os.ReadFile(potatoTomlFile)
	if err != nil {
		return nil, err
	}
	potatoToml := models.PotatoPackage{}
	err = toml.Unmarshal(potatoTomlFileData, &potatoToml)
	return &potatoToml, nil
}

// includePatternInfo holds information about an include pattern and its optional destination
type includePatternInfo struct {
	sourcePattern string
	destPath      string
	regex         *regexp.Regexp
}

func packageFilesV2(basePath string, opts *models.PackagingOptions, zipWriter *zip.Writer) error {

	// Normalize basePath to absolute path
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return err
	}

	// Handle nil opts
	if opts == nil {
		return nil
	}

	// Parse include patterns - support comma-separated source,destination syntax
	includePatterns := make([]includePatternInfo, 0, len(opts.IncludeFiles))
	for _, patternEntry := range opts.IncludeFiles {
		var sourcePattern, destPath string

		// Check if pattern contains a comma (source,destination syntax)
		if idx := strings.Index(patternEntry, ","); idx != -1 {
			sourcePattern = strings.TrimSpace(patternEntry[:idx])
			destPath = strings.TrimSpace(patternEntry[idx+1:])
		} else {
			sourcePattern = patternEntry
			destPath = "" // No destination, use original path
		}

		regex, err := globToRegex(sourcePattern)
		if err != nil {
			return fmt.Errorf("invalid include pattern %q: %w", sourcePattern, err)
		}

		includePatterns = append(includePatterns, includePatternInfo{
			sourcePattern: sourcePattern,
			destPath:      destPath,
			regex:         regex,
		})
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
		var matchedPattern *includePatternInfo
		if len(includePatterns) == 0 {
			// If no include patterns, include everything (no destination mapping)
			matchedPattern = nil
		} else {
			for i := range includePatterns {
				if includePatterns[i].regex.MatchString(normalizedPath) {
					matchedPattern = &includePatterns[i]
					break
				}
			}
			if matchedPattern == nil {
				// File doesn't match any include pattern
				return nil
			}
		}

		// Check if file matches any exclude pattern
		for _, pattern := range excludePatterns {
			if pattern.MatchString(normalizedPath) {
				// Excluded, skip this file
				return nil
			}
		}

		// Determine the zip path
		var zipPath string
		if matchedPattern != nil && matchedPattern.destPath != "" {
			// Transform path based on destination
			zipPath = transformPath(normalizedPath, matchedPattern.sourcePattern, matchedPattern.destPath)
		} else {
			// Use original path
			zipPath = normalizedPath
		}

		// Ensure forward slashes in zip paths (zip standard)
		zipPath = filepath.ToSlash(zipPath)

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

// transformPath transforms a matched file path based on the source pattern and destination
// Examples:
//   - "server.lua" with pattern "server.lua" -> "server_v1.lua" (exact match, use destination as-is)
//   - "public/index.html" with pattern "public/**/*" -> "newfolder/index.html" (extract suffix, prepend destination)
func transformPath(filePath, sourcePattern, destPath string) string {
	// Normalize paths
	filePath = filepath.ToSlash(filePath)
	sourcePattern = filepath.ToSlash(sourcePattern)

	// Check if it's an exact match (no wildcards in the pattern)
	// For exact matches, just return the destination
	if !strings.ContainsAny(sourcePattern, "*?") {
		if filePath == sourcePattern {
			return destPath
		}
		// Not an exact match, fall through to glob handling
	}

	// For glob patterns, find the longest literal prefix (before any wildcards)
	// and extract the suffix to prepend with destination
	literalPrefix := extractLiteralPrefix(sourcePattern)

	if literalPrefix != "" && strings.HasPrefix(filePath, literalPrefix) {
		// Extract the part after the literal prefix
		suffix := filePath[len(literalPrefix):]
		// Remove leading slash if present
		suffix = strings.TrimPrefix(suffix, "/")

		// Combine destination with suffix
		if suffix == "" {
			// File is exactly at the prefix boundary
			return destPath
		}
		if destPath == "" {
			return suffix
		}
		// Ensure destination ends with / if it's a directory
		if !strings.HasSuffix(destPath, "/") && !strings.HasSuffix(destPath, "\\") {
			return destPath + "/" + suffix
		}
		return destPath + suffix
	}

	// Fallback: if we can't determine the transformation, use destination as-is
	// This handles cases where the pattern doesn't match the expected structure
	return destPath
}

// extractLiteralPrefix extracts the longest literal path prefix before any wildcards
// Examples:
//   - "public/**/*" -> "public/"
//   - "public/css/*.css" -> "public/css/"
//   - "server.lua" -> "server.lua"
//   - "*.txt" -> ""
func extractLiteralPrefix(pattern string) string {
	pattern = filepath.ToSlash(pattern)

	// Find the first wildcard character
	wildcardIdx := -1
	for i, char := range pattern {
		if char == '*' || char == '?' {
			wildcardIdx = i
			break
		}
	}

	if wildcardIdx == -1 {
		// No wildcards, return the whole pattern
		return pattern
	}

	// Extract the prefix up to (but not including) the wildcard
	prefix := pattern[:wildcardIdx]

	// If the prefix doesn't end with a slash, find the last slash
	// This ensures we get a directory prefix
	if lastSlash := strings.LastIndex(prefix, "/"); lastSlash != -1 {
		prefix = prefix[:lastSlash+1]
	} else {
		// No slash found, return empty (no meaningful prefix)
		prefix = ""
	}

	return prefix
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
