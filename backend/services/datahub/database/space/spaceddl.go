package space

import "github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"

func (d *SpaceOperations) ListSpaceTables(spaceId int64) ([]string, error) {
	return nil, nil
}

func (d *SpaceOperations) ListSpaceTableColumns(spaceId int64, table string) ([]dbmodels.SpaceTableColumn, error) {
	return nil, nil
}

func (d *SpaceOperations) RunSpaceSQLQuery(spaceId int64, query string, data []any) ([]map[string]any, error) {
	return nil, nil
}

func (d *SpaceOperations) RunSpaceDDL(spaceId int64, ddl string) error {
	return nil
}
