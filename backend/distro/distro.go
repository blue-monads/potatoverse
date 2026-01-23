package distro

import (
	// Capabilities
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/db/autoseeder"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/db/migrator"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/db/staticseeder"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/system/ping"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/websocket/easyws"

	// Lua Executor
	_ "github.com/blue-monads/potatoverse/backend/engine/executors/luaz"

	// Repo Hub
	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/devrepo"
	_ "github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/providers/harvester"
)
