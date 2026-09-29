package remotehub

import (
	"github.com/blue-monads/potatoverse/backend/engine/executors/core"
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

func (b *RemoteHub) DBRunQuery(ctx RContext) (any, error) {
	var req core.DBQueryReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	res, err := dbOps.RunQuery(req.Query, req.Args...)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) DBRunQueryOne(ctx RContext) (any, error) {
	var req core.DBQueryReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	res, err := dbOps.RunQueryOne(req.Query, req.Args...)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) DBInsert(ctx RContext) (any, error) {
	var req core.DBInsertReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	res, err := dbOps.Insert(req.Table, req.Data)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) DBUpdateById(ctx RContext) (any, error) {
	var req core.DBUpdateByIdReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	err := dbOps.UpdateById(req.Table, req.ID, req.Data)
	return nil, err
}

func (b *RemoteHub) DBDeleteById(ctx RContext) (any, error) {
	var req core.DBIdReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	err := dbOps.DeleteById(req.Table, req.ID)
	return nil, err
}

func (b *RemoteHub) DBFindById(ctx RContext) (any, error) {
	var req core.DBIdReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	res, err := dbOps.FindById(req.Table, req.ID)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) DBUpdateByCond(ctx RContext) (any, error) {
	var req core.DBUpdateByCondReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	err := dbOps.UpdateByCond(req.Table, toMapAnyAny(req.Cond), req.Data)
	return nil, err
}

func (b *RemoteHub) DBDeleteByCond(ctx RContext) (any, error) {
	var req core.DBCondReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	err := dbOps.DeleteByCond(req.Table, toMapAnyAny(req.Cond))
	return nil, err
}

func (b *RemoteHub) DBFindAllByCond(ctx RContext) (any, error) {
	var req core.DBCondReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	res, err := dbOps.FindAllByCond(req.Table, toMapAnyAny(req.Cond))
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) DBFindOneByCond(ctx RContext) (any, error) {
	var req core.DBCondReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	res, err := dbOps.FindOneByCond(req.Table, toMapAnyAny(req.Cond))
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) DBFindAllByQuery(ctx RContext) (any, error) {
	req := &datahub.FindQuery{}
	if err := bindJSON(ctx, req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	res, err := dbOps.FindAllByQuery(req)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) DBFindByJoin(ctx RContext) (any, error) {
	req := &datahub.FindByJoin{}
	if err := bindJSON(ctx, req); err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	res, err := dbOps.FindByJoin(req)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) DBListTables(ctx RContext) (any, error) {
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	res, err := dbOps.ListTables()
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) DBListColumns(ctx RContext) (any, error) {
	tableName, err := ctx.GetMeta("table")
	if err != nil {
		return nil, err
	}
	dbOps := b.db.GetLowPackageDBOps(ctx.GetPackageId())
	res, err := dbOps.ListTableColumns(tableName)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

// KV Operations

func (b *RemoteHub) KVAdd(ctx RContext) (any, error) {
	req := &dbmodels.SpaceKV{}
	if err := bindJSON(ctx, req); err != nil {
		return nil, err
	}
	kvOps := b.db.GetSpaceKVOps()
	err := kvOps.AddSpaceKV(ctx.GetPackageId(), req)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, req)
	return req, nil
}

func (b *RemoteHub) KVGet(ctx RContext) (any, error) {
	group, err := ctx.GetMeta("group")
	if err != nil {
		return nil, err
	}
	key, err := ctx.GetMeta("key")
	if err != nil {
		return nil, err
	}
	kvOps := b.db.GetSpaceKVOps()
	res, err := kvOps.GetSpaceKV(ctx.GetPackageId(), group, key)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) KVQuery(ctx RContext) (any, error) {
	var req core.KVQueryReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	kvOps := b.db.GetSpaceKVOps()
	if req.IncludeValue {
		res, err := kvOps.QueryWithValueSpaceKV(ctx.GetPackageId(), toMapAnyAny(req.Cond), req.Offset, req.Limit)
		if err != nil {
			return nil, err
		}
		setDataJSON(ctx, res)
		return res, nil
	}
	res, err := kvOps.QuerySpaceKV(ctx.GetPackageId(), toMapAnyAny(req.Cond), req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) KVRemove(ctx RContext) (any, error) {
	var req core.KVKeyReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	kvOps := b.db.GetSpaceKVOps()
	err := kvOps.RemoveSpaceKV(ctx.GetPackageId(), req.Group, req.Key)
	return nil, err
}

func (b *RemoteHub) KVUpdate(ctx RContext) (any, error) {
	var req core.KVDataReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	kvOps := b.db.GetSpaceKVOps()
	err := kvOps.UpdateSpaceKV(ctx.GetPackageId(), req.Group, req.Key, req.Data)
	return nil, err
}

func (b *RemoteHub) KVUpsert(ctx RContext) (any, error) {
	var req core.KVDataReq
	if err := bindJSON(ctx, &req); err != nil {
		return nil, err
	}
	kvOps := b.db.GetSpaceKVOps()
	err := kvOps.UpsertSpaceKV(ctx.GetPackageId(), req.Group, req.Key, req.Data)
	return nil, err
}
