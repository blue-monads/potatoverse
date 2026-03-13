package xremotedb

import (
	"fmt"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
)

type RemoteDbCapability struct {
	installId int64
	db        datahub.DBLowOps
	capHandle xcapability.XCapabilityHandle
}

var actions = []string{
	"run_ddl",
	"run_query",
	"run_query_one",
	"exec",
	"insert",
	"update_by_id",
	"delete_by_id",
	"find_by_id",
	"update_by_cond",
	"delete_by_cond",
	"find_all_by_cond",
	"find_one_by_cond",
	"find_all_by_query",
	"find_by_join",
	"list_tables",
	"list_table_columns",
	"find_table_pk",
}

func (r *RemoteDbCapability) ListActions() ([]string, error) {
	return actions, nil
}

func (r *RemoteDbCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "run_ddl":
		return r.execRunDDL(params)
	case "run_query":
		return r.execRunQuery(params)
	case "run_query_one":
		return r.execRunQueryOne(params)
	case "exec":
		return r.execExec(params)
	case "insert":
		return r.execInsert(params)
	case "update_by_id":
		return r.execUpdateById(params)
	case "delete_by_id":
		return r.execDeleteById(params)
	case "find_by_id":
		return r.execFindById(params)
	case "update_by_cond":
		return r.execUpdateByCond(params)
	case "delete_by_cond":
		return r.execDeleteByCond(params)
	case "find_all_by_cond":
		return r.execFindAllByCond(params)
	case "find_one_by_cond":
		return r.execFindOneByCond(params)
	case "find_all_by_query":
		return r.execFindAllByQuery(params)
	case "find_by_join":
		return r.execFindByJoin(params)
	case "list_tables":
		return r.db.ListTables()
	case "list_table_columns":
		return r.execListTableColumns(params)
	case "find_table_pk":
		return r.execFindTablePK(params)
	default:
		return nil, fmt.Errorf("unknown action: %s", name)
	}
}

func (r *RemoteDbCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	return r, nil
}

func (r *RemoteDbCapability) Close() error {
	return nil
}

// param structs

type queryParams struct {
	Query string `json:"query"`
	Data  []any  `json:"data"`
}

type tableOnlyParams struct {
	Table string `json:"table"`
}

type tableIdParams struct {
	Table string `json:"table"`
	Id    int64  `json:"id"`
}

type insertParams struct {
	Table string         `json:"table"`
	Data  map[string]any `json:"data"`
}

type updateByIdParams struct {
	Table string         `json:"table"`
	Id    int64          `json:"id"`
	Data  map[string]any `json:"data"`
}

type condParams struct {
	Table string         `json:"table"`
	Cond  map[string]any `json:"cond"`
}

type updateByCondParams struct {
	Table string         `json:"table"`
	Cond  map[string]any `json:"cond"`
	Data  map[string]any `json:"data"`
}

type findAllByQueryParams struct {
	Table  string         `json:"table"`
	Cond   map[string]any `json:"cond"`
	Offset int            `json:"offset"`
	Limit  int            `json:"limit"`
	Order  string         `json:"order"`
	Fields []string       `json:"fields"`
}

type findByJoinParams struct {
	Joins  []datahub.Join `json:"joins"`
	Cond   map[string]any `json:"cond"`
	Order  string         `json:"order"`
	Fields []string       `json:"fields"`
}

func toAnyCond(cond map[string]any) map[any]any {
	result := make(map[any]any, len(cond))
	for k, v := range cond {
		result[k] = v
	}
	return result
}

// action implementations

func (r *RemoteDbCapability) execRunDDL(params lazydata.LazyData) (any, error) {
	p := &queryParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return nil, r.db.RunDDL(p.Query)
}

func (r *RemoteDbCapability) execRunQuery(params lazydata.LazyData) (any, error) {
	p := &queryParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return r.db.RunQuery(p.Query, p.Data...)
}

func (r *RemoteDbCapability) execRunQueryOne(params lazydata.LazyData) (any, error) {
	p := &queryParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return r.db.RunQueryOne(p.Query, p.Data...)
}

func (r *RemoteDbCapability) execExec(params lazydata.LazyData) (any, error) {
	p := &queryParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return r.db.Exec(p.Query, p.Data...)
}

func (r *RemoteDbCapability) execInsert(params lazydata.LazyData) (any, error) {
	p := &insertParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	id, err := r.db.Insert(p.Table, p.Data)
	if err != nil {
		return nil, err
	}
	return map[string]any{"id": id}, nil
}

func (r *RemoteDbCapability) execUpdateById(params lazydata.LazyData) (any, error) {
	p := &updateByIdParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return nil, r.db.UpdateById(p.Table, p.Id, p.Data)
}

func (r *RemoteDbCapability) execDeleteById(params lazydata.LazyData) (any, error) {
	p := &tableIdParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return nil, r.db.DeleteById(p.Table, p.Id)
}

func (r *RemoteDbCapability) execFindById(params lazydata.LazyData) (any, error) {
	p := &tableIdParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return r.db.FindById(p.Table, p.Id)
}

func (r *RemoteDbCapability) execUpdateByCond(params lazydata.LazyData) (any, error) {
	p := &updateByCondParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return nil, r.db.UpdateByCond(p.Table, toAnyCond(p.Cond), p.Data)
}

func (r *RemoteDbCapability) execDeleteByCond(params lazydata.LazyData) (any, error) {
	p := &condParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return nil, r.db.DeleteByCond(p.Table, toAnyCond(p.Cond))
}

func (r *RemoteDbCapability) execFindAllByCond(params lazydata.LazyData) (any, error) {
	p := &condParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return r.db.FindAllByCond(p.Table, toAnyCond(p.Cond))
}

func (r *RemoteDbCapability) execFindOneByCond(params lazydata.LazyData) (any, error) {
	p := &condParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return r.db.FindOneByCond(p.Table, toAnyCond(p.Cond))
}

func (r *RemoteDbCapability) execFindAllByQuery(params lazydata.LazyData) (any, error) {
	p := &findAllByQueryParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return r.db.FindAllByQuery(&datahub.FindQuery{
		Table:  p.Table,
		Offset: p.Offset,
		Limit:  p.Limit,
		Cond:   toAnyCond(p.Cond),
		Order:  p.Order,
		Fields: p.Fields,
	})
}

func (r *RemoteDbCapability) execFindByJoin(params lazydata.LazyData) (any, error) {
	p := &findByJoinParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return r.db.FindByJoin(&datahub.FindByJoin{
		Joins:  p.Joins,
		Cond:   toAnyCond(p.Cond),
		Order:  p.Order,
		Fields: p.Fields,
	})
}

func (r *RemoteDbCapability) execListTableColumns(params lazydata.LazyData) (any, error) {
	p := &tableOnlyParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	return r.db.ListTableColumns(p.Table)
}

func (r *RemoteDbCapability) execFindTablePK(params lazydata.LazyData) (any, error) {
	p := &tableOnlyParams{}
	if err := params.AsJson(p); err != nil {
		return nil, err
	}
	pk, err := r.db.FindTablePK(p.Table)
	if err != nil {
		return nil, err
	}
	return map[string]any{"pk": pk}, nil
}
