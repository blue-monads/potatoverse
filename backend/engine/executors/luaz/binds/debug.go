package binds

import (
	"fmt"

	"github.com/blue-monads/turnix/backend/utils/luaplus"
	lua "github.com/yuin/gopher-lua"
)

// AutoDump iterates through all globals and registered modules, listing their contents
// Returns a map with:
//   - globals: all global variables and their types
//   - modules: all loaded/preloaded modules with their methods/functions
func AutoDump(L *lua.LState) (map[string]any, error) {
	result := make(map[string]any)
	globalsMap := make(map[string]any)
	modulesMap := make(map[string]any)

	// Helper to convert Lua value to Go type with type info
	convertValue := func(value lua.LValue) map[string]any {
		info := make(map[string]any)
		info["type"] = value.Type().String()

		switch v := value.(type) {
		case *lua.LTable:
			// For tables, use LuaTypeToGoType which handles arrays vs maps
			info["value"] = luaplus.LuaTypeToGoType(L, v)
		case lua.LString:
			info["value"] = string(v)
		case lua.LNumber:
			// Check if it's an integer
			if v == lua.LNumber(int64(v)) {
				info["value"] = int64(v)
			} else {
				info["value"] = float64(v)
			}
		case lua.LBool:
			info["value"] = bool(v)
		case *lua.LFunction:
			info["value"] = fmt.Sprintf("function: %p", v)
		default:
			info["value"] = v.String()
		}

		return info
	}

	// Get the global table (_G)
	globalTable := L.GetGlobal("_G")
	if globalTable != lua.LNil {
		if gt, ok := globalTable.(*lua.LTable); ok {
			gt.ForEach(func(key, value lua.LValue) {
				keyStr := key.String()

				// Skip internal Lua globals
				if keyStr == "_G" || keyStr == "_VERSION" || keyStr == "package" {
					return
				}

				typeInfo := convertValue(value)

				// If it's a table, inspect it for functions/methods
				if tbl, ok := value.(*lua.LTable); ok {
					methods := make([]map[string]any, 0)
					tbl.ForEach(func(mKey, mValue lua.LValue) {
						if mValue.Type() == lua.LTFunction {
							methodInfo := map[string]any{
								"type": "function",
								"name": mKey.String(),
							}
							methods = append(methods, methodInfo)
						}
					})
					typeInfo["methods"] = methods
				}

				globalsMap[keyStr] = typeInfo
			})
		}
	}

	// Get preloaded modules from package.loaded
	packageLoaded := L.GetGlobal("package")
	if packageLoaded != lua.LNil {
		if pkg, ok := packageLoaded.(*lua.LTable); ok {
			loaded := L.GetField(pkg, "loaded")
			if loaded != lua.LNil {
				if loadedTbl, ok := loaded.(*lua.LTable); ok {
					loadedTbl.ForEach(func(key, value lua.LValue) {
						moduleName := key.String()

						moduleInfo := make(map[string]any)
						moduleInfo["name"] = moduleName
						moduleInfo["type"] = value.Type().String()

						// If module is a table, inspect its contents
						if modTbl, ok := value.(*lua.LTable); ok {
							methods := make([]map[string]any, 0)
							fields := make([]map[string]any, 0)

							modTbl.ForEach(func(mKey, mValue lua.LValue) {
								fieldInfo := make(map[string]any)
								fieldInfo["name"] = mKey.String()
								fieldInfo["type"] = mValue.Type().String()

								if mValue.Type() == lua.LTFunction {
									fieldInfo["is_function"] = true
									methods = append(methods, fieldInfo)
								} else {
									fieldInfo["value"] = convertValue(mValue)["value"]
									fields = append(fields, fieldInfo)
								}
							})

							moduleInfo["methods"] = methods
							moduleInfo["fields"] = fields
						} else {
							moduleInfo["value"] = convertValue(value)["value"]
						}

						modulesMap[moduleName] = moduleInfo
					})
				}
			}
		}
	}

	// Also try to load known preloaded modules if they're not in package.loaded
	knownModules := []string{"kv", "mcp", "capability"}
	for _, modName := range knownModules {
		// Check if already in modules map
		if _, exists := modulesMap[modName]; !exists {
			// Try to require it
			L.Push(L.GetGlobal("require"))
			L.Push(lua.LString(modName))
			err := L.PCall(1, 1, nil)
			if err == nil {
				modValue := L.Get(-1)
				L.Pop(1)

				if modValue != lua.LNil {
					moduleInfo := make(map[string]any)
					moduleInfo["name"] = modName
					moduleInfo["type"] = modValue.Type().String()

					if modTbl, ok := modValue.(*lua.LTable); ok {
						methods := make([]map[string]any, 0)
						fields := make([]map[string]any, 0)

						modTbl.ForEach(func(mKey, mValue lua.LValue) {
							fieldInfo := make(map[string]any)
							fieldInfo["name"] = mKey.String()
							fieldInfo["type"] = mValue.Type().String()

							if mValue.Type() == lua.LTFunction {
								fieldInfo["is_function"] = true
								methods = append(methods, fieldInfo)
							} else {
								fieldInfo["value"] = convertValue(mValue)["value"]
								fields = append(fields, fieldInfo)
							}
						})

						moduleInfo["methods"] = methods
						moduleInfo["fields"] = fields
					} else {
						moduleInfo["value"] = convertValue(modValue)["value"]
					}

					modulesMap[modName] = moduleInfo
				}
			}
		}
	}

	// If capability module is loaded, call list() and methods() for each capability
	// Try to require the capability module (it should be preloaded)
	L.Push(L.GetGlobal("require"))
	L.Push(lua.LString("capability"))
	err := L.PCall(1, 1, nil)
	if err == nil {
		capMod := L.Get(-1)
		L.Pop(1)

		if capModTbl, ok := capMod.(*lua.LTable); ok {
			// Get the list function
			listFunc := L.GetField(capModTbl, "list")
			if listFunc != lua.LNil && listFunc.Type() == lua.LTFunction {
				// Call list()
				L.Push(listFunc)
				err := L.PCall(0, 1, nil)
				if err == nil {
					capListResult := L.Get(-1)
					L.Pop(1)

					// Convert the result (should be a table of capability names)
					if capListTbl, ok := capListResult.(*lua.LTable); ok {
						capabilities := make([]map[string]any, 0)

						// Get the methods function
						methodsFunc := L.GetField(capModTbl, "methods")
						if methodsFunc != lua.LNil && methodsFunc.Type() == lua.LTFunction {
							// Iterate through each capability
							capListTbl.ForEach(func(_, capNameLVal lua.LValue) {
								capName := capNameLVal.String()

								// Call methods(capabilityName)
								L.Push(methodsFunc)
								L.Push(capNameLVal)
								err := L.PCall(1, 1, nil)
								if err == nil {
									methodsResult := L.Get(-1)
									L.Pop(1)

									capInfo := make(map[string]any)
									capInfo["name"] = capName

									// Convert methods result to Go type
									if methodsTbl, ok := methodsResult.(*lua.LTable); ok {
										methodsList := luaplus.LuaTypeToGoType(L, methodsTbl)
										capInfo["methods"] = methodsList
									} else {
										capInfo["methods"] = convertValue(methodsResult)["value"]
									}

									capabilities = append(capabilities, capInfo)
								}
							})
						}

						// Add capabilities info to the module info in modulesMap
						if capModInfo, exists := modulesMap["capability"]; exists {
							if capModInfoMap, ok := capModInfo.(map[string]any); ok {
								capModInfoMap["capabilities"] = capabilities
							}
						} else {
							// If capability module wasn't in modulesMap, add it now
							moduleInfo := make(map[string]any)
							moduleInfo["name"] = "capability"
							moduleInfo["type"] = capMod.Type().String()
							moduleInfo["capabilities"] = capabilities
							modulesMap["capability"] = moduleInfo
						}
					}
				}
			}
		}
	}

	result["globals"] = globalsMap
	result["modules"] = modulesMap

	return result, nil
}
