package distro

import (
	// Capabilities
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xFiles"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xDatabase/xMigrator"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xDatabase/xSeeder/xAutoSeeder"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xDatabase/xSeeder/xStaticSeeder"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xSystem/xCorn"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xSystem/xPing"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xSqlite"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xWebsocket/xEasyWS"

	// Lua Executor
	_ "github.com/blue-monads/potatoverse/backend/engine/executors/luaz"

	// Repo Hub
	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/devrepo"
	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/providers/harvester"

	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/targets"
)
