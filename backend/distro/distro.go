package distro

import (
	// Capabilities
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities"

	// Lua Executor
	_ "github.com/blue-monads/potatoverse/backend/engine/executors/luaz"
	_ "github.com/blue-monads/potatoverse/backend/engine/executors/workerd"

	// Repo Hub
	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/devrepo"
	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/providers/harvester"

	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/targets"
)
