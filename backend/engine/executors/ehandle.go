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

func (c *EHandle) GetSpaceFilePresigned(uid int64, path string, fileName string, expiry int64) (string, error) {
	// Get signer from app
	signerInstance := c.App.Signer()
	if signerInstance == nil {
		return "", fmt.Errorf("signer not available")
	}

	// Default expiry to 1 hour if not specified
	if expiry == 0 {
		expiry = 3600
	}

	// Create presigned claim
	claim := &signer.SpaceFilePresignedClaim{
		SpaceId:  c.RootSpaceId,
		UserId:   uid,
		PathName: path,
		FileName: fileName,
		Expiry:   expiry,
	}

	// Sign and return the token
	token, err := signerInstance.SignSpaceFilePresigned(claim)
	if err != nil {
		return "", fmt.Errorf("failed to sign presigned token: %w", err)
	}

	return token, nil
}
