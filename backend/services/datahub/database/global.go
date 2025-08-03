package database

import (
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/models"
	"github.com/upper/db/v4"
)

var _ datahub.GlobalOps = (*DB)(nil)

func (d *DB) GetGlobalConfig(key, group string) (*models.GlobalConfig, error) {
	table := d.globalConfigTable()
	var config models.GlobalConfig
	err := table.Find(db.Cond{"key": key, "group": group}).One(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (d *DB) ListGlobalConfigs(group string, offset int, limit int) ([]models.GlobalConfig, error) {
	table := d.globalConfigTable()
	var configs []models.GlobalConfig
	err := table.Find(db.Cond{"group": group}).Offset(offset).Limit(limit).All(&configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func (d *DB) AddGlobalConfig(data *models.GlobalConfig) (int64, error) {
	table := d.globalConfigTable()
	res, err := table.Insert(data)
	if err != nil {
		return 0, err
	}
	return res.ID().(int64), nil
}

func (d *DB) UpdateGlobalConfig(id int64, data map[string]any) error {
	table := d.globalConfigTable()
	return table.Find(db.Cond{"id": id}).Update(data)
}

func (d *DB) UpdateGlobalConfigByKey(key, group string, data map[string]any) error {
	table := d.globalConfigTable()
	return table.Find(db.Cond{"key": key, "group": group}).Update(data)
}

func (d *DB) DeleteGlobalConfig(id int64) error {
	table := d.globalConfigTable()
	return table.Find(db.Cond{"id": id}).Delete()
}

// private

func (d *DB) globalConfigTable() db.Collection {
	return d.Table("global_config")
}
