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
	"github.com/blue-monads/turnix/backend/engine/executors/luaz/binds"
	"github.com/blue-monads/turnix/backend/services/datahub/database"
	lua "github.com/yuin/gopher-lua"
)

// HandleLuazUfsTest tests the UFS Lua bindings by actually running Lua code
func HandleLuazUfsTest() {
	fmt.Println("@luaz_ufs_test_start")
	defer fmt.Println("@luaz_ufs_test_end")

	// Setup database
	sdb, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer sdb.Close()

	db, err := database.FromSqlHandle(sdb)
	if err != nil {
		log.Fatalf("Failed to create database instance: %v", err)
	}

	// Setup package
	ezip, err := engine.ZipEPackage("example1")
	if err != nil {
		log.Fatalf("Failed to zip epackage: %v", err)
	}

	tmpPath := "./tmp/luaz_test"
	os.RemoveAll(tmpPath)
	os.MkdirAll(tmpPath, 0755)
	defer os.RemoveAll(tmpPath)

	rfs, err := os.OpenRoot(tmpPath)
	if err != nil {
		log.Fatalf("Failed to open root: %v", err)
	}

	pkgId, err := actions.InstallPackageByFile(db, slog.Default(), 1, ezip)
	if err != nil {
		log.Fatalf("Failed to install package: %v", err)
	}

	spaces, err := db.ListSpaces()
	if err != nil {
		log.Fatalf("Failed to list spaces: %v", err)
	}

	spaceId := int64(0)
	for _, space := range spaces {
		if space.PackageID == pkgId {
			spaceId = space.ID
			break
		}
	}

	// Create EHandle
	eHandle := &executors.EHandle{
		Database:  db,
		FsRoot:    rfs,
		SpaceId:   spaceId,
		PackageId: pkgId,
		App:       nil,
		Logger:    slog.Default(),
	}

	// Create Lua state
	L := lua.NewState()
	defer L.Close()

	// Register UFS module
	L.PreloadModule("ufs", binds.UfsModule(eHandle))

	// Test 1: List root directories
	fmt.Println("\n=== Test 1: List Root Directories (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local dirs, err = ufs.list("/")
		if err then
			error("Failed to list root: " .. err)
		end
		print("Root directories: " .. #dirs .. " items")
		for i, dir in ipairs(dirs) do
			print(string.format("  - %s (folder: %s)", dir.name, tostring(dir.is_folder)))
		end
	`)
	if err != nil {
		log.Fatalf("Test 1 failed: %v", err)
	}

	// Test 2: Create directory in /home
	fmt.Println("\n=== Test 2: Create Directory (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local ok, err = ufs.mkdir(1, "/home/lua_test")
		if err then
			error("Failed to create directory: " .. err)
		end
		print("Successfully created /home/lua_test")
	`)
	if err != nil {
		log.Fatalf("Test 2 failed: %v", err)
	}

	// Test 3: Check if directory exists
	fmt.Println("\n=== Test 3: Check Existence (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local exists, err = ufs.exists("/home/lua_test")
		if err then
			error("Failed to check existence: " .. err)
		end
		print(string.format("Directory exists: %s", tostring(exists)))
		if not exists then
			error("Directory should exist but doesn't")
		end
	`)
	if err != nil {
		log.Fatalf("Test 3 failed: %v", err)
	}

	// Test 4: Write file
	fmt.Println("\n=== Test 4: Write File (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local content = "Hello from Lua UFS test!\nLine 2\nLine 3"
		local ok, err = ufs.write(1, "/home/lua_test/test.txt", content)
		if err then
			error("Failed to write file: " .. err)
		end
		print("Successfully wrote /home/lua_test/test.txt")
	`)
	if err != nil {
		log.Fatalf("Test 4 failed: %v", err)
	}

	// Test 5: Read file
	fmt.Println("\n=== Test 5: Read File (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local content, err = ufs.read("/home/lua_test/test.txt")
		if err then
			error("Failed to read file: " .. err)
		end
		print("File content:")
		print(content)
		if not content:match("Hello from Lua UFS test!") then
			error("Content doesn't match expected")
		end
	`)
	if err != nil {
		log.Fatalf("Test 5 failed: %v", err)
	}

	// Test 6: List files in directory
	fmt.Println("\n=== Test 6: List Files in Directory (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local files, err = ufs.list("/home/lua_test")
		if err then
			error("Failed to list files: " .. err)
		end
		print(string.format("Files in /home/lua_test: %d items", #files))
		for i, file in ipairs(files) do
			print(string.format("  - %s (size: %d, folder: %s)", 
				file.name, file.size or 0, tostring(file.is_folder)))
		end
	`)
	if err != nil {
		log.Fatalf("Test 6 failed: %v", err)
	}

	// Test 7: Path utilities
	fmt.Println("\n=== Test 7: Path Utilities (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local path = "/home/lua_test/test.txt"
		local dir = ufs.dirname(path)
		local base = ufs.basename(path)
		print(string.format("Full path: %s", path))
		print(string.format("Directory: %s", dir))
		print(string.format("Basename: %s", base))
		if dir ~= "/home/lua_test" then
			error("dirname returned unexpected result: " .. dir)
		end
		if base ~= "test.txt" then
			error("basename returned unexpected result: " .. base)
		end
	`)
	if err != nil {
		log.Fatalf("Test 7 failed: %v", err)
	}

	// Test 8: Share file
	fmt.Println("\n=== Test 8: Share File (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local shareId, err = ufs.share(1, "/home/lua_test/test.txt")
		if err then
			error("Failed to share file: " .. err)
		end
		print(string.format("Share ID: %s", shareId))
		if shareId == "" then
			error("Share ID should not be empty")
		end
	`)
	if err != nil {
		log.Fatalf("Test 8 failed: %v", err)
	}

	// Test 9: Work with /tmp
	fmt.Println("\n=== Test 9: Work with /tmp (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		
		-- Write to /tmp
		local ok, err = ufs.write(1, "/tmp/lua_temp.txt", "Temporary data from Lua")
		if err then
			error("Failed to write to /tmp: " .. err)
		end
		print("Wrote /tmp/lua_temp.txt")
		
		-- Read from /tmp
		local content, err = ufs.read("/tmp/lua_temp.txt")
		if err then
			error("Failed to read from /tmp: " .. err)
		end
		print(string.format("Temp content: %s", content))
		
		-- Check existence
		local exists, err = ufs.exists("/tmp/lua_temp.txt")
		if err then
			error("Failed to check /tmp file: " .. err)
		end
		if not exists then
			error("Temp file should exist")
		end
	`)
	if err != nil {
		log.Fatalf("Test 9 failed: %v", err)
	}

	// Test 10: List package files
	fmt.Println("\n=== Test 10: List Package Files (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local files, err = ufs.list("/pkg")
		if err then
			error("Failed to list package files: " .. err)
		end
		print(string.format("Package files: %d items", #files))
		for i, file in ipairs(files) do
			print(string.format("  - %s (folder: %s)", file.name, tostring(file.is_folder)))
		end
	`)
	if err != nil {
		log.Fatalf("Test 10 failed: %v", err)
	}

	// Test 11: Error handling - Write to /pkg (should fail)
	fmt.Println("\n=== Test 11: Error Handling - Write to /pkg (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local ok, err = ufs.write(1, "/pkg/test.txt", "should fail")
		if not err then
			error("Writing to /pkg should have failed but didn't")
		end
		print("Expected error for writing to /pkg: " .. err)
	`)
	if err != nil {
		log.Fatalf("Test 11 failed: %v", err)
	}

	// Test 12: Error handling - Invalid path
	fmt.Println("\n=== Test 12: Error Handling - Invalid Path (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local content, err = ufs.read("home/test.txt")
		if not err then
			error("Invalid path should have failed but didn't")
		end
		print("Expected error for invalid path: " .. err)
	`)
	if err != nil {
		log.Fatalf("Test 12 failed: %v", err)
	}

	// Test 13: Error handling - Share from /tmp (should fail)
	fmt.Println("\n=== Test 13: Error Handling - Share from /tmp (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local shareId, err = ufs.share(1, "/tmp/lua_temp.txt")
		if not err then
			error("Sharing from /tmp should have failed but didn't")
		end
		print("Expected error for sharing from /tmp: " .. err)
	`)
	if err != nil {
		log.Fatalf("Test 13 failed: %v", err)
	}

	// Test 14: Remove file
	fmt.Println("\n=== Test 14: Remove File (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local ok, err = ufs.remove(1, "/home/lua_test/test.txt")
		if err then
			error("Failed to remove file: " .. err)
		end
		print("Successfully removed /home/lua_test/test.txt")
		
		-- Verify it's gone
		local exists, err = ufs.exists("/home/lua_test/test.txt")
		if err then
			error("Failed to check removed file: " .. err)
		end
		if exists then
			error("File should not exist after removal")
		end
		print("Verified file was removed")
	`)
	if err != nil {
		log.Fatalf("Test 14 failed: %v", err)
	}

	// Test 15: Remove directory
	fmt.Println("\n=== Test 15: Remove Directory (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local ok, err = ufs.rmdir(1, "/home/lua_test")
		if err then
			error("Failed to remove directory: " .. err)
		end
		print("Successfully removed /home/lua_test")
		
		-- Verify it's gone
		local exists, err = ufs.exists("/home/lua_test")
		if err then
			error("Failed to check removed directory: " .. err)
		end
		if exists then
			error("Directory should not exist after removal")
		end
		print("Verified directory was removed")
	`)
	if err != nil {
		log.Fatalf("Test 15 failed: %v", err)
	}

	// Test 16: Cleanup /tmp
	fmt.Println("\n=== Test 16: Cleanup /tmp (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		local ok, err = ufs.remove(1, "/tmp/lua_temp.txt")
		if err then
			error("Failed to remove /tmp file: " .. err)
		end
		print("Successfully removed /tmp/lua_temp.txt")
	`)
	if err != nil {
		log.Fatalf("Test 16 failed: %v", err)
	}

	// Test 17: Complex workflow
	fmt.Println("\n=== Test 17: Complex Workflow (Lua) ===")
	err = L.DoString(`
		local ufs = require("ufs")
		
		-- Create nested structure
		local ok, err = ufs.mkdir(1, "/home/app")
		if err then error("Failed to create /home/app: " .. err) end
		
		local ok, err = ufs.mkdir(1, "/home/app/data")
		if err then error("Failed to create /home/app/data: " .. err) end
		
		-- Write multiple files
		local ok, err = ufs.write(1, "/home/app/config.json", '{"name":"test"}')
		if err then error("Failed to write config: " .. err) end
		
		local ok, err = ufs.write(1, "/home/app/data/file1.txt", "Data 1")
		if err then error("Failed to write file1: " .. err) end
		
		local ok, err = ufs.write(1, "/home/app/data/file2.txt", "Data 2")
		if err then error("Failed to write file2: " .. err) end
		
		-- List all files
		local files, err = ufs.list("/home/app")
		if err then error("Failed to list /home/app: " .. err) end
		print(string.format("Files in /home/app: %d", #files))
		
		local files, err = ufs.list("/home/app/data")
		if err then error("Failed to list /home/app/data: " .. err) end
		print(string.format("Files in /home/app/data: %d", #files))
		
		-- Read and verify
		local content, err = ufs.read("/home/app/config.json")
		if err then error("Failed to read config: " .. err) end
		if not content:match("test") then error("Config content incorrect") end
		
		-- Clean up
		local ok, err = ufs.remove(1, "/home/app/config.json")
		if err then error("Failed to remove config: " .. err) end
		
		local ok, err = ufs.remove(1, "/home/app/data/file1.txt")
		if err then error("Failed to remove file1: " .. err) end
		
		local ok, err = ufs.remove(1, "/home/app/data/file2.txt")
		if err then error("Failed to remove file2: " .. err) end
		
		local ok, err = ufs.rmdir(1, "/home/app/data")
		if err then error("Failed to remove data dir: " .. err) end
		
		local ok, err = ufs.rmdir(1, "/home/app")
		if err then error("Failed to remove app dir: " .. err) end
		
		print("Complex workflow completed successfully")
	`)
	if err != nil {
		log.Fatalf("Test 17 failed: %v", err)
	}

	fmt.Println("\n=== All Lua UFS Binding Tests Passed! ===")
}
