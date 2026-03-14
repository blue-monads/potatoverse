package actions

import "github.com/blue-monads/potatoverse/backend/services/datahub"

func (c *Controller) ListSpaceDataTables(installId int64) ([]datahub.TableInfo, error) {
	return c.database.GetLowPackageDBOps(installId).ListTables()
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
