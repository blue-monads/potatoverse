package remotehub

import (
	"encoding/json"
)

type RContext interface {
	GetSpaceId() int64
	GetPackageId() int64
	GetPackageVersion() int64

	GetData() ([]byte, error)
	SetData([]byte) error

	GetMeta(string) (string, error)
}

func bindJSON(ctx RContext, target any) error {
	data, err := ctx.GetData()
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, target)
}

func setDataJSON(ctx RContext, val any) {
	if val == nil {
		return
	}
	if b, ok := val.([]byte); ok {
		_ = ctx.SetData(b)
		return
	}
	if s, ok := val.(string); ok {
		_ = ctx.SetData([]byte(s))
		return
	}
	if out, err := json.Marshal(val); err == nil {
		_ = ctx.SetData(out)
	}
}
