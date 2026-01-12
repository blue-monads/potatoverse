package mesh

import "github.com/gin-gonic/gin"

func RegisterMeshProvider(meshId string, mesh Mesh) {

}

type Mesh interface {
	Route(nodeId string, ctx *gin.Context)
}
