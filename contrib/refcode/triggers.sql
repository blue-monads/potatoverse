
	CREATE TABLE IF NOT EXISTS zcdc_log_test (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            event_data TEXT NOT NULL,
            event_type TEXT NOT NULL,
            row_id INTEGER NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
DROP TRIGGER IF EXISTS test_insert_trigger;
DROP TRIGGER IF EXISTS test_update_trigger;
DROP TRIGGER IF EXISTS test_delete_trigger;
CREATE TRIGGER test_insert_trigger
AFTER INSERT ON test
FOR EACH ROW
BEGIN
    INSERT INTO zcdc_log_test (event_data, event_type, row_id)
    VALUES (json_object('id', NEW.id, 'name', NEW.name, 'age', NEW.age, 'created_at', NEW.created_at), 'insert', NEW.id);
END;
CREATE TRIGGER test_update_trigger
AFTER UPDATE ON test
FOR EACH ROW
BEGIN
    INSERT INTO zcdc_log_test (event_data, event_type, row_id)
    VALUES (json_object('id', NEW.id, 'name', NEW.name, 'age', NEW.age, 'created_at', NEW.created_at), 'update', NEW.id);
END;
CREATE TRIGGER test_delete_trigger
AFTER DELETE ON test
FOR EACH ROW
BEGIN
    INSERT INTO zcdc_log_test (event_data, event_type, row_id)
    VALUES (json_object('id', OLD.id, 'name', OLD.name, 'age', OLD.age, 'created_at', OLD.created_at), 'delete', OLD.id);
END;
