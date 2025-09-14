package engine

import (
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/models"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

type indexItem struct {
	packageId   int64
	spaceId     int64
	serveFolder string
}

type Engine struct {
	db           datahub.Database
	RoutingIndex map[string]indexItem
	riLock       sync.RWMutex

	app xtypes.App
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
				packageId:   space.PackageID,
				spaceId:     space.ID,
				serveFolder: "public",
			}
		}
	}

	e.riLock.Lock()
	e.RoutingIndex = nextRoutingIndex
	e.riLock.Unlock()

	return nil
}

func (e *Engine) Start(app xtypes.App) error {
	e.app = app

	return e.LoadRoutingIndex()
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
	pathParts := nameParts[:len(nameParts)-1]
	pathParts = append(pathParts, ri.serveFolder)

	path := strings.Join(pathParts, "/")
	path = strings.TrimLeft(path, "/")

	pp.Println("@name", name)
	pp.Println("@path", path)

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

func (e *Engine) InstallPackageByUrl(userId int64, url string) (int64, error) {

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

	return e.InstallPackageByFile(userId, file)

}

func (e *Engine) InstallPackageEmbed(userId int64, name string) (int64, error) {
	file, err := ZipEPackage(name)
	if err != nil {
		return 0, err
	}

	defer os.Remove(file)

	return e.InstallPackageByFile(userId, file)
}

func (e *Engine) InstallPackageByFile(userId int64, file string) (int64, error) {
	packageId, err := e.db.InstallPackage(userId, file)
	if err != nil {
		return 0, err
	}

	pkg, err := e.db.GetPackage(packageId)
	if err != nil {
		return 0, err
	}

	spaceId, err := e.db.AddSpace(&models.Space{
		PackageID:     packageId,
		NamespaceKey:  pkg.Slug,
		OwnsNamespace: true,
		ExecutorType:  "luaz",
		SubType:       "space",
		OwnerID:       pkg.InstalledBy,
		IsInitilized:  false,
		IsPublic:      true,
	})

	if err != nil {
		return 0, err
	}

	e.LoadRoutingIndex()

	pp.Println("@spaceId", spaceId)

	return packageId, nil
}
