package xlua

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/luaplus"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/cjoudrey/gluahttp"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
	luaJson "layeh.com/gopher-json"
)

var (
	Name         = "xLua"
	Icon         = `<i class="fa-solid fa-scroll"></i>`
	OptionFields = []xcapability.CapabilityOptionField{}
)

func init() {
	registry.RegisterCapability(xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			return &LuaBuilder{app: appTyped}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type LuaBuilder struct {
	app       xtypes.App
	vms       sync.Map
	vmCounter atomic.Int64
}

func (b *LuaBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	model := handle.GetModel()
	return &LuaCapability{
		builder: b,
		app:     b.app,
		spaceId: model.SpaceID,
	}, nil
}

func (b *LuaBuilder) Serve(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message":    "xLua capability",
		"capability": Name,
	})
}

func (b *LuaBuilder) Name() string {
	return Name
}

func (b *LuaBuilder) GetDebugData() map[string]any {
	return map[string]any{}
}

var luaHttpClient = &http.Client{}

type managedVM struct {
	L  *lua.LState
	mu sync.Mutex
}

type LuaCapability struct {
	builder *LuaBuilder
	app     xtypes.App
	spaceId int64
}

func (c *LuaCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	return c, nil
}

func (c *LuaCapability) Close() error {
	c.builder.vms.Range(func(key, value any) bool {
		if vm, ok := value.(*managedVM); ok {
			vm.mu.Lock()
			vm.L.Close()
			vm.mu.Unlock()
		}
		return true
	})
	return nil
}

func (c *LuaCapability) Handle(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message":    "xLua capability",
		"capability": Name,
		"space_id":   c.spaceId,
	})
}

func (c *LuaCapability) ListActions() ([]string, error) {
	return []string{"run_oneoff_script", "build_vm", "execute_vm", "destroy_vm", "list_vms"}, nil
}

func (c *LuaCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "run_oneoff_script":
		return c.runOneoffScript(params)
	case "build_vm":
		return c.buildVM(params)
	case "execute_vm":
		return c.executeVM(params)
	case "destroy_vm":
		return c.destroyVM(params)
	case "list_vms":
		return c.listVMs()
	default:
		return nil, errors.New("unknown action: " + name)
	}
}

func (c *LuaCapability) newLuaState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule("phttp", gluahttp.NewHttpModule(luaHttpClient).Loader)
	L.PreloadModule("json", luaJson.Loader)
	return L
}

func (c *LuaCapability) runOneoffScript(params lazydata.LazyData) (any, error) {
	script := params.GetFieldAsString("script")
	if script == "" {
		return nil, errors.New("script is required")
	}

	L := c.newLuaState()
	defer L.Close()

	err := L.DoString(script)
	if err != nil {
		return nil, fmt.Errorf("lua error: %w", err)
	}

	ret := L.Get(-1)
	if ret == lua.LNil {
		return nil, nil
	}

	return luaplus.LuaTypeToGoType(L, ret), nil
}

func (c *LuaCapability) buildVM(params lazydata.LazyData) (any, error) {
	script := params.GetFieldAsString("script")
	if script == "" {
		return nil, errors.New("script is required")
	}

	vmId := params.GetFieldAsString("vm_id")
	if vmId == "" {
		vmId = fmt.Sprintf("vm_%d", c.builder.vmCounter.Add(1))
	}

	if _, loaded := c.builder.vms.Load(vmId); loaded {
		return nil, fmt.Errorf("vm already exists: %s", vmId)
	}

	L := c.newLuaState()

	err := L.DoString(script)
	if err != nil {
		L.Close()
		return nil, fmt.Errorf("lua error: %w", err)
	}

	c.builder.vms.Store(vmId, &managedVM{L: L})

	return map[string]any{
		"vm_id": vmId,
	}, nil
}

func (c *LuaCapability) executeVM(params lazydata.LazyData) (any, error) {
	vmId := params.GetFieldAsString("vm_id")
	if vmId == "" {
		return nil, errors.New("vm_id is required")
	}

	method := params.GetFieldAsString("method")
	if method == "" {
		return nil, errors.New("method is required")
	}

	raw, ok := c.builder.vms.Load(vmId)
	if !ok {
		return nil, fmt.Errorf("vm not found: %s", vmId)
	}

	vm := raw.(*managedVM)
	vm.mu.Lock()
	defer vm.mu.Unlock()

	fn := vm.L.GetGlobal(method)
	if fn == lua.LNil {
		return nil, fmt.Errorf("method not found: %s", method)
	}

	callParams, err := params.AsMap()
	if err != nil {
		return nil, err
	}

	var argTable *lua.LTable
	if inner, ok := callParams["params"]; ok {
		if m, ok := inner.(map[string]any); ok {
			argTable = luaplus.MapToTable(vm.L, m)
		}
	}

	if argTable == nil {
		argTable = vm.L.NewTable()
	}

	err = vm.L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}, argTable)
	if err != nil {
		return nil, fmt.Errorf("lua error: %w", err)
	}

	ret := vm.L.Get(-1)
	vm.L.Pop(1)

	if ret == lua.LNil {
		return nil, nil
	}

	return luaplus.LuaTypeToGoType(vm.L, ret), nil
}

func (c *LuaCapability) destroyVM(params lazydata.LazyData) (any, error) {
	vmId := params.GetFieldAsString("vm_id")
	if vmId == "" {
		return nil, errors.New("vm_id is required")
	}

	raw, loaded := c.builder.vms.LoadAndDelete(vmId)
	if !loaded {
		return nil, fmt.Errorf("vm not found: %s", vmId)
	}

	vm := raw.(*managedVM)
	vm.mu.Lock()
	vm.L.Close()
	vm.mu.Unlock()

	return map[string]any{
		"vm_id":   vmId,
		"deleted": true,
	}, nil
}

func (c *LuaCapability) listVMs() (any, error) {
	ids := make([]string, 0)
	c.builder.vms.Range(func(key, _ any) bool {
		ids = append(ids, key.(string))
		return true
	})

	return map[string]any{
		"vm_ids": ids,
	}, nil
}
