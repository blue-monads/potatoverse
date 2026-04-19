package actions

import (
	"regexp"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
)

var tablePattern = regexp.MustCompile(`zz_P__\d+__`)

func (c *Controller) ListSpaceDataTables(installId int64) ([]datahub.TableInfo, error) {
	tables, err := c.database.GetLowPackageDBOps(installId).ListTables()
	if err != nil {
		return nil, err
	}

	resultTables := make([]datahub.TableInfo, 0, len(tables))

	for i := range tables {
		table := &tables[i]

		sanName := tablePattern.ReplaceAllString(table.Name, "")

		resultTables = append(resultTables, datahub.TableInfo{
			Name:      sanName,
			TableType: table.TableType,
			Schema:    strings.ReplaceAll(table.Schema, table.Name, sanName),
		})
	}

	return resultTables, nil
}

func (c *Controller) GetSpaceDataTable(installId int64, tableName string) ([]datahub.TableColumnInfo, error) {
	return c.database.GetLowPackageDBOps(installId).ListTableColumns(tableName)
}

func (c *Controller) QuerySpaceDataTable(installId int64, table string, offset int, limit int) ([]map[string]any, error) {
	db := c.database.GetLowPackageDBOps(installId)

	data, err := db.FindAllByQuery(&datahub.FindQuery{
		Table:  table,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}
