package binds

import (
	"database/sql"

	"github.com/k0kubun/pp"
	lua "github.com/yuin/gopher-lua"
)

type BluaDb struct {
	db *sql.DB
}

func (b *BluaDb) Bind(l *lua.LState) int {

	tb := l.NewTable()
	query := func(l *lua.LState) int {

		queryString := l.CheckString(1)

		pp.Println("@queryString", queryString)

		rows, err := b.db.Query(queryString)
		if err != nil {
			l.Push(lua.LNil)
			l.Push(lua.LString(err.Error()))
			return 2
		}

		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			l.Push(lua.LNil)
			l.Push(lua.LString(err.Error()))
			return 2
		}

		// Create a result table directly without using temporary hashmap
		result := l.NewTable()

		// Prepare value holders for scanning
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Iterate through rows
		for rows.Next() {
			if err := rows.Scan(valuePtrs...); err != nil {
				l.Push(lua.LNil)
				l.Push(lua.LString(err.Error()))
				return 2
			}

			pp.Println(columns, values)

			// Create a table for this row
			entry := l.NewTable()
			for i, col := range columns {
				pp.Println(col, values[i])
				entry.RawSetString(col, ToArbitraryValue(l, values[i]))
			}
			result.Append(entry)
		}

		if err := rows.Err(); err != nil {
			l.Push(lua.LNil)
			l.Push(lua.LString(err.Error()))
			return 2
		}

		l.Push(result)
		l.Push(lua.LNil)
		return 2
	}

	l.SetFuncs(tb, map[string]lua.LGFunction{
		"query": query,
	})

	l.Push(tb)

	return 1

}
