package enforcer

import (
	"fmt"
	"strings"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
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

		// Pre-pass: collect all aliases so we don't prefix them in QualifiedRef.
		aliases := collectAliases(stmt)
		qq.Println("aliases", i, aliases)

		transformedStmt, err := sql.Walk(sql.VisitEndFunc(func(n sql.Node) (sql.Node, error) {
			qq.Println("node type", i, fmt.Sprintf("%T", n))

			switch node := n.(type) {

			case sql.SelectExpr:
				subSQL := node.String()
				transformedSubSQL, err := transformQuery(ownerType, ownerID, subSQL)
				if err != nil {
					return nil, fmt.Errorf("failed to transform subquery expr: %w", err)
				}
				subParser := sql.NewParser(strings.NewReader(transformedSubSQL))
				subStmt, err := subParser.ParseStatement()
				if err != nil {
					return nil, fmt.Errorf("failed to re-parse subquery expr: %w", err)
				}
				return sql.SelectExpr{SelectStatement: subStmt.(*sql.SelectStatement)}, nil
			case *sql.QualifiedTableName:
				tableName := node.TableName()
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
				if node.Table != nil && node.Table.Name != "" {
					refTable := node.Table.Name
					// Skip aliases — they point to an already-renamed table.
					// Skip already-scoped names too.
					if aliases[refTable] || strings.HasPrefix(refTable, prefix) {
						return node, nil
					}
					cloned := node.Clone()
					if cloned.Table != nil {
						cloned.Table = cloned.Table.Clone()
						cloned.Table.Name = prefix + cloned.Table.Name
					}
					return cloned, nil
				}
				return node, nil

			case *sql.CreateTableStatement:
				if node.Name != nil && !strings.HasPrefix(node.Name.Name, prefix) {
					cloned := node.Clone()
					if cloned.Name != nil {
						cloned.Name = cloned.Name.Clone()
						cloned.Name.Name = prefix + cloned.Name.Name
					}
					return cloned, nil
				}
				return node, nil

			case *sql.CreateVirtualTableStatement:
				if node.Name != nil {
					tableName := node.Name.Name
					if node.Schema != nil && node.Schema.Name != "" {
						tableName = node.Schema.Name
					}
					if !strings.HasPrefix(tableName, prefix) {
						cloned := node.Clone()
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
				if node.Table != nil && !strings.HasPrefix(node.Table.Name, prefix) {
					cloned := node.Clone()
					if cloned.Table != nil {
						cloned.Table = cloned.Table.Clone()
						cloned.Table.Name = prefix + cloned.Table.Name
					}
					return cloned, nil
				}
				return node, nil

			case *sql.DropTableStatement:
				if node.Name != nil && !strings.HasPrefix(node.Name.Name, prefix) {
					cloned := node.Clone()
					if cloned.Name != nil {
						cloned.Name = cloned.Name.Clone()
						cloned.Name.Name = prefix + cloned.Name.Name
					}
					return cloned, nil
				}
				return node, nil

			case *sql.InsertStatement:
				if node.Table != nil && !strings.HasPrefix(node.Table.Name, prefix) {
					cloned := node.Clone()
					if cloned.Table != nil {
						cloned.Table = cloned.Table.Clone()
						cloned.Table.Name = prefix + cloned.Table.Name
					}
					return cloned, nil
				}
				return node, nil

			case *sql.UpdateStatement:
				// Actual table name transformation is handled by the QualifiedTableName case.
				return node, nil

			case *sql.DeleteStatement:
				// Actual table name transformation is handled by the QualifiedTableName case.
				return node, nil

			case *sql.AlterTableStatement:
				if node.Name != nil && !strings.HasPrefix(node.Name.Name, prefix) {
					cloned := node.Clone()
					if cloned.Name != nil {
						cloned.Name = cloned.Name.Clone()
						cloned.Name.Name = prefix + cloned.Name.Name
					}
					return cloned, nil
				}
				return node, nil

			case *sql.ForeignKeyConstraint:
				if node.ForeignTable != nil && !strings.HasPrefix(node.ForeignTable.Name, prefix) {
					cloned := node.Clone()
					if cloned.ForeignTable != nil {
						cloned.ForeignTable = cloned.ForeignTable.Clone()
						cloned.ForeignTable.Name = prefix + cloned.ForeignTable.Name
					}
					return cloned, nil
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

// collectAliases does a pre-pass over the AST and returns a set of all alias
// names defined in QualifiedTableName nodes (e.g. "e" from "Events AS e").
// These must NOT be prefixed when they appear as the table qualifier in a
// QualifiedRef, because they already refer to an aliased (and already-renamed)
// table.
func collectAliases(node sql.Node) map[string]bool {
	aliases := make(map[string]bool)
	_, _ = sql.Walk(sql.VisitEndFunc(func(n sql.Node) (sql.Node, error) {
		if qtn, ok := n.(*sql.QualifiedTableName); ok {
			if qtn.Alias != nil && qtn.Alias.Name != "" {
				aliases[qtn.Alias.Name] = true
			}
		}
		return n, nil
	}), node)
	return aliases
}

// splitStatements splits SQL input into individual statements by semicolon,
// respecting single- and double-quoted strings.
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
			if prevRune == '\\' && inSingleQuote {
				current.WriteRune(r)
				continue
			}
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
			current.WriteRune(r)
		case '"':
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
				if stmt := current.String(); strings.TrimSpace(stmt) != "" {
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

	if current.Len() > 0 {
		if stmt := current.String(); strings.TrimSpace(stmt) != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}
