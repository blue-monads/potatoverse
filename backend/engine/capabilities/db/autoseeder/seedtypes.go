package autoseeder

import (
	_ "github.com/brianvoe/gofakeit/v7"
)

type SeedStruct struct {
	TableName string
	Columns   []SeedColumn
	Rows      int
}

type SeedColumn struct {
	ColumnName      string
	DataType        string   // this should correspond to gofakeit data type
	NotNull         bool     // means randomly generate null for some column
	NullProbability float64  // probability of generating null for the column
	DefaultValues   []string // default values for the column, mostly for enum type
	RangeMin        int      // minimum value for number type and date unix timestamp
	RangeMax        int      // maximum value for number type and date unix timestamp
}
