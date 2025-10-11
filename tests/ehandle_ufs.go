package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/engine"
	"github.com/blue-monads/turnix/backend/engine/executors"
	"github.com/blue-monads/turnix/backend/services/datahub/database"
)

func HandleUfsTest() {

	// create sqlite db

	fmt.Println("@test_start")
	defer fmt.Println("@test_end")

	sdb, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer sdb.Close()

	database, err := database.FromSqlHandle(sdb)
	if err != nil {
		log.Fatalf("Failed to create database instance: %v", err)
	}

	ezip, err := engine.ZipEPackage("example1")
	if err != nil {
		log.Fatalf("Failed to zip epackage: %v", err)
	}

	tmpPath := "./tmp/example1"

	os.ReadFile(tmpPath)
	os.MkdirAll(tmpPath, 0755)

	rfs, err := os.OpenRoot(tmpPath)
	if err != nil {
		log.Fatalf("Failed to open root: %v", err)
	}

	pkgId, err := actions.InstallPackageByFile(database, slog.Default(), 1, ezip)
	if err != nil {
		log.Fatalf("Failed to install package: %v", err)
		return
	}

	spaces, err := database.ListSpaces()
	if err != nil {
		log.Fatalf("Failed to list spaces: %v", err)
		return
	}

	spaceId := int64(0)

	for _, space := range spaces {
		if space.PackageID == pkgId {
			spaceId = space.ID
		}
	}

	EHandle := executors.EHandle{
		Database:  database,
		FsRoot:    rfs,
		SpaceId:   spaceId,
		PackageId: pkgId,
		App:       nil,
	}

	// Test 1: List root directory
	fmt.Println("\n=== Test 1: List Root Directory ===")
	rootFiles, err := EHandle.ListFiles("/")
	if err != nil {
		log.Fatalf("Failed to list root: %v", err)
	}
	fmt.Printf("Root directories: %d items\n", len(rootFiles))
	for _, f := range rootFiles {
		fmt.Printf("  - %s (folder: %v)\n", f.Name, f.IsFolder)
	}

	// Test 2: List package root directory to see what's available
	fmt.Println("\n=== Test 2: List Package Root Directory ===")
	pkgRootFiles, err := EHandle.ListFiles("/pkg")
	if err != nil {
		log.Fatalf("Failed to list package root: %v", err)
	}
	fmt.Printf("Package root files: %d items\n", len(pkgRootFiles))
	for _, f := range pkgRootFiles {
		fmt.Printf("  - %s (folder: %v, size: %d)\n", f.Name, f.IsFolder, f.Size)
	}

	// Test 3: Read package file (if any exist)
	fmt.Println("\n=== Test 3: Read Package File ===")
	var testFilePath string
	var foundTestFile bool

	// Try to find a readable file in the package
	for _, f := range pkgRootFiles {
		if !f.IsFolder && f.Name != "" {
			testFilePath = "/pkg/" + f.Name
			foundTestFile = true
			break
		}
		if f.IsFolder && f.Name == "public" {
			// Check public folder
			publicFiles, err := EHandle.ListFiles("/pkg/public")
			if err == nil && len(publicFiles) > 0 {
				for _, pf := range publicFiles {
					if !pf.IsFolder {
						testFilePath = "/pkg/public/" + pf.Name
						foundTestFile = true
						break
					}
				}
			}
		}
		if foundTestFile {
			break
		}
	}

	if foundTestFile {
		content, err := EHandle.ReadFile(testFilePath)
		if err != nil {
			log.Fatalf("Failed to read package file %s: %v", testFilePath, err)
		}
		fmt.Printf("Package file %s content (%d bytes):\n%s\n", testFilePath, len(content), string(content))
	} else {
		fmt.Println("No readable files found in package, skipping read test")
	}

	// Test 4: List package directory (public if it exists)
	fmt.Println("\n=== Test 4: List Package Directory ===")
	hasPublicDir := false
	for _, f := range pkgRootFiles {
		if f.IsFolder && f.Name == "public" {
			hasPublicDir = true
			break
		}
	}

	if hasPublicDir {
		pkgFiles, err := EHandle.ListFiles("/pkg/public")
		if err != nil {
			log.Fatalf("Failed to list package directory: %v", err)
		}
		fmt.Printf("Package /public files: %d items\n", len(pkgFiles))
		for _, f := range pkgFiles {
			fmt.Printf("  - %s (folder: %v, size: %d)\n", f.Name, f.IsFolder, f.Size)
		}
	} else {
		fmt.Println("No public directory found in package")
	}

	// Test 5: Check if package file exists
	fmt.Println("\n=== Test 5: Check Package File Exists ===")
	if foundTestFile {
		exists, err := EHandle.Exists(testFilePath)
		if err != nil {
			log.Fatalf("Failed to check file existence: %v", err)
		}
		fmt.Printf("Package file %s exists: %v\n", testFilePath, exists)
	} else {
		fmt.Println("Skipping existence check (no test file found)")
	}

	// Test 6: Write file to /home
	fmt.Println("\n=== Test 6: Write File to /home ===")
	testContent := []byte("Hello from unified file system test!")
	err = EHandle.WriteFile("/home/test.txt", testContent)
	if err != nil {
		log.Fatalf("Failed to write file to /home: %v", err)
	}
	fmt.Println("Successfully wrote file to /home/test.txt")

	// Test 7: Read file from /home
	fmt.Println("\n=== Test 7: Read File from /home ===")
	homeContent, err := EHandle.ReadFile("/home/test.txt")
	if err != nil {
		log.Fatalf("Failed to read file from /home: %v", err)
	}
	fmt.Printf("Home file content: %s\n", string(homeContent))

	// Test 8: Check if /home file exists
	fmt.Println("\n=== Test 8: Check /home File Exists ===")
	homeExists, err := EHandle.Exists("/home/test.txt")
	if err != nil {
		log.Fatalf("Failed to check home file existence: %v", err)
	}
	fmt.Printf("/home file exists: %v\n", homeExists)

	// Test 9: Create directory in /home
	fmt.Println("\n=== Test 9: Create Directory in /home ===")
	err = EHandle.Mkdir("/home/testdir")
	if err != nil {
		log.Fatalf("Failed to create directory in /home: %v", err)
	}
	fmt.Println("Successfully created directory /home/testdir")

	// Test 10: Write file in subdirectory
	fmt.Println("\n=== Test 10: Write File in Subdirectory ===")
	err = EHandle.WriteFile("/home/testdir/nested.txt", []byte("nested content"))
	if err != nil {
		log.Fatalf("Failed to write file in subdirectory: %v", err)
	}
	fmt.Println("Successfully wrote file to /home/testdir/nested.txt")

	// Test 11: List /home directory
	fmt.Println("\n=== Test 11: List /home Directory ===")
	homeFiles, err := EHandle.ListFiles("/home")
	if err != nil {
		log.Fatalf("Failed to list /home directory: %v", err)
	}
	fmt.Printf("/home files: %d items\n", len(homeFiles))
	for _, f := range homeFiles {
		fmt.Printf("  - %s (folder: %v, size: %d)\n", f.Name, f.IsFolder, f.Size)
	}

	// Test 12: Write to /tmp
	fmt.Println("\n=== Test 12: Write File to /tmp ===")
	err = EHandle.WriteFile("/tmp/tmpfile.txt", []byte("temporary content"))
	if err != nil {
		log.Fatalf("Failed to write file to /tmp: %v", err)
	}
	fmt.Println("Successfully wrote file to /tmp/tmpfile.txt")

	// Test 13: Read from /tmp
	fmt.Println("\n=== Test 13: Read File from /tmp ===")
	tmpContent, err := EHandle.ReadFile("/tmp/tmpfile.txt")
	if err != nil {
		log.Fatalf("Failed to read file from /tmp: %v", err)
	}
	fmt.Printf("/tmp file content: %s\n", string(tmpContent))

	// Test 14: Create directory in /tmp
	fmt.Println("\n=== Test 14: Create Directory in /tmp ===")
	err = EHandle.Mkdir("/tmp/tmpdir")
	if err != nil {
		log.Fatalf("Failed to create directory in /tmp: %v", err)
	}
	fmt.Println("Successfully created directory /tmp/tmpdir")

	// Test 15: List /tmp directory
	fmt.Println("\n=== Test 15: List /tmp Directory ===")
	tmpFiles, err := EHandle.ListFiles("/tmp")
	if err != nil {
		log.Fatalf("Failed to list /tmp directory: %v", err)
	}
	fmt.Printf("/tmp files: %d items\n", len(tmpFiles))
	for _, f := range tmpFiles {
		fmt.Printf("  - %s (folder: %v, size: %d)\n", f.Name, f.IsFolder, f.Size)
	}

	// Test 16: Check /tmp file exists
	fmt.Println("\n=== Test 16: Check /tmp File Exists ===")
	tmpExists, err := EHandle.Exists("/tmp/tmpfile.txt")
	if err != nil {
		log.Fatalf("Failed to check /tmp file existence: %v", err)
	}
	fmt.Printf("/tmp file exists: %v\n", tmpExists)

	// Test 17: Remove file from /home
	fmt.Println("\n=== Test 17: Remove File from /home ===")
	err = EHandle.RemoveFile("/home/test.txt")
	if err != nil {
		log.Fatalf("Failed to remove file from /home: %v", err)
	}
	fmt.Println("Successfully removed /home/test.txt")

	// Test 18: Verify file was removed
	fmt.Println("\n=== Test 18: Verify File Removal ===")
	removedExists, err := EHandle.Exists("/home/test.txt")
	if err != nil {
		log.Fatalf("Failed to check removed file: %v", err)
	}
	fmt.Printf("Removed file exists: %v (should be false)\n", removedExists)

	// Test 19: Remove directory from /home
	fmt.Println("\n=== Test 19: Remove Directory from /home ===")
	err = EHandle.Rmdir("/home/testdir")
	if err != nil {
		log.Fatalf("Failed to remove directory from /home: %v", err)
	}
	fmt.Println("Successfully removed /home/testdir")

	// Test 20: Remove file from /tmp
	fmt.Println("\n=== Test 20: Remove File from /tmp ===")
	err = EHandle.RemoveFile("/tmp/tmpfile.txt")
	if err != nil {
		log.Fatalf("Failed to remove file from /tmp: %v", err)
	}
	fmt.Println("Successfully removed /tmp/tmpfile.txt")

	// Test 21: Remove directory from /tmp
	fmt.Println("\n=== Test 21: Remove Directory from /tmp ===")
	err = EHandle.Rmdir("/tmp/tmpdir")
	if err != nil {
		log.Fatalf("Failed to remove directory from /tmp: %v", err)
	}
	fmt.Println("Successfully removed /tmp/tmpdir")

	// Test 22: Error cases - Invalid path (no leading slash)
	fmt.Println("\n=== Test 22: Error Case - Invalid Path ===")
	_, err = EHandle.ReadFile("home/test.txt")
	if err != nil {
		fmt.Printf("Expected error for path without leading slash: %v\n", err)
	} else {
		log.Fatalf("Should have failed for path without leading slash")
	}

	// Test 23: Error case - Try to write to read-only /pkg
	fmt.Println("\n=== Test 23: Error Case - Write to Read-Only /pkg ===")
	err = EHandle.WriteFile("/pkg/test.txt", []byte("should fail"))
	if err != nil {
		fmt.Printf("Expected error for writing to /pkg: %v\n", err)
	} else {
		log.Fatalf("Should have failed for writing to /pkg")
	}

	// Test 24: Error case - Try to read root directory
	fmt.Println("\n=== Test 24: Error Case - Read Root Directory ===")
	_, err = EHandle.ReadFile("/")
	if err != nil {
		fmt.Printf("Expected error for reading root: %v\n", err)
	} else {
		log.Fatalf("Should have failed for reading root")
	}

	// Test 25: Share a file from /home
	fmt.Println("\n=== Test 25: Share File from /home ===")
	// First create a file to share
	shareTestContent := []byte("This file will be shared")
	err = EHandle.WriteFile("/home/shareable.txt", shareTestContent)
	if err != nil {
		log.Fatalf("Failed to create file for sharing: %v", err)
	}
	fmt.Println("Created file /home/shareable.txt")

	// Now share it
	shareId, err := EHandle.ShareFile(spaceId, "/home/shareable.txt")
	if err != nil {
		log.Fatalf("Failed to share file: %v", err)
	}
	fmt.Printf("Successfully shared file, share ID: %s\n", shareId)

	// Test 26: Error case - Try to share file from /pkg
	fmt.Println("\n=== Test 26: Error Case - Share File from /pkg ===")
	if foundTestFile {
		_, err = EHandle.ShareFile(0, testFilePath)
		if err != nil {
			fmt.Printf("Expected error for sharing from /pkg: %v\n", err)
		} else {
			log.Fatalf("Should have failed for sharing from /pkg")
		}
	} else {
		fmt.Println("Skipping (no test file found in /pkg)")
	}

	// Test 27: Error case - Try to share file from /tmp
	fmt.Println("\n=== Test 27: Error Case - Share File from /tmp ===")
	// Create a temp file first
	err = EHandle.WriteFile("/tmp/temp_shareable.txt", []byte("temp content"))
	if err != nil {
		log.Fatalf("Failed to create temp file: %v", err)
	}

	_, err = EHandle.ShareFile(spaceId, "/tmp/temp_shareable.txt")
	if err != nil {
		fmt.Printf("Expected error for sharing from /tmp: %v\n", err)
	} else {
		log.Fatalf("Should have failed for sharing from /tmp")
	}

	// Clean up the temp file
	err = EHandle.RemoveFile("/tmp/temp_shareable.txt")
	if err != nil {
		log.Fatalf("Failed to clean up temp file: %v", err)
	}

	// Test 28: Clean up the shared file
	fmt.Println("\n=== Test 28: Clean Up Shared File ===")
	err = EHandle.RemoveFile("/home/shareable.txt")
	if err != nil {
		log.Fatalf("Failed to remove shared file: %v", err)
	}
	fmt.Println("Successfully removed shared file")

	fmt.Println("\n=== All Tests Passed Successfully! ===")

}
