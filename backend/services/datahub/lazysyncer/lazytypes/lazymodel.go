package lazytypes

import (
	"bytes"
	"fmt"
	"text/template"
)

const CDCTableTemplate = `
CREATE TABLE IF NOT EXISTS {{.TableName}} (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	record_id INTEGER NOT NULL,
	operation INTEGER NOT NULL, -- 0: insert, 1: update, 2: delete, 3: schema_init 4:schema_change
	payload blob,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

var (
	cdcTableTmpl *template.Template
)

func init() {
	var err error
	cdcTableTmpl, err = template.New("cdc_table").Parse(CDCTableTemplate)
	if err != nil {
		panic(fmt.Errorf("failed to parse template: %w", err))
	}
}

func BuildCDCTableSchema(tableName string) (string, error) {
	var buf bytes.Buffer
	if err := cdcTableTmpl.Execute(&buf, map[string]string{"TableName": tableName}); err != nil {
		return "", fmt.Errorf("failed to execute template for %s: %w", tableName, err)
	}

	return buf.String(), nil
}
