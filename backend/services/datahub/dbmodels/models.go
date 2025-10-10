package dbmodels

type GlobalConfig struct {
	ID        int64  `db:"id,omitempty" json:"id"`
	Key       string `db:"key" json:"key"`
	GroupName string `db:"group_name" json:"group_name"`
	Value     string `db:"value" json:"value"`
}
