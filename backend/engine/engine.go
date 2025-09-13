package engine

import (
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/models"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

type indexItem struct {
	packageId int64
	spaceId   int64
}

type Engine struct {
	db           datahub.Database
	RoutingIndex map[string]indexItem
	riLock       sync.RWMutex
}

func NewEngine(db datahub.Database) *Engine {
	return &Engine{
		db:           db,
		RoutingIndex: make(map[string]indexItem),
	}
}

func (e *Engine) LoadRoutingIndex() error {

	nextRoutingIndex := make(map[string]indexItem)

	spaces, err := e.db.ListSpaces()
	if err != nil {
		return err
	}

	for _, space := range spaces {
		if space.OwnsNamespace {
			nextRoutingIndex[space.NamespaceKey] = indexItem{
				packageId: space.PackageID,
				spaceId:   space.ID,
			}
		}
	}

	e.riLock.Lock()
	e.RoutingIndex = nextRoutingIndex
	e.riLock.Unlock()

	return nil
}

func (e *Engine) ServeSpaceFile(ctx *gin.Context) {

	spaceKey := ctx.Param("space_key")

	e.riLock.RLock()
	ri, ok := e.RoutingIndex[spaceKey]
	e.riLock.RUnlock()
	if !ok {
		ctx.JSON(404, gin.H{"error": "space not found"})
		return
	}

	filePath := ctx.Param("files")

	nameParts := strings.Split(filePath, "/")
	name := nameParts[len(nameParts)-1]
	path := strings.Join(nameParts[:len(nameParts)-1], "/")

	fmeta, err := e.db.GetPackageFileMetaByPath(ri.packageId, path, name)
	if err != nil {
		ctx.JSON(404, gin.H{"error": "file not found"})
		return
	}

	e.db.GetPackageFileStreaming(ri.packageId, fmeta.ID, ctx.Writer)

}

func (e *Engine) ServePluginFile(ctx *gin.Context) {

}

func (e *Engine) SpaceApi(ctx *gin.Context) {

}

func (e *Engine) PluginApi(ctx *gin.Context) {

}

func (e *Engine) InstallPackageByUrl(url string) (int64, error) {

	tmpFile, err := os.CreateTemp("", "turnix-package-*.zip")
	if err != nil {
		return 0, err
	}
	defer os.Remove(tmpFile.Name())

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return 0, err
	}

	file := tmpFile.Name()

	return e.InstallPackageByFile(file)

}

func (e *Engine) InstallPackageByFile(file string) (int64, error) {
	packageId, err := e.db.InstallPackage(file)
	if err != nil {
		return 0, err
	}

	pkg, err := e.db.GetPackage(packageId)
	if err != nil {
		return 0, err
	}

	spaceId, err := e.db.AddSpace(&models.Space{
		PackageID:     packageId,
		Name:          pkg.Name,
		Info:          pkg.Info,
		NamespaceKey:  pkg.Slug,
		OwnsNamespace: true,
		Stype:         pkg.Type,
		OwnerID:       pkg.InstalledBy,
		ExtraMeta:     pkg.Info,
		IsInitilized:  true,
		IsPublic:      true,
	})

	if err != nil {
		return 0, err
	}

	e.LoadRoutingIndex()

	pp.Println("@spaceId", spaceId)

	return packageId, nil
}
