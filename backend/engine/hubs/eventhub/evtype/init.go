package evtype

import (
	"maps"
	"sync"
)

var (
	targetBuilders = make(map[string]Builder)

	targetBuildersLock = sync.RWMutex{}
)

func RegisterTargetBuilder(name string, builder Builder) {
	targetBuildersLock.Lock()
	defer targetBuildersLock.Unlock()
	targetBuilders[name] = builder
}

func GetTargetBuilders() map[string]Builder {
	targetBuildersLock.RLock()
	defer targetBuildersLock.RUnlock()
	return maps.Clone(targetBuilders)
}
