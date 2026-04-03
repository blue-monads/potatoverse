package bllm

import (
	"os"
	"testing"

	luajit "github.com/yuin/gopher-lua"
)

func skipIfNoAPIKey(t *testing.T) {
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		t.Skip("Skipping test that requires OPENROUTER_API_KEY")
	}
}

func TestNewProvider(t *testing.T) {
	L := luajit.NewState()
	defer L.Close()

	Init(L)

	code := `
	local provider = llm.new({
		model = "google/gemini-3-flash-preview",
		apiKey = "test-key"
	})
	return provider ~= nil
	`

	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	if L.GetTop() != 1 || !L.ToBool(1) {
		t.Error("Expected provider to be created")
	}
}

func TestNewProviderDefaults(t *testing.T) {
	L := luajit.NewState()
	defer L.Close()

	Init(L)

	code := `
	local provider = llm.new({})
	return provider ~= nil
	`

	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	if L.GetTop() != 1 || !L.ToBool(1) {
		t.Error("Expected provider to be created with defaults")
	}
}

func TestNewProviderWithModel(t *testing.T) {
	L := luajit.NewState()
	defer L.Close()

	Init(L)

	code := `
	local provider = llm.new({
		model = "google/gemini-3-flash-preview",
		apiKey = "test-key"
	})
	return provider ~= nil
	`

	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	if L.GetTop() != 1 || !L.ToBool(1) {
		t.Error("Expected provider to be created with model")
	}
}

func TestGenerateV1Basic(t *testing.T) {
	skipIfNoAPIKey(t)

	L := luajit.NewState()
	defer L.Close()

	Init(L)

	code := `
	local provider = llm.new({
		apiKey = "` + os.Getenv("OPENROUTER_API_KEY") + `"
	})
	
	local result = provider:generate_v1({
		messages = {
			{ role = "user", content = "Hello" }
		}
	})
	return result ~= nil and result.content ~= nil
	`

	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
}

func TestGenerateV1WithMessages(t *testing.T) {
	skipIfNoAPIKey(t)

	L := luajit.NewState()
	defer L.Close()

	Init(L)

	code := `
	local provider = llm.new({
		apiKey = "` + os.Getenv("OPENROUTER_API_KEY") + `"
	})
	
	local result = provider:generate_v1({
		messages = {
			{ role = "system", content = "You are helpful" },
			{ role = "user", content = "Hello" },
			{ role = "assistant", content = "Hi there!" }
		}
	})
	return result ~= nil and result.content ~= nil
	`

	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
}

func TestGenerateV1WithTools(t *testing.T) {
	skipIfNoAPIKey(t)

	L := luajit.NewState()
	defer L.Close()

	Init(L)

	code := `
	local provider = llm.new({
		apiKey = "` + os.Getenv("OPENROUTER_API_KEY") + `"
	})
	
	local result = provider:generate_v1({
		messages = {{ role = "user", content = "What is the weather?" }},
		tools = {
			get_weather = {
				description = "Get the weather for a location",
				parameters = {
					type = "object",
					properties = {
						location = { type = "string" }
					},
					required = {"location"}
				}
			}
		}
	})
	return result ~= nil
	`

	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
}

func TestGenerateV1InvalidConfig(t *testing.T) {
	L := luajit.NewState()
	defer L.Close()

	Init(L)

	code := `
	local provider = llm.new({})
	
	local result = pcall(function()
		return provider:generate_v1()
	end)
	return result == false
	`

	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	if L.GetTop() != 1 || !L.ToBool(1) {
		t.Error("Expected pcall to return false for invalid config")
	}
}

func TestMessageRoles(t *testing.T) {
	skipIfNoAPIKey(t)

	testCases := []struct {
		role    string
		content string
	}{
		{"system", "You are a helpful assistant"},
		{"user", "Hello!"},
		{"assistant", "Hi there!"},
	}

	for _, tc := range testCases {
		L := luajit.NewState()
		Init(L)

		code := `
		local provider = llm.new({
			apiKey = "` + os.Getenv("OPENROUTER_API_KEY") + `"
		})
		
		local ok = pcall(function()
			provider:generate_v1({
				messages = {
					{ role = "` + tc.role + `", content = "` + tc.content + `" }
				}
			})
		end)
		return ok
		`

		if err := L.DoString(code); err != nil {
			t.Errorf("Role %s: DoString failed: %v", tc.role, err)
		}

		if L.GetTop() != 1 {
			t.Errorf("Role %s: expected 1 result", tc.role)
		}
		L.Close()
	}
}

func TestEmptyMessages(t *testing.T) {
	L := luajit.NewState()
	defer L.Close()

	Init(L)

	code := `
	local provider = llm.new({
		apiKey = "test-key"
	})
	
	local ok = pcall(function()
		provider:generate_v1({
			messages = {}
		})
	end)
	return ok
	`

	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
}

func TestNoTools(t *testing.T) {
	skipIfNoAPIKey(t)

	L := luajit.NewState()
	defer L.Close()

	Init(L)

	code := `
	local provider = llm.new({
		apiKey = "` + os.Getenv("OPENROUTER_API_KEY") + `"
	})
	
	local result = provider:generate_v1({
		messages = {{ role = "user", content = "Hello" }}
	})
	return result ~= nil
	`

	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
}

func TestLuaSyntaxMessages(t *testing.T) {
	L := luajit.NewState()
	defer L.Close()

	Init(L)

	code := `
	local provider = llm.new({
		apiKey = "test-key"
	})
	
	local messages = {
		{ role = "user", content = "Say 'test'" },
		{ role = "assistant", content = "test" }
	}
	
	return #messages == 2
	`

	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	if L.GetTop() != 1 || !L.ToBool(1) {
		t.Error("Expected messages array to have 2 elements")
	}
}

func TestLuaSyntaxTools(t *testing.T) {
	L := luajit.NewState()
	defer L.Close()

	Init(L)

	code := `
	local tools = {
		get_weather = {
			description = "Get weather",
			parameters = {
				type = "object",
				properties = {
					location = { type = "string" }
				}
			}
		}
	}
	
	return tools.get_weather ~= nil
	`

	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	if L.GetTop() != 1 || !L.ToBool(1) {
		t.Error("Expected tools table to be parsed correctly")
	}
}

func TestProviderType(t *testing.T) {
	L := luajit.NewState()
	defer L.Close()

	Init(L)

	code := `
	local provider = llm.new({})
	local mt = getmetatable(provider)
	return mt ~= nil and mt.__index ~= nil
	`

	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
}
