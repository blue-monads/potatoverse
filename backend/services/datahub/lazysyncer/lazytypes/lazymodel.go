package lazytypes

import (
	"bytes"
	"fmt"
	"text/template"
)

// operation =>  0: insert, 1: update, 2: delete, 3: schema_init 4:schema_change

const selfCDCTableTemplate = `
CREATE TABLE IF NOT EXISTS {{.TableName}} (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	record_id INTEGER NOT NULL,
	operation INTEGER NOT NULL,
	payload blob,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

const buddyCDCTableTemplate = `
CREATE TABLE IF NOT EXISTS {{.TableName}} (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	record_id INTEGER NOT NULL,
	operation INTEGER NOT NULL,
	payload blob,
	linked_cdc_id INTEGER NOT NULL DEFAULT 0
);
`

var (
	selfCDCTemplate  *template.Template
	buddyCDCTemplate *template.Template
)

func init() {
	var err error
	selfCDCTemplate, err = template.New("cdc_table").Parse(selfCDCTableTemplate)
	if err != nil {
		panic(fmt.Errorf("failed to parse template: %w", err))
	}

	buddyCDCTemplate, err = template.New("cdc_table").Parse(buddyCDCTableTemplate)
	if err != nil {
		panic(fmt.Errorf("failed to parse template: %w", err))
	}
}

func BuildSelfCDCTableSchema(tableName string) (string, error) {
	var buf bytes.Buffer
	if err := selfCDCTemplate.Execute(&buf, map[string]string{"TableName": tableName}); err != nil {
		return "", fmt.Errorf("failed to execute template for %s: %w", tableName, err)
	}

	return buf.String(), nil
}

func BuildBuddyCDCTableSchema(tableName string) (string, error) {
	var buf bytes.Buffer
	if err := buddyCDCTemplate.Execute(&buf, map[string]string{"TableName": tableName}); err != nil {
		return "", fmt.Errorf("failed to execute template for %s: %w", tableName, err)
	}

	return buf.String(), nil
}
