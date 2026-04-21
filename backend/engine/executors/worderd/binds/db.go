package binds

import (
	"fmt"
)

func (s *BindingServer) handleDB(method string, params []any) (any, error) {
	db := s.app.Database().GetLowPackageDBOps(s.installId)
	switch method {
	case "run_query":
		if len(params) < 1 {
			return nil, fmt.Errorf("missing query")
		}
		query := params[0].(string)
		args := params[1:]
		return db.RunQuery(query, args...)
	case "insert":
		if len(params) < 2 {
			return nil, fmt.Errorf("missing table or data")
		}
		table := params[0].(string)
		data := params[1].(map[string]any)
		return db.Insert(table, data)
	case "update_by_id":
		if len(params) < 3 {
			return nil, fmt.Errorf("missing table, id or data")
		}
		table := params[0].(string)
		id := int64(params[1].(float64))
		data := params[2].(map[string]any)
		return nil, db.UpdateById(table, id, data)
	case "delete_by_id":
		if len(params) < 2 {
			return nil, fmt.Errorf("missing table or id")
		}
		table := params[0].(string)
		id := int64(params[1].(float64))
		return nil, db.DeleteById(table, id)
	case "find_by_id":
		if len(params) < 2 {
			return nil, fmt.Errorf("missing table or id")
		}
		table := params[0].(string)
		id := int64(params[1].(float64))
		return db.FindById(table, id)
	case "find_all_by_cond":
		if len(params) < 2 {
			return nil, fmt.Errorf("missing table or cond")
		}
		table := params[0].(string)
		cond := params[1].(map[string]any)
		condAny := make(map[any]any)
		for k, v := range cond {
			condAny[k] = v
		}
		return db.FindAllByCond(table, condAny)
	case "find_one_by_cond":
		if len(params) < 2 {
			return nil, fmt.Errorf("missing table or cond")
		}
		table := params[0].(string)
		cond := params[1].(map[string]any)
		condAny := make(map[any]any)
		for k, v := range cond {
			condAny[k] = v
		}
		return db.FindOneByCond(table, condAny)
	}
	return nil, fmt.Errorf("unknown db method: %s", method)
}
