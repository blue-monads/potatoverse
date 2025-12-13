package low

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/enforcer"
	"github.com/blue-monads/turnix/backend/utils/libx/dbutils"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

type Txnish interface {
	Commit() error

	Rollback() error
}

var (
	_ datahub.DBLowOps = (*LowDB)(nil)
)

type LowDB struct {
	sess      db.Session
	ownerType string
	ownerID   string
	isTxn     bool
}

func NewLowDB(sess db.Session, ownerType string, ownerID string) *LowDB {
	return &LowDB{
		sess:      sess,
		ownerType: ownerType,
		ownerID:   ownerID,
		isTxn:     false,
	}
}

// db.NewTx

func (d *LowDB) tableName(table string) string {
	if strings.Contains(table, "__") {
		panic("table name cannot contain '__'")
	}

	tt := enforcer.TableName(d.ownerType, d.ownerID, table)

	// qq.Println("tableName", table, "=>", tt)

	return tt

}

func (d *LowDB) StartTxn() (datahub.DBLowTxnOps, error) {
	if d.isTxn {
		return nil, errors.New("already in a transaction")
	}
	d.isTxn = true

	driver := d.sess.Driver().(*sql.DB)
	tx, err := driver.Begin()
	if err != nil {
		return nil, err
	}
	txn, err := sqlite.NewTx(tx)
	if err != nil {
		return nil, err
	}

	return &LowDB{
		sess:      txn,
		ownerType: d.ownerType,
		ownerID:   d.ownerID,
		isTxn:     true,
	}, nil

}

func (d *LowDB) Commit() error {
	if !d.isTxn {
		return errors.New("not in a transaction")
	}

	txnHandle := d.sess.(Txnish)
	return txnHandle.Commit()
}

func (d *LowDB) Rollback() error {
	if !d.isTxn {
		return errors.New("not in a transaction")
	}
	txnHandle := d.sess.(Txnish)
	return txnHandle.Rollback()
}

func (d *LowDB) RunDDL(ddl string) error {
	fmt.Println("RunDDL/0", ddl)
	driver := d.sess.Driver().(*sql.DB)
	transformedDDL, err := enforcer.TransformQuery(d.ownerType, d.ownerID, ddl)
	if err != nil {
		return err
	}

	fmt.Println("RunDDL/1", transformedDDL)

	_, err = driver.Exec(transformedDDL)
	if err != nil {
		return err
	}
	return err
}

