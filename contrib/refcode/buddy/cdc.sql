CREATE TABLE change_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT DEFAULT CURRENT_TIMESTAMP,
    table_name TEXT,
    action TEXT,
    changes TEXT
);


CREATE TRIGGER log_insert
AFTER INSERT ON your_table_name
BEGIN
    INSERT INTO change_log (table_name, action, changes)
    SELECT 'your_table_name', 'INSERT', json_group_object(
        'id', NEW.id,
        'column1', NEW.column1,
        'column2', NEW.column2
        -- Add more columns as needed
    );
END;

CREATE TRIGGER log_update
AFTER UPDATE ON your_table_name
BEGIN
    INSERT INTO change_log (table_name, action, changes)
    SELECT 'your_table_name', 'UPDATE', json_patch(
        json_object(
            'old', json_group_object(
                'id', OLD.id,
                'column1', OLD.column1,
                'column2', OLD.column2
                -- Add more columns as needed
            )
        ),
        json_object(
            'new', json_group_object(
                'id', NEW.id,
                'column1', NEW.column1,
                'column2', NEW.column2
                -- Add more columns as needed
            )
        )
    );
END;

CREATE TRIGGER log_delete
AFTER DELETE ON your_table_name
BEGIN
    INSERT INTO change_log (table_name, action, changes)
    SELECT 'your_table_name', 'DELETE', json_group_object(
        'id', OLD.id,
        'column1', OLD.column1,
        'column2', OLD.column2
        -- Add more columns as needed
    );
END;


CREATE TABLE employees (
    id INTEGER PRIMARY KEY,
    name TEXT,
    salary REAL
);