package luaz

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

type mockApp struct {
	xtypes.App
	db datahub.Database
}

func (m *mockApp) Database() datahub.Database { return m.db }
func (m *mockApp) Logger() *slog.Logger       { return slog.Default() }

type mockDB struct {
	datahub.Database
	pfops datahub.FileOps
}

func (m *mockDB) GetPackageFileOps() datahub.FileOps { return m.pfops }

type mockFileOps struct {
	datahub.FileOps
	files map[string]string
}

func (m *mockFileOps) GetFileContentByPath(ownerID int64, path, name string) ([]byte, error) {
	fullPath := name
	if path != "" {
		fullPath = path + "/" + name
	}
	if content, ok := m.files[fullPath]; ok {
		return []byte(content), nil
	}
	return nil, fmt.Errorf("file not found: %s", fullPath)
}

func TestRelativeLoader(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()

	files := map[string]string{
		"my_utils.lua":   "return {hello = 'world'}",
		"sub/helper.lua": "return {name = 'helper'}",
	}

	mfops := &mockFileOps{files: files}
	mdb := &mockDB{pfops: mfops}
	mapp := &mockApp{db: mdb}

	exBuilder := &LuazExecutorBuilder{app: mapp}
	ex := &LuazExecutor{parent: exBuilder, handle: &xtypes.ExecutorBuilderOption{PackageVersionId: 1}}
	lh := &LuaH{parent: ex, L: L}

	L.SetGlobal("_relative_loader", L.NewFunction(lh.relativeLoader))

	tests := []struct {
		name    string
		luaCode string
	}{
		{
			"simple relative require",
			`
				local loader = _relative_loader("./my_utils")
				if type(loader) ~= "function" then error("loader should be a function, got " .. type(loader)) end
				local mod = loader()
				assert(mod.hello == "world")
			`,
		},
		{
			"nested relative require",
			`
				local loader = _relative_loader("./sub.helper")
				if type(loader) ~= "function" then error("loader should be a function, got " .. type(loader)) end
				local mod = loader()
				assert(mod.name == "helper")
			`,
		},
		{
			"non-relative path returns error string",
			`
				local res = _relative_loader("standard_mod")
				assert(type(res) == "string")
				assert(res:find("no relative path prefix"))
			`,
		},
		{
			"file not found returns error string",
			`
				local res = _relative_loader("./non_existent")
				assert(type(res) == "string")
				assert(res:find("file not found in package"))
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.luaCode); err != nil {
				t.Errorf("Lua test failed: %v", err)
			}
		})
	}
}

func TestRequireIntegration(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()

	files := map[string]string{
		"my_utils.lua": "return {hello = 'world'}",
	}

	mfops := &mockFileOps{files: files}
	mdb := &mockDB{pfops: mfops}
	mapp := &mockApp{db: mdb}

	exBuilder := &LuazExecutorBuilder{app: mapp, binds: make(map[string]map[string]lua.LGFunction)}
	ex := &LuazExecutor{parent: exBuilder, handle: &xtypes.ExecutorBuilderOption{PackageVersionId: 1}}
	lh := &LuaH{parent: ex, L: L}

	// registerModules needs package table
	lh.registerModules()

	err := L.DoString(`
		local mod = require("./my_utils")
		assert(mod.hello == "world")
	`)
	if err != nil {
		t.Errorf("Require integration failed: %v", err)
	}
}
