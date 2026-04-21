package rtbinds

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
)

func toMapAnyAny(m map[string]any) map[any]any {
	res := make(map[any]any, len(m))
	for k, v := range m {
		res[k] = v
	}
	return res
}

// DB Operations

func (b *BindServer) DBRunQuery(ctx *HttpBindContext) (any, error) {
	var req struct {
		Query string `json:"query"`
		Args  []any  `json:"args"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.RunQuery(req.Query, req.Args...)
}

func (b *BindServer) DBRunQueryOne(ctx *HttpBindContext) (any, error) {
	var req struct {
		Query string `json:"query"`
		Args  []any  `json:"args"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.RunQueryOne(req.Query, req.Args...)
}

func (b *BindServer) DBInsert(ctx *HttpBindContext) (any, error) {
	var req struct {
		Table string         `json:"table"`
		Data  map[string]any `json:"data"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.Insert(req.Table, req.Data)
}

func (b *BindServer) DBUpdateById(ctx *HttpBindContext) (any, error) {
	var req struct {
		Table string         `json:"table"`
		ID    int64          `json:"id"`
		Data  map[string]any `json:"data"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	err := dbOps.UpdateById(req.Table, req.ID, req.Data)
	return nil, err
}

func (b *BindServer) DBDeleteById(ctx *HttpBindContext) (any, error) {
	var req struct {
		Table string `json:"table"`
		ID    int64  `json:"id"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	err := dbOps.DeleteById(req.Table, req.ID)
	return nil, err
}

func (b *BindServer) DBFindById(ctx *HttpBindContext) (any, error) {
	var req struct {
		Table string `json:"table"`
		ID    int64  `json:"id"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.FindById(req.Table, req.ID)
}

func (b *BindServer) DBUpdateByCond(ctx *HttpBindContext) (any, error) {
	var req struct {
		Table string         `json:"table"`
		Cond  map[string]any `json:"cond"`
		Data  map[string]any `json:"data"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	err := dbOps.UpdateByCond(req.Table, toMapAnyAny(req.Cond), req.Data)
	return nil, err
}

func (b *BindServer) DBDeleteByCond(ctx *HttpBindContext) (any, error) {
	var req struct {
		Table string         `json:"table"`
		Cond  map[string]any `json:"cond"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	err := dbOps.DeleteByCond(req.Table, toMapAnyAny(req.Cond))
	return nil, err
}

func (b *BindServer) DBFindAllByCond(ctx *HttpBindContext) (any, error) {
	var req struct {
		Table string         `json:"table"`
		Cond  map[string]any `json:"cond"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.FindAllByCond(req.Table, toMapAnyAny(req.Cond))
}

func (b *BindServer) DBFindOneByCond(ctx *HttpBindContext) (any, error) {
	var req struct {
		Table string         `json:"table"`
		Cond  map[string]any `json:"cond"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.FindOneByCond(req.Table, toMapAnyAny(req.Cond))
}

func (b *BindServer) DBFindAllByQuery(ctx *HttpBindContext) (any, error) {
	req := &datahub.FindQuery{}
	if err := ctx.Http.BindJSON(req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.FindAllByQuery(req)
}

func (b *BindServer) DBFindByJoin(ctx *HttpBindContext) (any, error) {
	req := &datahub.FindByJoin{}
	if err := ctx.Http.BindJSON(req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.FindByJoin(req)
}

func (b *BindServer) DBListTables(ctx *HttpBindContext) (any, error) {
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.ListTables()
}

func (b *BindServer) DBListColumns(ctx *HttpBindContext) (any, error) {
	tableName := ctx.Http.Param("table")
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.ListTableColumns(tableName)
}

// KV Operations

func (b *BindServer) KVAdd(ctx *HttpBindContext) (any, error) {
	req := &dbmodels.SpaceKV{}
	if err := ctx.Http.BindJSON(req); err != nil {
		return nil, err
	}
	kvOps := b.db.GetSpaceKVOps()
	err := kvOps.AddSpaceKV(ctx.PackageId, req)
	return req, err
}

func (b *BindServer) KVGet(ctx *HttpBindContext) (any, error) {
	group := ctx.Http.Param("group")
	key := ctx.Http.Param("key")
	kvOps := b.db.GetSpaceKVOps()
	return kvOps.GetSpaceKV(ctx.PackageId, group, key)
}

func (b *BindServer) KVQuery(ctx *HttpBindContext) (any, error) {
	var req struct {
		Cond         map[string]any `json:"cond"`
		Offset       int            `json:"offset"`
		Limit        int            `json:"limit"`
		IncludeValue bool           `json:"include_value"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	kvOps := b.db.GetSpaceKVOps()
	if req.IncludeValue {
		return kvOps.QueryWithValueSpaceKV(ctx.PackageId, toMapAnyAny(req.Cond), req.Offset, req.Limit)
	}
	return kvOps.QuerySpaceKV(ctx.PackageId, toMapAnyAny(req.Cond), req.Offset, req.Limit)
}

func (b *BindServer) KVRemove(ctx *HttpBindContext) (any, error) {
	var req struct {
		Group string `json:"group"`
		Key   string `json:"key"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	kvOps := b.db.GetSpaceKVOps()
	err := kvOps.RemoveSpaceKV(ctx.PackageId, req.Group, req.Key)
	return nil, err
}

func (b *BindServer) KVUpdate(ctx *HttpBindContext) (any, error) {
	var req struct {
		Group string         `json:"group"`
		Key   string         `json:"key"`
		Data  map[string]any `json:"data"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	kvOps := b.db.GetSpaceKVOps()
	err := kvOps.UpdateSpaceKV(ctx.PackageId, req.Group, req.Key, req.Data)
	return nil, err
}

func (b *BindServer) KVUpsert(ctx *HttpBindContext) (any, error) {
	var req struct {
		Group string         `json:"group"`
		Key   string         `json:"key"`
		Data  map[string]any `json:"data"`
	}
	if err := ctx.Http.BindJSON(&req); err != nil {
		return nil, err
	}
	kvOps := b.db.GetSpaceKVOps()
	err := kvOps.UpsertSpaceKV(ctx.PackageId, req.Group, req.Key, req.Data)
	return nil, err
}
