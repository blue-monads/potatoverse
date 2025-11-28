package dbmodels

type GlobalConfig struct {
	ID        int64  `db:"id,omitempty" json:"id"`
	Key       string `db:"key" json:"key"`
	GroupName string `db:"group" json:"group"`
	Value     string `db:"value" json:"value"`
}

type EntityId struct {
	Id int64 `db:"id" json:"id"`
}
