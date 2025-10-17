package engine

import (
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

func (e *Engine) serveSimpleRoute(ctx *gin.Context, indexItem *SpaceRouteIndexItem) {

	pp.Println("@indexItem", indexItem)

	filePath := ctx.Param("files")

	name, path := buildPackageFilePath(filePath, &indexItem.routeOption)

	pp.Println("@simple_route/name", name)
	pp.Println("@simple_route/path", path)

	err := e.db.GetPackageFileStreamingByPath(indexItem.packageId, path, name, ctx.Writer)
	if err != nil {
		ctx.JSON(404, gin.H{"error": "file not found"})
		return
	}

}
