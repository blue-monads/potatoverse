package binds

import (
	"fmt"
)

func (s *BindingServer) handleKV(method string, params []any) (any, error) {
	kv := s.app.Database().GetSpaceKVOps()
	switch method {
	case "get":
		if len(params) < 2 {
			return nil, fmt.Errorf("missing group or key")
		}
		group := params[0].(string)
		key := params[1].(string)
		return kv.GetSpaceKV(s.installId, group, key)
	case "set":
		if len(params) < 3 {
			return nil, fmt.Errorf("missing group, key or value")
		}
		group := params[0].(string)
		key := params[1].(string)
		value := params[2]
		return nil, kv.UpsertSpaceKV(s.installId, group, key, map[string]any{"value": value})
	case "remove":
		if len(params) < 2 {
			return nil, fmt.Errorf("missing group or key")
		}
		group := params[0].(string)
		key := params[1].(string)
		return nil, kv.RemoveSpaceKV(s.installId, group, key)
	}
	return nil, fmt.Errorf("unknown kv method: %s", method)
}
