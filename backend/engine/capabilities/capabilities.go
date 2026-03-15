package capabilities

import (

	// xdatabase
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xDatabase/xMigrator"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xDatabase/xSeeder/xAutoSeeder"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xDatabase/xSeeder/xStaticSeeder"

	// xfiles
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xFiles"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xFiles/xFileRelay"

	// xsystem
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xSystem/xCorn"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xSystem/xEngine/xLua"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xSystem/xPing"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xSystem/xSqlite"

	// xuser
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xUser"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xUser/xUgroup"

	// xwebsocket
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xWebsocket"
	_ "github.com/blue-monads/potatoverse/backend/engine/capabilities/xWebsocket/xEasyWS"
)
