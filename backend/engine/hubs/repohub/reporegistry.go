package repohub

import (
	"sync"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/repotypes"
)

var (
	repoProviders      map[string]repotypes.RepoProvider
	repoProvidersMutex sync.RWMutex
)

func RegisterRepoProvider(name string, provider repotypes.RepoProvider) {
	repoProvidersMutex.Lock()
	defer repoProvidersMutex.Unlock()
	repoProviders[name] = provider
}
