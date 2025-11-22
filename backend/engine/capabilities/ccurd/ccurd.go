package ccurd

import (
	"fmt"
	"maps"
	"strings"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type CcurdCapability struct {
	spaceId int64
	db      datahub.DBLowOps
	signer  *signer.Signer
	methods map[string]*Methods
}

func (p *CcurdCapability) Reload(opts xtypes.LazyData) (xtypes.Capability, error) {

	methods, err := LoadMethods(opts)
	if err != nil {
		return nil, err
	}

	next := &CcurdCapability{
		spaceId: p.spaceId,
		db:      p.db,
		signer:  p.signer,
		methods: methods,
	}

	return next, nil
}

func (p *CcurdCapability) Close() error {
	return nil
}

func (p *CcurdCapability) Handle(ctx *gin.Context) {
	token := ctx.Request.Header.Get("x-cap-token")
	if token == "" {
		httpx.WriteErrString(ctx, "Empty token")
		return
	}

	claim, err := p.signer.ParseCapability(token)
	if err != nil {
		httpx.WriteErrString(ctx, "token error")
		return
	}

	if claim.SpaceId != p.spaceId {
		httpx.WriteErrString(ctx, "token `error")
	}

	subpath := ctx.Param("subpath")
	methodName := strings.Split(subpath, "/")[0]
	method := p.methods[methodName]

	if method == nil {
		httpx.WriteErrString(ctx, "method not found")
		return
	}

	switch method.Mode {
	case "insert":

		data := map[string]any{}
		if err := ctx.ShouldBindJSON(&data); err != nil {
			httpx.WriteErrString(ctx, fmt.Sprintf("bind data error: %s", err.Error()))
			return
		}

		if err := ValidateData(data, p.methods); err != nil {
			httpx.WriteErrString(ctx, fmt.Sprintf("validate data error: %s", err.Error()))
			return
		}

		if len(method.StaticFields) > 0 {
			maps.Copy(data, method.StaticFields)
		}

		id, err := p.db.Insert(method.Table, data)
		if err != nil {
			httpx.WriteErrString(ctx, fmt.Sprintf("insert data error: %s", err.Error()))
			return
		}

		httpx.WriteJSON(ctx, gin.H{"id": id}, nil)

	case "batch_insert":
		data := []map[string]any{}
		if err := ctx.ShouldBindJSON(&data); err != nil {
			httpx.WriteErrString(ctx, fmt.Sprintf("bind data error: %s", err.Error()))
			return
		}

		for _, item := range data {
			if err := ValidateData(item, p.methods); err != nil {
				httpx.WriteErrString(ctx, fmt.Sprintf("validate data error: %s", err.Error()))
				return
			}
		}

		ids := []int64{}
		for _, item := range data {
			id, err := p.db.Insert(method.Table, item)
			if err != nil {
				httpx.WriteErrString(ctx, fmt.Sprintf("batch insert data error: %s", err.Error()))
				return
			}
			ids = append(ids, id)
		}

		httpx.WriteJSON(ctx, gin.H{"ids": ids}, nil)
	case "select":
		cond := map[any]any{}
		if err := ctx.ShouldBindJSON(&cond); err != nil {
			httpx.WriteErrString(ctx, fmt.Sprintf("bind cond error: %s", err.Error()))
			return
		}

		if len(method.StaticFields) > 0 {
			for k, v := range method.StaticFields {
				cond[k] = v
			}
		}

		results, err := p.db.FindAllByCond(method.Table, cond)
		httpx.WriteJSON(ctx, results, err)
	default:
		httpx.WriteErrString(ctx, fmt.Sprintf("invalid method mode: %s", method.Mode))
		return
	}
}

func ValidateData(data map[string]any, methods map[string]*Methods) error {

	for vname, vvalue := range data {
		method := methods[vname]

		validator := method.Validators[vname]
		if validator == nil {
			return fmt.Errorf("validator %s not found", vname)
		}

		if validator.Required && vvalue == nil {
			return fmt.Errorf("field %s is required", vname)
		}

		if validator.Type == "string" {
			if len(vvalue.(string)) < int(validator.Min) {
				return fmt.Errorf("field %s is too short", vname)
			}

			if len(vvalue.(string)) > int(validator.Max) {
				return fmt.Errorf("field %s is too long", vname)
			}

			if !validator.compiledRegex.MatchString(vvalue.(string)) {
				return fmt.Errorf("field %s is invalid", vname)
			}
		}

		if validator.Type == "number" {

			if validator.Min != 0 {
				if vvalue.(float64) < float64(validator.Min) {
					return fmt.Errorf("field %s is too small", vname)
				}
			}

			if validator.Max != 0 {
				if vvalue.(float64) > float64(validator.Max) {
					return fmt.Errorf("field %s is too large", vname)
				}
			}

		}
	}

	return nil
}

func (p *CcurdCapability) ListActions() ([]string, error) {
	return []string{"loaded_methods"}, nil
}

func (p *CcurdCapability) Execute(name string, params xtypes.LazyData) (map[string]any, error) {

	switch name {
	case "loaded_methods":
		return map[string]any{
			"methods": p.methods,
		}, nil

	default:
		return nil, fmt.Errorf("invalid action: %s", name)
	}
}
