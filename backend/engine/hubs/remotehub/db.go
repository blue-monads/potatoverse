package remotehub

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

func (b *RemoteHub) DBRunQuery(ctx *HttpBindContext) (any, error) {
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

func (b *RemoteHub) DBRunQueryOne(ctx *HttpBindContext) (any, error) {
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

func (b *RemoteHub) DBInsert(ctx *HttpBindContext) (any, error) {
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

func (b *RemoteHub) DBUpdateById(ctx *HttpBindContext) (any, error) {
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

func (b *RemoteHub) DBDeleteById(ctx *HttpBindContext) (any, error) {
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

func (b *RemoteHub) DBFindById(ctx *HttpBindContext) (any, error) {
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

func (b *RemoteHub) DBUpdateByCond(ctx *HttpBindContext) (any, error) {
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

func (b *RemoteHub) DBDeleteByCond(ctx *HttpBindContext) (any, error) {
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

func (b *RemoteHub) DBFindAllByCond(ctx *HttpBindContext) (any, error) {
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

func (b *RemoteHub) DBFindOneByCond(ctx *HttpBindContext) (any, error) {
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

func (b *RemoteHub) DBFindAllByQuery(ctx *HttpBindContext) (any, error) {
	req := &datahub.FindQuery{}
	if err := ctx.Http.BindJSON(req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.FindAllByQuery(req)
}

func (b *RemoteHub) DBFindByJoin(ctx *HttpBindContext) (any, error) {
	req := &datahub.FindByJoin{}
	if err := ctx.Http.BindJSON(req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.FindByJoin(req)
}

func (b *RemoteHub) DBListTables(ctx *HttpBindContext) (any, error) {
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.ListTables()
}

func (b *RemoteHub) DBListColumns(ctx *HttpBindContext) (any, error) {
	tableName := ctx.Http.Param("table")
	dbOps := b.db.GetLowPackageDBOps(ctx.PackageId)
	return dbOps.ListTableColumns(tableName)
}

// KV Operations

func (b *RemoteHub) KVAdd(ctx *HttpBindContext) (any, error) {
	req := &dbmodels.SpaceKV{}
	if err := ctx.Http.BindJSON(req); err != nil {
		return nil, err
	}
	kvOps := b.db.GetSpaceKVOps()
	err := kvOps.AddSpaceKV(ctx.PackageId, req)
	return req, err
}

func (b *RemoteHub) KVGet(ctx *HttpBindContext) (any, error) {
	group := ctx.Http.Param("group")
	key := ctx.Http.Param("key")
	kvOps := b.db.GetSpaceKVOps()
	return kvOps.GetSpaceKV(ctx.PackageId, group, key)
}

func (b *RemoteHub) KVQuery(ctx *HttpBindContext) (any, error) {
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

func (b *RemoteHub) KVRemove(ctx *HttpBindContext) (any, error) {
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

func (b *RemoteHub) KVUpdate(ctx *HttpBindContext) (any, error) {
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

func (b *RemoteHub) KVUpsert(ctx *HttpBindContext) (any, error) {
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
