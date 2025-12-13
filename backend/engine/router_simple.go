package engine

import (
	"net/http"

	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/gin-gonic/gin"
)

func (e *Engine) serveSimpleRoute(ctx *gin.Context, indexItem *SpaceRouteIndexItem) {
	qq.Println("@indexItem", indexItem)

	filePath := ctx.Param("subpath")

	e.processSimpleRoute(ctx, filePath, indexItem)
}

func (e *Engine) processSimpleRoute(ctx *gin.Context, filePath string, indexItem *SpaceRouteIndexItem) {

	name, path := buildPackageFilePath(filePath, &indexItem.routeOption)

	qq.Println("@simple_route/name", name)
	qq.Println("@simple_route/path", path)

	pFileOps := e.db.GetPackageFileOps()
	err := pFileOps.StreamFileToHTTP(indexItem.packageVersionId, path, name, ctx)
	if err != nil {
		if !e.db.IsEmptyRowsError(err) {
			httpx.WriteErr(ctx, err)
			return
		}

		nofoundFile := indexItem.routeOption.OnNotFoundFile
		serveFolder := indexItem.routeOption.ServeFolder

		qq.Println("@nofoundFile", nofoundFile)
		qq.Println("@serveFolder", serveFolder)

		if nofoundFile == "" {
			ctx.Status(http.StatusNotFound)
			ctx.Writer.Write([]byte("File not found"))
			return
		}

		err = pFileOps.StreamFileToHTTP(indexItem.packageVersionId, serveFolder, nofoundFile, ctx)

		qq.Println("@finish", err)

	}

}
