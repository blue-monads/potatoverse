package selfcdc

import (
	"bytes"
	"fmt"
)

func buildTriggerSchema(tableName string, pkColumn string) (string, error) {

	var buf bytes.Buffer

	cdcTableName := tableName + "__cdc"

	insertTrigger := fmt.Sprintf(`
		CREATE TRIGGER IF NOT EXISTS %s_insert__cdc
		AFTER INSERT ON %s
		BEGIN
			INSERT INTO %s (record_id, operation) VALUES (NEW.%s, 0);
		END;	

	`, tableName, tableName, cdcTableName, pkColumn)

	buf.WriteString(insertTrigger)

	updateTrigger := fmt.Sprintf(`
		CREATE TRIGGER IF NOT EXISTS %s_update__cdc
		AFTER UPDATE ON %s
		BEGIN
			INSERT INTO %s (record_id, operation) VALUES (NEW.%s, 1);
		END;
	`, tableName, tableName, cdcTableName, pkColumn)

	buf.WriteString(updateTrigger)

	deleteTrigger := fmt.Sprintf(`
		CREATE TRIGGER IF NOT EXISTS %s_delete__cdc
		AFTER DELETE ON %s
		BEGIN
			INSERT INTO %s (record_id, operation) VALUES (OLD.%s, 2);
		END;
	`, tableName, tableName, cdcTableName, pkColumn)

	buf.WriteString(deleteTrigger)

	return buf.String(), nil

}

func buildDropTriggerSchema(tableName string) (string, error) {
	var buf bytes.Buffer

	insertTrigger := fmt.Sprintf(`
		DROP TRIGGER IF EXISTS %s_insert__cdc;
	`, tableName)

	buf.WriteString(insertTrigger)

	updateTrigger := fmt.Sprintf(`
		DROP TRIGGER IF EXISTS %s_update__cdc;
	`, tableName)

	buf.WriteString(updateTrigger)

	deleteTrigger := fmt.Sprintf(`
		DROP TRIGGER IF EXISTS %s_delete__cdc;
	`, tableName)

	buf.WriteString(deleteTrigger)

	cdcTableDrop := fmt.Sprintf(`
		DROP TABLE IF EXISTS %s;
	`, tableName+"__cdc")

	buf.WriteString(cdcTableDrop)

	return buf.String(), nil
}
