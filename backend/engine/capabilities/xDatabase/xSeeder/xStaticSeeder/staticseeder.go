package staticseeder

type StaticSeederStruct struct {
	TableName string           `json:"table_name"`
	Data      []map[string]any `json:"data"`
}
