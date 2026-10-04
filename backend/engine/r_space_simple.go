package engine

import (
	"net/http"

	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/gin-gonic/gin"
)

func (r *SpaceRouter) serveSimpleRoute(ctx *gin.Context, indexItem *SpaceRouteIndexItem) {
	qq.Println("@indexItem", indexItem)

	filePath := ctx.Param("subpath")
	if filePath == "" {
		filePath = ctx.Request.URL.Path
	}

	r.processSimpleRoute(ctx, filePath, indexItem)
}

func (r *SpaceRouter) processSimpleRoute(ctx *gin.Context, filePath string, indexItem *SpaceRouteIndexItem) {

	name, path := buildPackageFilePath(filePath, &indexItem.routeOption)

	qq.Println("@simple_route/name", name)
	qq.Println("@simple_route/path", path)

	pFileOps := r.engine.db.GetPackageFileOps()
	err := pFileOps.StreamFileToHTTP(indexItem.packageVersionId, path, name, ctx)
	if err != nil {
		if !r.engine.db.IsEmptyRowsError(err) {
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
