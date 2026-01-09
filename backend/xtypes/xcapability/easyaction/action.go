package easyaction

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/tidwall/gjson"
)

var Methods = []string{
	"as_bytes",
	"as_map",
	"as_json_bytes",
	"as_json_value",
	"get_value",
	"get_field_as_int",
	"get_field_as_float",
	"get_field_as_string",
	"get_field_as_bool",
}

type Context struct {
	Capability xcapability.Capability
	Payload    []byte
	Handler    func(name string, params lazydata.LazyData) (any, error)
}

func (c *Context) ListActions() ([]string, error) {
	if c.Capability == nil {
		return Methods, nil
	}

	actions, err := c.Capability.ListActions()
	if err != nil {
		return nil, err
	}

	actions = append(actions, Methods...)
	return actions, nil

}

func (c *Context) ExecuteAction(name string, params lazydata.LazyData) (any, error) {

	if c.Payload == nil {
		if c.Handler != nil {
			return c.Handler(name, params)
		}

		return c.Capability.Execute(name, params)
	}

	resp, err := BytelazyDataActions(c.Payload, name, params)
	if err != nil {
		if errors.Is(err, ErrUnknownAction) {
			return c.Capability.Execute(name, params)
		}
		return nil, err
	}

	return resp, nil

}

var ErrUnknownAction = errors.New("unknown action")

func BytelazyDataActions(data []byte, name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "as_bytes":
		return data, nil
	case "as_map":
		ld := lazydata.LazyDataBytes(data)
		return ld.AsMap()
	case "as_json_bytes":
		ld := lazydata.LazyDataBytes(data)
		return ld.AsBytes()
	case "as_json_value":
		var target any
		err := json.Unmarshal(data, &target)
		if err != nil {
			return nil, err
		}
		return target, nil

	case "get_value":
		path := params.GetFieldAsString("path")
		finalPath := fmt.Sprintf("data.%s", path)
		return gjson.GetBytes(data, finalPath).Value(), nil

	case "get_field_as_int":
		path := params.GetFieldAsString("path")
		finalPath := fmt.Sprintf("data.%s", path)
		return gjson.GetBytes(data, finalPath).Int(), nil
	case "get_field_as_float":
		path := params.GetFieldAsString("path")
		finalPath := fmt.Sprintf("data.%s", path)
		return gjson.GetBytes(data, finalPath).Float(), nil

	case "get_field_as_string":

		path := params.GetFieldAsString("path")
		finalPath := fmt.Sprintf("data.%s", path)

		return gjson.GetBytes(data, finalPath).String(), nil
	case "get_field_as_bool":
		path := params.GetFieldAsString("path")
		finalPath := fmt.Sprintf("data.%s", path)
		return gjson.GetBytes(data, finalPath).Bool(), nil
	default:
		return nil, ErrUnknownAction
	}

}
