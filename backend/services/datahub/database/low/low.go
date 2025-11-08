package low

import (
	"database/sql"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/upper/db/v4"
)

var (
	_ datahub.DBLowOps = (*LowDB)(nil)
)

type LowDB struct {
	sess db.Session
}

func NewLowDB(sess db.Session, ownerType string, ownerID string) *LowDB {
	return &LowDB{
		sess: sess,
	}
}

func (d *LowDB) RunDDL(ddl string) error {
	driver := d.sess.Driver().(*sql.DB)
	_, err := driver.Exec(ddl)
	return err
}

func (d *LowDB) RunQuery(query string, data ...any) ([]map[string]any, error) {
	driver := d.sess.Driver().(*sql.DB)
	rows, err := driver.Query(query, data...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	numColumns := len(columns)

	values := make([]any, numColumns)
	for i := range values {
		values[i] = new(any)
	}

	var results []map[string]any
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}

		dest := make(map[string]any, numColumns)
		for i, column := range columns {
			dest[column] = *(values[i].(*any))
		}
		results = append(results, dest)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (d *LowDB) RunQueryOne(query string, data ...any) (map[string]any, error) {
	driver := d.sess.Driver().(*sql.DB)
	rows, err := driver.Query(query, data...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	numColumns := len(columns)

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	values := make([]any, numColumns)
	for i := range values {
		values[i] = new(any)
	}

	if err := rows.Scan(values...); err != nil {
		return nil, err
	}

	result := make(map[string]any, numColumns)
	for i, column := range columns {
		result[column] = *(values[i].(*any))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *LowDB) Insert(table string, data map[string]any) (int64, error) {
	collection := d.sess.Collection(table)
	res, err := collection.Insert(data)
	if err != nil {
		return 0, err
	}
	return res.ID().(int64), nil
}

func (d *LowDB) UpdateById(table string, id int64, data map[string]any) error {
	collection := d.sess.Collection(table)
	return collection.Find(db.Cond{"id": id}).Update(data)
}

func (d *LowDB) DeleteById(table string, id int64) error {
	collection := d.sess.Collection(table)
	return collection.Find(db.Cond{"id": id}).Delete()
}

func (d *LowDB) FindById(table string, id int64) (map[string]any, error) {
	collection := d.sess.Collection(table)
	var result map[string]any
	err := collection.Find(db.Cond{"id": id}).One(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *LowDB) UpdateByCond(table string, cond map[any]any, data map[string]any) error {
	collection := d.sess.Collection(table)
	return collection.Find(buildCond(cond)).Update(data)
}

func (d *LowDB) DeleteByCond(table string, cond map[any]any) error {
	collection := d.sess.Collection(table)
	dbCond := db.Cond(cond)
	return collection.Find(dbCond).Delete()
}

func (d *LowDB) FindAllByCond(table string, cond map[any]any) ([]map[string]any, error) {
	collection := d.sess.Collection(table)
	var results []map[string]any
	query := collection.Find(buildCond(cond))
	err := query.All(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (d *LowDB) FindOneByCond(table string, cond map[any]any) (map[string]any, error) {
	collection := d.sess.Collection(table)
	var result map[string]any
	err := collection.Find(buildCond(cond)).One(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *LowDB) FindAllByQuery(query *datahub.FindQuery) ([]map[string]any, error) {
	collection := d.sess.Collection(query.Table)
	var results []map[string]any
	queryf := collection.Find(buildCond(query.Cond))
	if query.Offset > 0 {
		queryf = queryf.Offset(query.Offset)
	}
	if query.Limit > 0 {
		queryf = queryf.Limit(query.Limit)
	}
	if query.Order != "" {
		queryf = queryf.OrderBy(query.Order)
	}

	if len(query.Fields) > 0 {
		queryf = queryf.Select(query.Fields)
	}
	err := queryf.All(&results)
	if err != nil {
		return nil, err
	}
	return results, nil

}

func (d *LowDB) FindByQuerySQL(query *datahub.FindByQuerySQL) ([]map[string]any, error) {
	sqlQuery := d.sess.SQL().Select("*").From(query.Table)

	// Add joins
	for _, join := range query.Joins {
		joinTable := join.Table
		if join.As != "" {
			joinTable = joinTable + " AS " + join.As
		}

		// Handle different join types
		switch join.JoinType {
		case "LEFT", "LEFT JOIN":
			sqlQuery = sqlQuery.LeftJoin(joinTable)
		case "RIGHT", "RIGHT JOIN":
			sqlQuery = sqlQuery.RightJoin(joinTable)
		case "FULL", "FULL OUTER", "FULL OUTER JOIN":
			sqlQuery = sqlQuery.FullJoin(joinTable)
		case "INNER", "INNER JOIN", "":
			// Default to inner join
			sqlQuery = sqlQuery.Join(joinTable)
		default:
			// Default to inner join for unknown types
			sqlQuery = sqlQuery.Join(joinTable)
		}

		if join.On != "" {
			sqlQuery = sqlQuery.On(join.On)
		}
	}

	// Add root condition
	if len(query.Cond) > 0 {
		sqlQuery = sqlQuery.Where(buildCond(query.Cond))
	}

	// Execute query and get rows
	rows, err := sqlQuery.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	numColumns := len(columns)

	values := make([]any, numColumns)
	for i := range values {
		values[i] = new(any)
	}

	var results []map[string]any
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}

		dest := make(map[string]any, numColumns)
		for i, column := range columns {
			dest[column] = *(values[i].(*any))
		}
		results = append(results, dest)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
