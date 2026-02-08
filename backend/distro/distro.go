package distro

import (
	// Capabilities
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xAutoSeeder"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xCorn"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xEasyWS"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xMigrator"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xPing"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xStaticSeeder"

	// Lua Executor
	_ "github.com/blue-monads/potatoverse/backend/engine/executors/luaz"

	// Repo Hub
	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/devrepo"
	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/providers/harvester"

	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/targets"
)

/*
backend/engine/capabilities/xAutoSeeder
backend/engine/capabilities/xCorn
backend/engine/capabilities/xCtxUser
backend/engine/capabilities/xEasyWS
backend/engine/capabilities/xFileRelay
backend/engine/capabilities/xFiles
backend/engine/capabilities/xMigrator
backend/engine/capabilities/xPing
backend/engine/capabilities/xSelfPkg
backend/engine/capabilities/xStaticSeeder
backend/engine/capabilities/xTemplate
backend/engine/capabilities/xUser

*/
