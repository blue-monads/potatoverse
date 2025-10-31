package engine

import (
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

func (e *Engine) serveSimpleRoute(ctx *gin.Context, indexItem *SpaceRouteIndexItem) {

	pp.Println("@indexItem", indexItem)

	filePath := ctx.Param("subpath")

	name, path := buildPackageFilePath(filePath, &indexItem.routeOption)

	pp.Println("@simple_route/name", name)
	pp.Println("@simple_route/path", path)

	pFileOps := e.db.GetPackageFileOps()
	err := pFileOps.StreamFileToHTTP(indexItem.packageVersionId, path, name, ctx.Writer)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

}
