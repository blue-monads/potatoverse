package erepo

import (
	"embed"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/repotypes"
)

var (
	_ repotypes.IRepo = (*EmbedRepo)(nil)
)

type EmbedRepo struct {
	fs embed.FS
}

func NewEmbedRepo(fs embed.FS) *EmbedRepo {
	return &EmbedRepo{fs: fs}
}

func (r *EmbedRepo) ListPackages() ([]repotypes.PotatoPackage, error) {
	return listEmbeddedPackagesFromFS(r.fs)
}

func (r *EmbedRepo) ZipPackage(packageName string, version string) (string, error) {
	return zipEmbeddedPackageFromFS(r.fs, packageName)
}
