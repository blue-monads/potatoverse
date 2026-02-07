package database

import (
	"strconv"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database/low"
)

func (db *DB) GetGlobalOps() datahub.GlobalOps {
	return db.globalOps
}

func (db *DB) GetUserOps() datahub.UserOps {
	return db.userOps
}

func (db *DB) GetSpaceOps() datahub.SpaceOps {
	return db.spaceOps
}

func (db *DB) GetSpaceKVOps() datahub.SpaceKVOps {
	return db.spaceOps
}

func (db *DB) GetPackageInstallOps() datahub.PackageInstallOps {
	return db.packageInstallOps
}

func (db *DB) GetFileOps() datahub.FileOps {
	return db.fileOps
}

func (db *DB) GetPackageFileOps() datahub.FileOps {
	return db.packageFileOps
}

func (db *DB) GetLowDBOps(ownerType string, ownerID string) datahub.DBLowOps {
	return low.NewLowDB(db.sess, ownerType, ownerID)
}

func (db *DB) GetLowPackageDBOps(installId int64) datahub.DBLowOps {
	return low.NewLowDB(db.sess, "P", strconv.FormatInt(installId, 10))
}

func (db *DB) GetLowCapabilityDBOps(capabilityId int64) datahub.DBLowOps {
	return low.NewLowDB(db.sess, "C", strconv.FormatInt(capabilityId, 10))
}

func (db *DB) GetMQSynk() datahub.MQSynk {
	return db.eventOps
}
