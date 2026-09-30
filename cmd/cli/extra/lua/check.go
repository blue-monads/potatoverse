package luaextra

import (
	"errors"
	"fmt"
	"os"

	lua "github.com/yuin/gopher-lua"
)

// CheckResult holds the verification result for a single file.
type CheckResult struct {
	File  string
	Error error
}

// CheckFile verifies the syntax and compilation of a Lua file without executing it.
func CheckFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	L := lua.NewState(lua.Options{
		SkipOpenLibs: true,
	})
	defer L.Close()

	_, err = L.LoadFile(filePath)
	return err
}

// CheckFiles verifies multiple Lua files and returns a list of results.
func CheckFiles(filePaths []string) []CheckResult {
	results := make([]CheckResult, 0, len(filePaths))
	for _, fp := range filePaths {
		err := CheckFile(fp)
		results = append(results, CheckResult{
			File:  fp,
			Error: err,
		})
	}
	return results
}

// RunCheck handles the "lua check" command with arguments.
func RunCheck(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: potatoverse extra lua check <file1.lua> [file2.lua ...]")
	}

	results := CheckFiles(args)
	hasError := false

	for _, res := range results {
		if res.Error != nil {
			hasError = true
			fmt.Fprintf(os.Stderr, "✖ [FAIL] %s: %v\n", res.File, res.Error)
		} else {
			fmt.Printf("✔ [OK]   %s: syntax OK\n", res.File)
		}
	}

	if hasError {
		return fmt.Errorf("lua check failed for one or more files")
	}

	if len(results) > 1 {
		fmt.Printf("\nAll %d file(s) checked successfully.\n", len(results))
	}
	return nil
}
