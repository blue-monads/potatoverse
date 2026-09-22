package xtypes

import "fmt"

type UNIXRpcRequest struct {
	Method string         `json:"method"`
	Args   map[string]any `json:"args,omitempty"`
}

func (r *UNIXRpcRequest) GetStringArg(keys ...string) string {
	if r.Args == nil {
		return ""
	}
	for _, k := range keys {
		if v, ok := r.Args[k]; ok && v != nil {
			return fmt.Sprint(v)
		}
	}
	return ""
}

type UNIXRpcResponse struct {
	Ok   bool           `json:"ok"`
	Msg  string         `json:"msg,omitempty"`
	Data map[string]any `json:"data,omitempty"`
}
