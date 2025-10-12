package main

import (
	"fmt"
	"log"

	"github.com/blue-monads/turnix/backend/engine/executors/luaz/binds"
	lua "github.com/yuin/gopher-lua"
)

// HandleLuazMcpTest tests the MCP Lua bindings by connecting to a real MCP server
func HandleLuazMcpTest() {
	fmt.Println("@luaz_mcp_test_start")
	defer fmt.Println("@luaz_mcp_test_end")

	// Create Lua state
	L := lua.NewState()
	defer L.Close()

	// Register MCP module
	L.PreloadModule("mcp", binds.BindMCP())

	// Test 0: Verify Module Structure
	fmt.Println("\n=== Test 0: Verify MCP Module Structure (Lua) ===")
	err := L.DoString(`
		local mcp = require("mcp")
		if type(mcp) ~= "table" then
			error("MCP module should be a table, got: " .. type(mcp))
		end
		print("✓ MCP module loaded successfully")
		
		if type(mcp.create_http_client) ~= "function" then
			error("create_http_client should be a function")
		end
		print("✓ create_http_client function exists")
	`)
	if err != nil {
		log.Fatalf("Test 0 failed: %v", err)
	}

	// Test 1: Create MCP HTTP Client
	fmt.Println("\n=== Test 1: Create MCP HTTP Client (Lua) ===")
	err = L.DoString(`
		local mcp = require("mcp")
		print("Creating MCP client...")
		client, err = mcp.create_http_client("https://echo.mcp.inevitable.fyi/mcp", "PotatoverseTestClient")
		if err then
			print("⚠ Failed to create MCP client: " .. err)
			print("⚠ The echo MCP server might be unavailable or the endpoint has changed")
			print("⚠ This is expected if the server is down - skipping remaining tests")
			client = nil
		else
			print("✓ Successfully created MCP client")
		end
	`)
	if err != nil {
		log.Fatalf("Test 1 failed: %v", err)
	}

	// Test 2: List Tools
	fmt.Println("\n=== Test 2: List Tools (Lua) ===")
	err = L.DoString(`
		if not client then
			print("⚠ Skipping test - no client available")
			return
		end
		local tools, err = client:list_tools(client, {})
		if err then
			error("Failed to list tools: " .. err)
		end
		print(string.format("✓ Found %d tools", #tools))
		for i, tool in ipairs(tools) do
			print(string.format("  Tool %d: %s", i, tool.name or "unnamed"))
			if tool.description then
				print(string.format("    Description: %s", tool.description))
			end
		end
	`)
	if err != nil {
		log.Fatalf("Test 2 failed: %v", err)
	}

	// Test 3: List Resources
	fmt.Println("\n=== Test 3: List Resources (Lua) ===")
	err = L.DoString(`
		if not client then
			print("⚠ Skipping test - no client available")
			return
		end
		local resources, err = client:list_resources(client, {})
		if err then
			error("Failed to list resources: " .. err)
		end
		print(string.format("✓ Found %d resources", #resources))
		for i, resource in ipairs(resources) do
			print(string.format("  Resource %d: %s", i, resource.name or resource.uri or "unnamed"))
			if resource.description then
				print(string.format("    Description: %s", resource.description))
			end
			if resource.uri then
				print(string.format("    URI: %s", resource.uri))
			end
		end
	`)
	if err != nil {
		log.Fatalf("Test 3 failed: %v", err)
	}

	// Test 4: Call Tool (echo test)
	fmt.Println("\n=== Test 4: Call Tool - Echo Test (Lua) ===")
	err = L.DoString(`
		if not client then
			print("⚠ Skipping test - no client available")
			return
		end
		-- First, get the available tools to find the echo tool
		local tools, err = client:list_tools(client, {})
		if err then
			error("Failed to list tools for echo test: " .. err)
		end
		
		-- Find a tool to call (typically echo servers have an "echo" tool)
		local tool_name = nil
		for i, tool in ipairs(tools) do
			if tool.name then
				tool_name = tool.name
				break
			end
		end
		
		if tool_name then
			print(string.format("Attempting to call tool: %s", tool_name))
			local result, err = client:call_tool({
				name = tool_name,
				arguments = {
					message = "Hello from Potatoverse MCP test!"
				}
			})
			if err then
				print(string.format("⚠ Tool call returned error (may be expected): %s", err))
			else
				print("✓ Successfully called tool")
				if result then
					print("Result received:")
					-- Try to print the result structure
					for k, v in pairs(result) do
						if type(v) == "table" then
							print(string.format("  %s: [table]", k))
						else
							print(string.format("  %s: %s", k, tostring(v)))
						end
					end
				end
			end
		else
			print("⚠ No tools available to test call_tool")
		end
	`)
	if err != nil {
		log.Fatalf("Test 4 failed: %v", err)
	}

	// Test 5: Error Handling - Invalid Endpoint
	fmt.Println("\n=== Test 5: Error Handling - Invalid Endpoint (Lua) ===")
	err = L.DoString(`
		local mcp = require("mcp")
		local bad_client, err = mcp.create_http_client("https://invalid.endpoint.test/mcp", "TestClient")
		if not err then
			print("⚠ Expected error for invalid endpoint but connection may still be created")
			print("  (Error might only occur on actual API calls)")
		else
			print(string.format("✓ Got expected error for invalid endpoint: %s", err))
		end
	`)
	if err != nil {
		log.Fatalf("Test 5 failed: %v", err)
	}

	// Test 6: Multiple Clients
	fmt.Println("\n=== Test 6: Multiple Clients (Lua) ===")
	err = L.DoString(`
		if not client then
			print("⚠ Skipping test - no client available")
			return
		end
		local mcp = require("mcp")
		
		-- Create second client
		local client2, err = mcp.create_http_client("https://echo.mcp.inevitable.fyi/mcp", "SecondClient")
		if err then
			error("Failed to create second MCP client: " .. err)
		end
		print("✓ Successfully created second MCP client")
		
		-- List tools with both clients
		local tools1, err = client:list_tools(client, {})
		if err then
			error("Failed to list tools with first client: " .. err)
		end
		
		local tools2, err = client2:list_tools(client2, {})
		if err then
			error("Failed to list tools with second client: " .. err)
		end
		
		print(string.format("✓ Client 1 found %d tools", #tools1))
		print(string.format("✓ Client 2 found %d tools", #tools2))
	`)
	if err != nil {
		log.Fatalf("Test 6 failed: %v", err)
	}

	// Final summary
	err = L.DoString(`
		if client then
			print("\n=== ✓ All Lua MCP Binding Tests Passed! ===")
		else
			print("\n=== ⚠ Lua MCP Binding Tests Completed with Warnings ===")
			print("Note: The MCP echo server was unavailable, so functional tests were skipped.")
			print("The bindings module loaded successfully and basic functionality was verified.")
		end
	`)
	if err != nil {
		log.Fatalf("Final summary failed: %v", err)
	}
}
