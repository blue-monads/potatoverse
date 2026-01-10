package file

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

func (f *FileOperations) GetFileByRefId(refId string) (*dbmodels.FileMeta, error) {
	file := &dbmodels.FileMeta{}
	err := f.fileMetaTable().Find(db.Cond{"ref_id": refId}).One(file)
	if err != nil {
		return nil, err
	}
	return file, nil
}
