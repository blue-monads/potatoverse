package engine

import "github.com/gin-gonic/gin"

type RootRouter struct {
	engine *Engine
}

func NewRootRouter(engine *Engine) *RootRouter {
	return &RootRouter{
		engine: engine,
	}
}

func (r *RootRouter) Serve(ctx *gin.Context) {
	// TODO: implement root router (later feature)
	// /* (all except /zz path) -> root router
}
