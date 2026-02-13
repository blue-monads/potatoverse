package distro

import (
	// Capabilities
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xMigrator"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xPing"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xSeeder/xAutoSeeder"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xSeeder/xStaticSeeder"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xSystem/xCorn"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xWebsocket/xEasyWS"

	// Lua Executor
	_ "github.com/blue-monads/potatoverse/backend/engine/executors/luaz"

	// Repo Hub
	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/devrepo"
	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/providers/harvester"

	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/targets"
)
