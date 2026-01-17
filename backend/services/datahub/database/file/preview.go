package file

import "github.com/upper/db/v4"

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
