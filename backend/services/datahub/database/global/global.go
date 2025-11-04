package global

import (
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/upper/db/v4"
)

type GlobalOperations struct {
	db db.Session
}

func NewGlobalOperations(db db.Session) *GlobalOperations {
	return &GlobalOperations{db: db}
}

func (d *GlobalOperations) GetGlobalConfig(key, group string) (*dbmodels.GlobalConfig, error) {
	table := d.globalConfigTable()
	qq.Println("@1", key, group)
	var config dbmodels.GlobalConfig
	err := table.Find(db.Cond{"key": key, "group": group}).One(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (d *GlobalOperations) ListGlobalConfigs(group string, offset int, limit int) ([]dbmodels.GlobalConfig, error) {
	table := d.globalConfigTable()
	var configs []dbmodels.GlobalConfig
	err := table.Find(db.Cond{"group": group}).Offset(offset).Limit(limit).All(&configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func (d *GlobalOperations) AddGlobalConfig(data *dbmodels.GlobalConfig) (int64, error) {
	table := d.globalConfigTable()
	res, err := table.Insert(data)
	if err != nil {
		return 0, err
	}
	return res.ID().(int64), nil
}

func (d *GlobalOperations) UpdateGlobalConfig(id int64, data map[string]any) error {
	table := d.globalConfigTable()
	return table.Find(db.Cond{"id": id}).Update(data)
}

func (d *GlobalOperations) UpdateGlobalConfigByKey(key, group string, data map[string]any) error {
	table := d.globalConfigTable()
	return table.Find(db.Cond{"key": key, "group": group}).Update(data)
}

func (d *GlobalOperations) DeleteGlobalConfig(id int64) error {
	table := d.globalConfigTable()
	return table.Find(db.Cond{"id": id}).Delete()
}

// private

func (d *GlobalOperations) globalConfigTable() db.Collection {
	return d.db.Collection("GlobalConfig")
}
