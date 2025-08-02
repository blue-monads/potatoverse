package blua

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type BluaDb struct {
	db            *sql.DB
	preparedCache map[string]*sql.Stmt
}

func (b *BluaDb) Bind(l *lua.LState) error {
	tb := l.NewTable()

	query := func(l *lua.LState) int {
		query := l.CheckString(1)
		rows, err := b.db.Query(query)
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
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
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

			// Create a table for this row
			entry := l.NewTable()
			for i, col := range columns {
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
		return 1
	}

	prepare := func(l *lua.LState) int {
		query := l.CheckString(1)
		stmt, err := b.db.Prepare(query)
		if err != nil {
			l.Push(lua.LNil)
			l.Push(lua.LString(err.Error()))
			return 2
		}
		b.preparedCache[query] = stmt
		l.Push(lua.LString(query))
		return 1
	}

	l.SetFuncs(tb, map[string]lua.LGFunction{
		"query":   query,
		"prepare": prepare,
	})

	return nil
}

type SimpleLuaORM interface {
	QueryData(conds map[string]any) ([]map[string]any, error)

	InsertData(table string, data map[string]any) (int64, error)
	UpdateData(table string, cond map[string]any, data map[string]any) (int64, error)
	DeleteData(table string) (int64, error)
}

type Record struct {
}

// ToArbitraryValue converts Go values to Lua values
func ToArbitraryValue(l *lua.LState, i interface{}) lua.LValue {
	if i == nil {
		return lua.LNil
	}

	switch ii := i.(type) {
	case bool:
		return lua.LBool(ii)
	case int:
		return lua.LNumber(ii)
	case int8:
		return lua.LNumber(ii)
	case int16:
		return lua.LNumber(ii)
	case int32:
		return lua.LNumber(ii)
	case int64:
		return lua.LNumber(ii)
	case uint:
		return lua.LNumber(ii)
	case uint8:
		return lua.LNumber(ii)
	case uint16:
		return lua.LNumber(ii)
	case uint32:
		return lua.LNumber(ii)
	case uint64:
		return lua.LNumber(ii)
	case float64:
		return lua.LNumber(ii)
	case float32:
		return lua.LNumber(ii)
	case string:
		return lua.LString(ii)
	case []byte:
		return lua.LString(ii)
	default:
		v := reflect.ValueOf(i)
		switch v.Kind() {
		case reflect.Ptr:
			return ToArbitraryValue(l, v.Elem().Interface())
		case reflect.Struct:
			return ToTableFromStruct(l, v)
		case reflect.Map:
			return ToTableFromMap(l, v)
		case reflect.Slice:
			return ToTableFromSlice(l, v)
		default:
			// Handle sql.RawBytes specifically, which is commonly returned by database/sql
			return lua.LString(fmt.Sprintf("%v", i))
		}
	}
}

func ToTableFromStruct(l *lua.LState, v reflect.Value) lua.LValue {
	tb := l.NewTable()
	return toTableFromStructInner(l, tb, v)
}

func toTableFromStructInner(l *lua.LState, tb *lua.LTable, v reflect.Value) lua.LValue {
	t := v.Type()
	for j := 0; j < v.NumField(); j++ {
		var inline bool
		name := t.Field(j).Name
		if tag := t.Field(j).Tag.Get("luautil"); tag != "" {
			tagParts := strings.Split(tag, ",")
			if tagParts[0] == "-" {
				continue
			} else if tagParts[0] != "" {
				name = tagParts[0]
			}
			if len(tagParts) > 1 && tagParts[1] == "inline" {
				inline = true
			}
		}
		if inline {
			toTableFromStructInner(l, tb, v.Field(j))
		} else {
			tb.RawSetString(name, ToArbitraryValue(l, v.Field(j).Interface()))
		}
	}
	return tb
}

func ToTableFromMap(l *lua.LState, v reflect.Value) lua.LValue {
	tb := &lua.LTable{}
	for _, k := range v.MapKeys() {
		tb.RawSet(ToArbitraryValue(l, k.Interface()),
			ToArbitraryValue(l, v.MapIndex(k).Interface()))
	}
	return tb
}

func ToTableFromSlice(l *lua.LState, v reflect.Value) lua.LValue {
	tb := &lua.LTable{}
	for j := 0; j < v.Len(); j++ {
		tb.RawSet(ToArbitraryValue(l, j+1), // because lua is 1-indexed
			ToArbitraryValue(l, v.Index(j).Interface()))
	}
	return tb
}