func (d *LowDB) RunQuery(query string, data ...any) ([]map[string]any, error) {
	driver := d.sess.Driver().(*sql.DB)
	transformedQuery, err := enforcer.TransformQuery(d.ownerType, d.ownerID, query)
	if err != nil {
		return nil, err
	}
	rows, err := driver.Query(transformedQuery, data...)
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
	transformedQuery, err := enforcer.TransformQuery(d.ownerType, d.ownerID, query)
	if err != nil {
		return nil, err
	}
	rows, err := driver.Query(transformedQuery, data...)
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

func (d *LowDB) Exec(query string, data ...any) (any, error) {
	driver := d.sess.Driver().(*sql.DB)
	transformedQuery, err := enforcer.TransformQuery(d.ownerType, d.ownerID, query)
	if err != nil {
		return nil, err
	}

	r, err := driver.Exec(transformedQuery, data...)
	return r, err
}

func (d *LowDB) Insert(table string, data map[string]any) (int64, error) {
	collection := d.sess.Collection(d.tableName(table))
	res, err := collection.Insert(data)
	if err != nil {
		return 0, err
	}
	return res.ID().(int64), nil
}

func (d *LowDB) UpdateById(table string, id int64, data map[string]any) error {
	collection := d.sess.Collection(d.tableName(table))
	return collection.Find(db.Cond{"id": id}).Update(data)
}

func (d *LowDB) DeleteById(table string, id int64) error {
	collection := d.sess.Collection(d.tableName(table))
	return collection.Find(db.Cond{"id": id}).Delete()
}

func (d *LowDB) FindById(table string, id int64) (map[string]any, error) {
	collection := d.sess.Collection(d.tableName(table))
	var result map[string]any
	err := collection.Find(db.Cond{"id": id}).One(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *LowDB) UpdateByCond(table string, cond map[any]any, data map[string]any) error {
	collection := d.sess.Collection(d.tableName(table))
	return collection.Find(buildCond(cond)).Update(data)
}

func (d *LowDB) DeleteByCond(table string, cond map[any]any) error {
	collection := d.sess.Collection(d.tableName(table))
	dbCond := db.Cond(cond)
	return collection.Find(dbCond).Delete()
}

func (d *LowDB) FindAllByCond(table string, cond map[any]any) ([]map[string]any, error) {
	collection := d.sess.Collection(d.tableName(table))
	var results []map[string]any
	query := collection.Find(buildCond(cond))
	err := query.All(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (d *LowDB) FindOneByCond(table string, cond map[any]any) (map[string]any, error) {
	collection := d.sess.Collection(d.tableName(table))
	var result map[string]any
	err := collection.Find(buildCond(cond)).One(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *LowDB) FindAllByQuery(query *datahub.FindQuery) ([]map[string]any, error) {
	collection := d.sess.Collection(d.tableName(query.Table))
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

func (d *LowDB) FindByJoin(query *datahub.FindByJoin) ([]map[string]any, error) {
	if len(query.Joins) == 0 {
		return nil, errors.New("no joins provided")
	}

	for i := range query.Joins {
		join := &query.Joins[i]
		if join.LeftAs == "" {
			join.LeftAs = join.LeftTable
		}

		if join.RightAs == "" {
			join.RightAs = join.RightTable
		}

		join.LeftTable = d.tableName(join.LeftTable)
		join.RightTable = d.tableName(join.RightTable)

	}

	firstJoin := query.Joins[0]
	mainTable := firstJoin.LeftTable
	if firstJoin.LeftAs != "" {
		mainTable = mainTable + " AS " + firstJoin.LeftAs
	}

	// Build select query
	var sqlQuery db.Selector
	if len(query.Fields) > 0 {
		// Convert []string to []interface{} for Select variadic
		fields := make([]any, len(query.Fields))
		for i, f := range query.Fields {
			fields[i] = f
		}
		sqlQuery = d.sess.SQL().Select(fields...).From(mainTable)
	} else {
		sqlQuery = d.sess.SQL().Select("*").From(mainTable)
	}

	// Add joins
	for _, join := range query.Joins {
		rightTable := join.RightTable
		if join.RightAs != "" {
			rightTable = rightTable + " AS " + join.RightAs
		}

		// Build ON clause using aliases if provided
		leftTableRef := join.LeftTable
		if join.LeftAs != "" {
			leftTableRef = join.LeftAs
		}
		rightTableRef := join.RightTable
		if join.RightAs != "" {
			rightTableRef = join.RightAs
		}
		onClause := leftTableRef + "." + join.LeftOn + " = " + rightTableRef + "." + join.RightOn

		// Handle different join types
		switch join.JoinType {
		case "LEFT", "LEFT JOIN":
			sqlQuery = sqlQuery.LeftJoin(rightTable)
		case "RIGHT", "RIGHT JOIN":
			sqlQuery = sqlQuery.RightJoin(rightTable)
		case "FULL", "FULL OUTER", "FULL OUTER JOIN":
			sqlQuery = sqlQuery.FullJoin(rightTable)
		case "INNER", "INNER JOIN", "":
			// Default to inner join
			sqlQuery = sqlQuery.Join(rightTable)
		default:
			// Default to inner join for unknown types
			sqlQuery = sqlQuery.Join(rightTable)
		}

		sqlQuery = sqlQuery.On(onClause)
	}

	// Add root condition
	if len(query.Cond) > 0 {
		sqlQuery = sqlQuery.Where(buildCond(query.Cond))
	}

	// Add ordering
	if query.Order != "" {
		sqlQuery = sqlQuery.OrderBy(query.Order)
	}

	// qq.Println(sqlQuery.String())

	// Execute query and get rows
	rows, err := sqlQuery.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dbutils.SelectScan(rows)
}

func (d *LowDB) ListTables() ([]string, error) {

	pattern := enforcer.TableNamePattern(d.ownerType, d.ownerID)

	query := d.sess.SQL().
		Select("name").
		From("sqlite_master").
		Where(db.Cond{
			"type": "table",
			"name": db.Like(pattern),
		})

	rows, err := query.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := dbutils.SelectScan(rows)
	if err != nil {
		return nil, err
	}

	tableNames := make([]string, len(results))
	for i, result := range results {
		tableNames[i] = result["name"].(string)
	}

	return tableNames, nil
}

func (d *LowDB) ListTableColumns(table string) ([]map[string]any, error) {
	finalTableName := d.tableName(table)
	rawquery := fmt.Sprintf(`SELECT * FROM pragma_table_info('%s')`, finalTableName)
	rows, err := d.sess.SQL().Query(rawquery)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return dbutils.SelectScan(rows)
}
