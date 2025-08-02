package autocdc

import (
	"fmt"
	"strings"
)

func (ac *AutoCapture) buildCaptureDDL2(tableName string, columns []string) []string {
	var b strings.Builder
	templates := make([]string, 0)

	logTableDDl := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS zcdc_log_%s (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            event_data TEXT NOT NULL,
            event_type TEXT NOT NULL,
            row_id INTEGER NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`, tableName)
	templates = append(templates, logTableDDl)

	templates = append(templates, fmt.Sprintf(`DROP TRIGGER IF EXISTS %s_insert_trigger;`, tableName))
	templates = append(templates, fmt.Sprintf(`DROP TRIGGER IF EXISTS %s_update_trigger;`, tableName))
	templates = append(templates, fmt.Sprintf(`DROP TRIGGER IF EXISTS %s_delete_trigger;`, tableName))

	var jsonPartsInsert, jsonPartsUpdate, jsonPartsDelete []string
	for _, col := range columns {
		jsonPartsInsert = append(jsonPartsInsert, fmt.Sprintf("'%s', NEW.%s", col, col))
		jsonPartsUpdate = append(jsonPartsUpdate, fmt.Sprintf("'%s', NEW.%s", col, col))
		jsonPartsDelete = append(jsonPartsDelete, fmt.Sprintf("'%s', OLD.%s", col, col))
	}

	insertValues := strings.Join(jsonPartsInsert, ", ")
	updateValues := strings.Join(jsonPartsUpdate, ", ")
	deleteValues := strings.Join(jsonPartsDelete, ", ")

	b.Reset()
	b.WriteString(fmt.Sprintf("CREATE TRIGGER %s_insert_trigger\n", tableName))
	b.WriteString(fmt.Sprintf("AFTER INSERT ON %s\n", tableName))
	b.WriteString("FOR EACH ROW\nBEGIN\n")
	b.WriteString(fmt.Sprintf("    INSERT INTO zcdc_log_%s (event_data, event_type, row_id)\n", tableName))
	b.WriteString(fmt.Sprintf("    VALUES (json_object(%s), 'insert', NEW.id);\n", insertValues))
	b.WriteString("END;")
	templates = append(templates, b.String())

	b.Reset()
	b.WriteString(fmt.Sprintf("CREATE TRIGGER %s_update_trigger\n", tableName))
	b.WriteString(fmt.Sprintf("AFTER UPDATE ON %s\n", tableName))
	b.WriteString("FOR EACH ROW\nBEGIN\n")
	b.WriteString(fmt.Sprintf("    INSERT INTO zcdc_log_%s (event_data, event_type, row_id)\n", tableName))
	b.WriteString(fmt.Sprintf("    VALUES (json_object(%s), 'update', NEW.id);\n", updateValues))
	b.WriteString("END;")
	templates = append(templates, b.String())

	// --- Delete Trigger ---
	b.Reset()
	b.WriteString(fmt.Sprintf("CREATE TRIGGER %s_delete_trigger\n", tableName))
	b.WriteString(fmt.Sprintf("AFTER DELETE ON %s\n", tableName))
	b.WriteString("FOR EACH ROW\nBEGIN\n")
	b.WriteString(fmt.Sprintf("    INSERT INTO zcdc_log_%s (event_data, event_type, row_id)\n", tableName))
	b.WriteString(fmt.Sprintf("    VALUES (json_object(%s), 'delete', OLD.id);\n", deleteValues))
	b.WriteString("END;")
	templates = append(templates, b.String())

	return templates
}
