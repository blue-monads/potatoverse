package file

import (
	"time"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/gin-gonic/gin"
	nanoid "github.com/jaevor/go-nanoid"
	"github.com/upper/db/v4"
)

var idgen, _ = nanoid.ASCII(10)

func (f *FileOperations) AddFileShare(ownerID int64, fileId int64, userId int64) (string, error) {

	now := time.Now()
	id := idgen()
	_, err := f.fileShareTable().Insert(dbmodels.FileShare{
		ID:        id,
		FileID:    fileId,
		UserID:    userId,
		CreatedAt: &now,
	})

	if err != nil {
		return "", err
	}

	return id, nil
}
func (f *FileOperations) GetSharedFile(ownerID int64, id string, ctx *gin.Context) error {
	fileShare := &dbmodels.FileShare{}
	err := f.fileShareTable().Find(db.Cond{
		"id": id,
	}).One(fileShare)
	if err != nil {
		return err
	}

	file, err := f.GetFileMeta(fileShare.FileID)
	if err != nil {
		return err
	}

	err = f.validateFileOwnership(file, ownerID)
	if err != nil {
		return err
	}

	isCached := f.setupHTTPHeaders(ctx, file)
	if isCached {
		return nil
	}

	return f.streamFileByMeta(file, ctx.Writer)
}

func (f *FileOperations) ListFileShares(ownerID int64, fileId int64) ([]dbmodels.FileShare, error) {
	shares := make([]dbmodels.FileShare, 0)
	err := f.fileShareTable().Find(db.Cond{
		"file_id": fileId,
	}).All(&shares)
	if err != nil {
		return nil, err
	}
	return shares, nil
}

func (f *FileOperations) RemoveFileShare(ownerID int64, userId int64, id string) error {
	return f.fileShareTable().Find(db.Cond{
		"id":      id,
		"user_id": userId,
	}).Delete()
}
