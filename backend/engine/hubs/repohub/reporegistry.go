package repohub

import (
	"sync"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/repotypes"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

var (
	repoProviders      map[string]repotypes.RepoProvider = make(map[string]repotypes.RepoProvider)
	repoProvidersMutex sync.RWMutex
)

func RegisterRepoProvider(name string, provider repotypes.RepoProvider) {
	repoProvidersMutex.Lock()
	defer repoProvidersMutex.Unlock()
	repoProviders[name] = provider
}

var Default = []xtypes.RepoOptions{
	{
		Name: "Official Potato Field",
		Type: "harvester-v1",
		Slug: "Official",
		URL:  "https://github.com/blue-monads/store/raw/refs/heads/master",
	},

	{
		Name: "Third Party Potato Field",
		Type: "harvester-v1",
		Slug: "ThirdParty",
		URL:  "https://github.com/blue-monads/store-thirdparty/raw/refs/heads/master",
	},

	{
		Name: "Development Packages",
		Type: "dev",
		Slug: "Dev",
	},
}
