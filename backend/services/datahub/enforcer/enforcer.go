package enforcer

import (
	"fmt"
	"strings"

	"github.com/blue-monads/turnix/backend/utils/qq"
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

func TableNamePattern(ownerType string, ownerID string) string {
	return fmt.Sprintf("zz_%s__%s__%%", ownerType, ownerID)
}

func transformQuery(ownerType string, ownerID string, input string) (string, error) {
	prefix := fmt.Sprintf("zz_%s__%s__", ownerType, ownerID)

	// Split input into individual statements by semicolon
	statements := splitStatements(input)

	qq.Println("statements", statements)

	var transformedStatements []string
	for i, stmtStr := range statements {

		qq.Println("stmtStr", i, stmtStr)

		stmtStr = strings.TrimSpace(stmtStr)
		if stmtStr == "" {
			continue
		}

		parser := sql.NewParser(strings.NewReader(stmtStr))
		stmt, err := parser.ParseStatement()
		if err != nil {
			qq.Println("failed to parse SQL", i, err, "stmtStr:", stmtStr)
			return "", fmt.Errorf("failed to parse SQL: %w", err)
		}

		// Transform the AST by walking and modifying table names
		transformedStmt, err := sql.Walk(sql.VisitEndFunc(func(n sql.Node) (sql.Node, error) {
			// Log node type for debugging
			qq.Println("node type", i, fmt.Sprintf("%T", n))

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

			case *sql.CreateVirtualTableStatement:
				// Transform table name in CREATE VIRTUAL TABLE
				if node.Name != nil {
					tableName := node.Name.Name
					// If Schema is set, use Schema.Name as the table name, otherwise use Name.Name
					if node.Schema != nil && node.Schema.Name != "" {
						tableName = node.Schema.Name
					}
					if !strings.HasPrefix(tableName, prefix) {
						cloned := node.Clone()
						// Clear Schema if it was set (we want unqualified table names)
						cloned.Schema = nil
						if cloned.Name != nil {
							cloned.Name = cloned.Name.Clone()
							cloned.Name.Name = prefix + tableName
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

			case *sql.ForeignKeyConstraint:
				// Transform foreign table name in FOREIGN KEY constraints
				if node.ForeignTable != nil {
					tableName := node.ForeignTable.Name
					// Skip if already scoped
					if !strings.HasPrefix(tableName, prefix) {
						cloned := node.Clone()
						if cloned.ForeignTable != nil {
							cloned.ForeignTable = cloned.ForeignTable.Clone()
							cloned.ForeignTable.Name = prefix + cloned.ForeignTable.Name
						}
						return cloned, nil
					}
				}
				return node, nil
			}

			qq.Println("node/end", i)

			return n, nil
		}), stmt)

		if err != nil {
			qq.Println("failed to transform SQL", i, err, "stmtStr:", stmtStr)
			return "", fmt.Errorf("failed to transform SQL: %w", err)
		}

		transformedSQL := transformedStmt.String()
		transformedStatements = append(transformedStatements, transformedSQL)

		qq.Println("transformedSQL", i, transformedSQL)
	}

	result := strings.Join(transformedStatements, ";\n\n")

	qq.Println("input", input)
	qq.Println("transformedSQL", result)

	return result, nil
}

// splitStatements splits SQL input into individual statements by semicolon.
func splitStatements(input string) []string {
	var statements []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		prevRune := rune(0)
		if i > 0 {
			prevRune = runes[i-1]
		}

		switch r {
		case '\'':
			// Check if quote is escaped (previous char is backslash)
			if prevRune == '\\' && inSingleQuote {
				current.WriteRune(r)
				continue
			}
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
			current.WriteRune(r)
		case '"':
			// Check if quote is escaped (previous char is backslash)
			if prevRune == '\\' && inDoubleQuote {
				current.WriteRune(r)
				continue
			}
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
			current.WriteRune(r)
		case ';':
			if !inSingleQuote && !inDoubleQuote {
				stmt := current.String()
				if strings.TrimSpace(stmt) != "" {
					statements = append(statements, stmt)
				}
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	// Add the last statement if there's no trailing semicolon
	if current.Len() > 0 {
		stmt := current.String()
		if strings.TrimSpace(stmt) != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}
