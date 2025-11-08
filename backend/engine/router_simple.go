package engine

import (
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/gin-gonic/gin"
)

func (e *Engine) serveSimpleRoute(ctx *gin.Context, indexItem *SpaceRouteIndexItem) {

	qq.Println("@indexItem", indexItem)

	filePath := ctx.Param("subpath")

	name, path := buildPackageFilePath(filePath, &indexItem.routeOption)

	qq.Println("@simple_route/name", name)
	qq.Println("@simple_route/path", path)

	pFileOps := e.db.GetPackageFileOps()
	err := pFileOps.StreamFileToHTTP(indexItem.packageVersionId, path, name, ctx.Writer)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

}
