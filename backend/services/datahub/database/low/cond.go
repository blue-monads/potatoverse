package low

import (
	"github.com/upper/db/v4"
)

var EmptyCond = db.Cond{}

func buildCond(cond map[any]any) any {
	if len(cond) == 0 {
		return EmptyCond
	}

	nestedAnd, ok := cond["AND"]
	if ok {
		return transformNestedCond(nestedAnd, 0, true) // true = AND
	}

	nestedOr, ok := cond["OR"]
	if ok {
		return transformNestedCond(nestedOr, 0, false) // false = OR
	}

	return db.Cond(cond)

}

func transformNestedCond(nested any, depth int, isAnd bool) db.LogicalExpr {
	if depth > 10 {
		panic("depth limit reached")
	}

	casted := nested.([]any)
	transformed := make([]db.LogicalExpr, 0, len(casted))
	for _, nested := range casted {
		subCond, ok := nested.(map[any]any)
		if !ok {
			panic("nested condition is not a map")
		}

		// Handle nested AND conditions
		nestedAnd, ok := subCond["AND"]
		if ok {
			transformed = append(transformed, transformNestedCond(nestedAnd, depth+1, true))
		}

		// Handle nested OR conditions
		nestedOr, ok := subCond["OR"]
		if ok {
			transformed = append(transformed, transformNestedCond(nestedOr, depth+1, false))
		}

		// Add regular conditions, but skip "AND" and "OR" keys
		for k, v := range subCond {
			if k != "AND" && k != "OR" {
				transformed = append(transformed, db.Cond{k: v})
			}
		}
	}

	if isAnd {
		return db.And(transformed...)
	}
	return db.Or(transformed...)
}
