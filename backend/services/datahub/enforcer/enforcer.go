package enforcer

import (
	"fmt"
	"strings"

	"github.com/rqlite/sql"
)

// TransformQuery transforms all table names in the SQL query to the scoped format:
// zz_{ownerType}__{ownerID}__{tableName}
func TransformQuery(ownerType string, ownerID string, input string) (string, error) {
	return transformQuery(ownerType, ownerID, input)
}

func TableName(ownerType string, ownerID string, tableName string) string {
	return fmt.Sprintf("zz_%s__%s__%s", ownerType, ownerID, tableName)
}

func transformQuery(ownerType string, ownerID string, input string) (string, error) {
	prefix := fmt.Sprintf("zz_%s__%s__", ownerType, ownerID)

	parser := sql.NewParser(strings.NewReader(input))
	stmt, err := parser.ParseStatement()
	if err != nil {
		return "", fmt.Errorf("failed to parse SQL: %w", err)
	}

	// Transform the AST by walking and modifying table names
	transformedStmt, err := sql.Walk(sql.VisitEndFunc(func(n sql.Node) (sql.Node, error) {
		switch node := n.(type) {
		case *sql.QualifiedTableName:
			// Transform table name in FROM, JOIN, etc.
			tableName := node.TableName()
			// Skip if already scoped
			if !strings.HasPrefix(tableName, prefix) {
				cloned := node.Clone()
				if cloned.Name != nil {
					cloned.Name = cloned.Name.Clone()
					cloned.Name.Name = prefix + cloned.Name.Name
				}
				return cloned, nil
			}
			return node, nil

		case *sql.QualifiedRef:
			// Transform table name in column references (table.column)
			if node.Table != nil && node.Table.Name != "" {
				tableName := node.Table.Name
				// Skip if already scoped
				if !strings.HasPrefix(tableName, prefix) {
					cloned := node.Clone()
					if cloned.Table != nil {
						cloned.Table = cloned.Table.Clone()
						cloned.Table.Name = prefix + cloned.Table.Name
					}
					return cloned, nil
				}
			}
			return node, nil

		case *sql.CreateTableStatement:
			// Transform table name in CREATE TABLE
			if node.Name != nil {
				tableName := node.Name.Name
				if !strings.HasPrefix(tableName, prefix) {
					cloned := node.Clone()
					if cloned.Name != nil {
						cloned.Name = cloned.Name.Clone()
						cloned.Name.Name = prefix + cloned.Name.Name
					}
					return cloned, nil
				}
			}
			return node, nil

		case *sql.CreateIndexStatement:
			// Transform table name in CREATE INDEX
			if node.Table != nil {
				tableName := node.Table.Name
				if !strings.HasPrefix(tableName, prefix) {
					cloned := node.Clone()
					if cloned.Table != nil {
						cloned.Table = cloned.Table.Clone()
						cloned.Table.Name = prefix + cloned.Table.Name
					}
					return cloned, nil
				}
			}
			return node, nil

		case *sql.DropTableStatement:
			// Transform table name in DROP TABLE
			if node.Name != nil {
				tableName := node.Name.Name
				if !strings.HasPrefix(tableName, prefix) {
					cloned := node.Clone()
					if cloned.Name != nil {
						cloned.Name = cloned.Name.Clone()
						cloned.Name.Name = prefix + cloned.Name.Name
					}
					return cloned, nil
				}
			}
			return node, nil

		case *sql.InsertStatement:
			// Transform table name in INSERT
			if node.Table != nil {
				tableName := node.Table.Name
				if !strings.HasPrefix(tableName, prefix) {
					cloned := node.Clone()
					if cloned.Table != nil {
						cloned.Table = cloned.Table.Clone()
						cloned.Table.Name = prefix + cloned.Table.Name
					}
					return cloned, nil
				}
			}
			return node, nil

		case *sql.UpdateStatement:
			// Transform table name in UPDATE (Table is *QualifiedTableName, handled by QualifiedTableName case)
			// This case is here for completeness but the actual transformation happens in QualifiedTableName
			return node, nil

		case *sql.DeleteStatement:
			// Transform table name in DELETE (Table is *QualifiedTableName, handled by QualifiedTableName case)
			// This case is here for completeness but the actual transformation happens in QualifiedTableName
			return node, nil

		case *sql.AlterTableStatement:
			// Transform table name in ALTER TABLE
			if node.Name != nil {
				tableName := node.Name.Name
				if !strings.HasPrefix(tableName, prefix) {
					cloned := node.Clone()
					if cloned.Name != nil {
						cloned.Name = cloned.Name.Clone()
						cloned.Name.Name = prefix + cloned.Name.Name
					}
					return cloned, nil
				}
			}
			return node, nil
		}

		return n, nil
	}), stmt)

	if err != nil {
		return "", fmt.Errorf("failed to transform SQL: %w", err)
	}

	// Convert the transformed AST back to SQL string
	return transformedStmt.String(), nil
}
