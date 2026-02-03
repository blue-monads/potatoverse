package schema

import _ "embed"

//go:embed schema.sql
var schema string

func Get() string {
	return schema
}
