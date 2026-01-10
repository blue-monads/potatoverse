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

type FilePreviewResult struct {
	Preview []byte `db:"preview"`
}

func (f *FileOperations) GetFilePreview(ownerID int64, id int64) ([]byte, error) {
	file := &FilePreviewResult{}
	err := f.fileMetaTable().Find(db.Cond{
		"owner_id": ownerID,
		"id":       id,
	}).Select("preview").One(file)

	if err != nil {
		return nil, err
	}
	return file.Preview, nil
}
