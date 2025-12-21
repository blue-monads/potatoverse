package easyaction

import (
	"github.com/blue-monads/turnix/backend/xtypes/lazydata"
	"github.com/blue-monads/turnix/backend/xtypes/xcapability"
)

var Methods = []string{
	"as_bytes",
	"as_map",
	"as_json",
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

	switch name {
	case "as_bytes":
		ld := lazydata.LazyDataBytes(c.Payload)
		return ld.AsBytes()
	case "as_map":
		ld := lazydata.LazyDataBytes(c.Payload)
		return ld.AsMap()
	case "as_json":
		ld := lazydata.LazyDataBytes(c.Payload)
		return ld.AsBytes()

	case "get_field_as_int":
		return params.GetFieldAsInt(params.GetFieldAsString("path")), nil
	case "get_field_as_float":
		return params.GetFieldAsFloat(params.GetFieldAsString("path")), nil
	case "get_field_as_string":
		return params.GetFieldAsString(params.GetFieldAsString("path")), nil
	case "get_field_as_bool":
		return params.GetFieldAsBool(params.GetFieldAsString("path")), nil

	default:

		if c.Handler != nil {
			return c.Handler(name, params)
		}

		return c.Capability.Execute(name, params)
	}

}
