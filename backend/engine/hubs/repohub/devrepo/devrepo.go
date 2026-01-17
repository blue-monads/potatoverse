package devrepo

import (
	"embed"

	repohub "github.com/blue-monads/potatoverse/backend/engine/hubs/repohub"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/providers/erepo"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/repotypes"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

//go:embed all:epackages/*
var embedPackages embed.FS

func init() {
	repohub.RegisterRepoProvider("dev", func(app xtypes.App, repoOptions *xtypes.RepoOptions) (repotypes.IRepo, error) {
		return erepo.NewEmbedRepo(embedPackages), nil
	})
}
