package executors

import (
	"log/slog"
	"os"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type EHandle struct {
	Logger      *slog.Logger
	App         xtypes.App
	FsRoot      *os.Root
	RootSpaceId int64
	SpaceId     int64
	PackageId   int64
	Database    datahub.Database
}

type PresignedOptions struct {
	Uid      int64  `json:"uid,omitempty"`
	Path     string `json:"path,omitempty"`
	FileName string `json:"file_name,omitempty"`
	Expiry   int64  `json:"expiry,omitempty"`
}
