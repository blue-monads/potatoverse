package server

import (
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

func NewMain() {

	g := gin.Default()

	g.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	g.GET("/bob/:which", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"which": c.Param("which"),
		})
	})

	routes := g.Routes()
	for _, route := range routes {
		pp.Println(route.Method, route.Path, route.Handler)

	}

}

type HTTPSpecServer struct {
	handlers map[string]*HttpRoute
}

type HttpRoute struct {
	Name     string
	Path     string
	Info     string
	AuthName string
}

/*

OLD API ROUTES FOR REFERENCE


[GIN-debug] POST   /z/auth/signup/direct     --> github.com/blue-monads/turnix/backend/app/server.(*Server).signUpDirect-fm (4 handlers)
[GIN-debug] POST   /z/auth/signup/invite     --> github.com/blue-monads/turnix/backend/app/server.(*Server).signUpInvite-fm (4 handlers)
[GIN-debug] POST   /z/auth/login             --> github.com/blue-monads/turnix/backend/app/server.(*Server).login-fm (4 handlers)
[GIN-debug] GET    /z/api/v1/self            --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func1 (4 handlers)
[GIN-debug] GET    /z/api/v1/self/change_password --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func2 (4 handlers)
[GIN-debug] GET    /z/api/v1/self/users      --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func3 (4 handlers)
[GIN-debug] POST   /z/api/v1/self/users      --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func4 (4 handlers)
[GIN-debug] GET    /z/api/v1/self/users/:uid --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func5 (4 handlers)
[GIN-debug] POST   /z/api/v1/self/users/:uid --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func6 (4 handlers)
[GIN-debug] DELETE /z/api/v1/self/users/:uid --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func7 (4 handlers)
[GIN-debug] GET    /z/api/v1/self/self       --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func8 (4 handlers)
[GIN-debug] GET    /z/api/v1/self/messages   --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func9 (4 handlers)
[GIN-debug] GET    /z/api/v1/self/files      --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func10 (4 handlers)
[GIN-debug] POST   /z/api/v1/self/files      --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func11 (4 handlers)
[GIN-debug] PUT    /z/api/v1/self/files      --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func12 (4 handlers)
[GIN-debug] GET    /z/api/v1/self/files/:id  --> github.com/blue-monads/turnix/backend/app/server.(*Server).getSelfFile-fm (4 handlers)
[GIN-debug] DELETE /z/api/v1/self/files/:id  --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func13 (4 handlers)
[GIN-debug] POST   /z/api/v1/self/messages/:uid --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func14 (4 handlers)
[GIN-debug] GET    /z/api/v1/self/files/:id/shares --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func15 (4 handlers)
[GIN-debug] POST   /z/api/v1/self/files/:id/shares --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func16 (4 handlers)
[GIN-debug] DELETE /z/api/v1/self/files/:id/shares/:id --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func17 (4 handlers)
[GIN-debug] GET    /z/api/v1/file/shared/:file --> github.com/blue-monads/turnix/backend/app/server.(*Server).getSharedFile-fm (4 handlers)
[GIN-debug] POST   /z/api/v1/file/shared/:file --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func18 (4 handlers)
[GIN-debug] DELETE /z/api/v1/file/shared/:file --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func19 (4 handlers)
[GIN-debug] GET    /z/api/v1/file/shortKey/:shortkey --> github.com/blue-monads/turnix/backend/app/server.(*Server).GetFileWithShortKey-fm (4 handlers)
[GIN-debug] POST   /z/api/v1/file/:fid/shortkey --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func20 (4 handlers)
[GIN-debug] GET    /z/api/v1/user/:uid       --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func21 (4 handlers)
[GIN-debug] GET    /z/api/v1/project         --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func22 (4 handlers)
[GIN-debug] POST   /z/api/v1/project         --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func23 (4 handlers)
[GIN-debug] POST   /z/api/v1/project/:pid    --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func24 (4 handlers)
[GIN-debug] GET    /z/api/v1/project/:pid    --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func25 (4 handlers)
[GIN-debug] DELETE /z/api/v1/project/:pid    --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func26 (4 handlers)
[GIN-debug] POST   /z/api/v1/project/:pid/user --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func27 (4 handlers)
[GIN-debug] DELETE /z/api/v1/project/:pid/user --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func28 (4 handlers)
[GIN-debug] POST   /z/api/v1/project_type_install --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func29 (4 handlers)
[GIN-debug] GET    /z/api/v1/project_types   --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func30 (4 handlers)
[GIN-debug] GET    /z/api/v1/project_types/:ptype/form --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func31 (4 handlers)
[GIN-debug] GET    /z/api/v1/project_types/:ptype/reload --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func32 (4 handlers)
[GIN-debug] GET    /z/api/v1/project_types/:ptype --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func33 (4 handlers)
[GIN-debug] GET    /z/api/v1/project/:pid/hook --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func34 (4 handlers)
[GIN-debug] POST   /z/api/v1/project/:pid/hook --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func35 (4 handlers)
[GIN-debug] GET    /z/api/v1/project/:pid/hook/:id --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func36 (4 handlers)
[GIN-debug] POST   /z/api/v1/project/:pid/hook/:id --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func37 (4 handlers)
[GIN-debug] DELETE /z/api/v1/project/:pid/hook/:id --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func38 (4 handlers)
[GIN-debug] GET    /z/api/v1/project/:pid/files --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func39 (4 handlers)
[GIN-debug] POST   /z/api/v1/project/:pid/files --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func40 (4 handlers)
[GIN-debug] PUT    /z/api/v1/project/:pid/files --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func41 (4 handlers)
[GIN-debug] GET    /z/api/v1/project/:pid/files/:id --> github.com/blue-monads/turnix/backend/app/server.(*Server).getProjectFile-fm (4 handlers)
[GIN-debug] DELETE /z/api/v1/project/:pid/files/:id --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func42 (4 handlers)
[GIN-debug] POST   /z/api/v1/project/:pid/sqlexec --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func43 (4 handlers)
[GIN-debug] POST   /z/api/v1/project/:pid/sqlexec2 --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func44 (4 handlers)
[GIN-debug] GET    /z/api/v1/project/:pid/tables --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func45 (4 handlers)
[GIN-debug] GET    /z/api/v1/project/:pid/tables/:table/columns --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func46 (4 handlers)
[GIN-debug] POST   /z/api/v1/project/:pid/autoquery --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func47 (4 handlers)
[GIN-debug] GET    /z/api/v1/project/:pid/plugins --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func48 (4 handlers)
[GIN-debug] POST   /z/api/v1/project/:pid/plugins --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func49 (4 handlers)
[GIN-debug] DELETE /z/api/v1/project/:pid/plugins/:id --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func50 (4 handlers)
[GIN-debug] POST   /z/api/v1/project/:pid/plugins/:id --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func51 (4 handlers)
[GIN-debug] GET    /z/api/v1/project/:pid/plugins/:id --> github.com/blue-monads/turnix/backend/app/server.(*Server).apiRoutes.(*Server).accessMiddleware.func52 (4 handlers)
"@using_dev_proxy" "http://localhost:5173"
[GIN-debug] GET    /z/pages                  --> github.com/blue-monads/turnix/backend/app/server/assets.PagesRoutesServer.func1 (4 handlers)
[GIN-debug] GET    /z/pages/*files           --> github.com/blue-monads/turnix/backend/app/server/assets.PagesRoutesServer.func1 (4 handlers)
[GIN-debug] GET    /z/x/:pname/*files        --> github.com/blue-monads/turnix/backend/app/server.(*Server).externalAssets.func1 (4 handlers)
[GIN-debug] GET    /z/x/:pname               --> github.com/blue-monads/turnix/backend/app/server.(*Server).externalAssets.func1 (4 handlers)
[GIN-debug] GET    /z/lib/*file              --> github.com/blue-monads/turnix/backend/app/server.(*Server).pages.func1 (4 handlers)
[GIN-debug] GET    /z/global.js              --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func1 (4 handlers)
[GIN-debug] GET    /z/projects/:ptype        --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func2 (4 handlers)
[GIN-debug] POST   /z/projects/:ptype        --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func2 (4 handlers)
[GIN-debug] PUT    /z/projects/:ptype        --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func2 (4 handlers)
[GIN-debug] PATCH  /z/projects/:ptype        --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func2 (4 handlers)
[GIN-debug] HEAD   /z/projects/:ptype        --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func2 (4 handlers)
[GIN-debug] OPTIONS /z/projects/:ptype        --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func2 (4 handlers)
[GIN-debug] DELETE /z/projects/:ptype        --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func2 (4 handlers)
[GIN-debug] CONNECT /z/projects/:ptype        --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func2 (4 handlers)
[GIN-debug] TRACE  /z/projects/:ptype        --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func2 (4 handlers)
[GIN-debug] GET    /z/projects/:ptype/*file  --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func3 (4 handlers)
[GIN-debug] POST   /z/projects/:ptype/*file  --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func3 (4 handlers)
[GIN-debug] PUT    /z/projects/:ptype/*file  --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func3 (4 handlers)
[GIN-debug] PATCH  /z/projects/:ptype/*file  --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func3 (4 handlers)
[GIN-debug] HEAD   /z/projects/:ptype/*file  --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func3 (4 handlers)
[GIN-debug] OPTIONS /z/projects/:ptype/*file  --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func3 (4 handlers)
[GIN-debug] DELETE /z/projects/:ptype/*file  --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func3 (4 handlers)
[GIN-debug] CONNECT /z/projects/:ptype/*file  --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func3 (4 handlers)
[GIN-debug] TRACE  /z/projects/:ptype/*file  --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func3 (4 handlers)
[GIN-debug] GET    /z/p/:ptype               --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func4 (4 handlers)
[GIN-debug] POST   /z/p/:ptype               --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func4 (4 handlers)
[GIN-debug] PUT    /z/p/:ptype               --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func4 (4 handlers)
[GIN-debug] PATCH  /z/p/:ptype               --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func4 (4 handlers)
[GIN-debug] HEAD   /z/p/:ptype               --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func4 (4 handlers)
[GIN-debug] OPTIONS /z/p/:ptype               --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func4 (4 handlers)
[GIN-debug] DELETE /z/p/:ptype               --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func4 (4 handlers)
[GIN-debug] CONNECT /z/p/:ptype               --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func4 (4 handlers)
[GIN-debug] TRACE  /z/p/:ptype               --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func4 (4 handlers)
[GIN-debug] GET    /z/p/:ptype/*file         --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func5 (4 handlers)
[GIN-debug] POST   /z/p/:ptype/*file         --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func5 (4 handlers)
[GIN-debug] PUT    /z/p/:ptype/*file         --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func5 (4 handlers)
[GIN-debug] PATCH  /z/p/:ptype/*file         --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func5 (4 handlers)
[GIN-debug] HEAD   /z/p/:ptype/*file         --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func5 (4 handlers)
[GIN-debug] OPTIONS /z/p/:ptype/*file         --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func5 (4 handlers)
[GIN-debug] DELETE /z/p/:ptype/*file         --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func5 (4 handlers)
[GIN-debug] CONNECT /z/p/:ptype/*file         --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func5 (4 handlers)
[GIN-debug] TRACE  /z/p/:ptype/*file         --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func5 (4 handlers)
[GIN-debug] GET    /ping                     --> github.com/blue-monads/turnix/backend/app/server.(*Server).ping-fm (4 handlers)

*/
