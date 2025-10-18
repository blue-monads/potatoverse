package executors

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/signer"
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

func (c *EHandle) GetSpaceFilePresigned(opts PresignedOptions) (string, error) {
	signerInstance := c.App.Signer()

	if opts.Expiry == 0 {
		opts.Expiry = 3600
	}

	backend, cleanPath := backend(opts.Path)
	spaceId := c.RootSpaceId
	if backend == "private" {
		spaceId = c.SpaceId
	}

	if cleanPath == "/" || cleanPath == "." {
		cleanPath = ""
	}

	claim := &signer.SpaceFilePresignedClaim{
		SpaceId:  spaceId,
		UserId:   opts.Uid,
		PathName: cleanPath,
		FileName: opts.FileName,
		Expiry:   opts.Expiry,
	}

	token, err := signerInstance.SignSpaceFilePresigned(claim)
	if err != nil {
		return "", fmt.Errorf("failed to sign presigned token: %w", err)
	}

	return token, nil
}
