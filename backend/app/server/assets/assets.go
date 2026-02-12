package assets

import (
	"crypto/md5"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	output "github.com/blue-monads/potatoverse/frontend"
	"github.com/gin-gonic/gin"
)

const NoPreBuildFiles = false

var (
	etagCache     sync.Map
	etagCacheOnce sync.Once
)

// Pre-compute ETags for all embedded files at startup
func initETagCache() {
	etagCacheOnce.Do(func() {
		qq.Println("@computing_etags_for_embedded_files")

		fs.WalkDir(output.BuildProd, "output/build", func(filePath string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			// Read file content
			content, err := output.BuildProd.ReadFile(filePath)
			if err != nil {
				return nil
			}

			// Compute MD5 hash
			hash := md5.Sum(content)
			etag := fmt.Sprintf(`"%x"`, hash)

			// Store with path relative to "output/build"
			relativePath := strings.TrimPrefix(filePath, "output/build/")
			etagCache.Store(relativePath, etag)

			return nil
		})

		qq.Println("@etag_cache_initialized")
	})
}

func PagesRoutesServer() gin.HandlerFunc {
	var proxy *httputil.ReverseProxy
	pserver := os.Getenv("FRONTEND_DEV_SERVER")

	if pserver != "" && !NoPreBuildFiles {
		url, err := url.Parse(pserver)
		if err != nil {
			panic(err)
		}
		qq.Println("@using_dev_proxy", pserver)
		proxy = httputil.NewSingleHostReverseProxy(url)

		return func(ctx *gin.Context) {
			qq.Println("[PROXY]", ctx.Request.URL.String())
			proxy.ServeHTTP(ctx.Writer, ctx.Request)
		}
	}

	qq.Println("@not_using_dev_proxy")

	// Initialize ETag cache for production builds
	initETagCache()

	return func(ctx *gin.Context) {
		ppath := strings.TrimSuffix(strings.TrimPrefix(ctx.Request.URL.Path, "/zz/pages"), "/")
		if ppath == "" {
			ppath = "index.html"
		}

		pitems := strings.Split(ppath, "/")
		lastpath := pitems[len(pitems)-1]
		if !strings.Contains(lastpath, ".") {
			ppath = ppath + ".html"
		}

		qq.Println("@FILE ==>", ppath)

		// Check if we have an ETag for this file
		if etag, ok := etagCache.Load(ppath); ok {
			etagStr := etag.(string)

			// Set ETag header
			ctx.Header("ETag", etagStr)

			if strings.HasSuffix(etagStr, ".html") {
				ctx.Header("Cache-Control", "public, max-age=60, must-revalidate")
			} else {
				ctx.Header("Cache-Control", "public, max-age=86400, must-revalidate")
			}

			ctx.Header("Vary", "Accept-Encoding")

			// Check If-None-Match header for conditional requests
			if match := ctx.GetHeader("If-None-Match"); match == etagStr {
				qq.Println("@etag_match_304", ppath)
				ctx.Status(http.StatusNotModified)
				return
			}
		} else {

			ctx.Header("Cache-Control", "public, max-age=60")
			qq.Println("@no_etag_fallback_cache", ppath)
		}

		// Read and serve the file
		out, err := output.BuildProd.ReadFile(path.Join("output/build", ppath))
		if err != nil {
			qq.Println("@open_err", err.Error())
			ctx.Status(http.StatusNotFound)
			return
		}

		ext := filepath.Ext(ppath)
		mimeType := mime.TypeByExtension(ext)
		ctx.Data(200, mimeType, out)

	}
}
