package server

import (
	"fmt"
	"io"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
)

type InstallPackageRequest struct {
	URL string `json:"url"`
}

func (a *Server) InstallPackage(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	var req InstallPackageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {

		return nil, err
	}

	packageId, err := a.ctrl.InstallPackageByUrl(claim.UserId, req.URL)
	if err != nil {
		return nil, err
	}

	return gin.H{"package_id": packageId}, nil

}

func (a *Server) InstallPackageZip(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {

	tempFile, err := os.CreateTemp("", "turnix-package-*.zip")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, ctx.Request.Body)
	if err != nil {
		return nil, err
	}
	packageId, err := a.ctrl.InstallPackageByFile(claim.UserId, tempFile.Name())
	if err != nil {
		return nil, err
	}

	return gin.H{"package_id": packageId}, nil
}

type InstallPackageEmbedRequest struct {
	Name string `json:"name"`
}

func (a *Server) InstallPackageEmbed(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	var req InstallPackageEmbedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	packageId, err := a.ctrl.InstallPackageEmbed(claim.UserId, req.Name)
	if err != nil {
		return nil, err
	}

	return gin.H{"package_id": packageId}, nil
}

func (a *Server) ListEPackages(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	epackages, err := a.ctrl.ListEPackages()
	if err != nil {
		return nil, err
	}

	return epackages, nil
}

func (a *Server) ListInstalledSpaces(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {

	spaces, err := a.ctrl.ListInstalledSpaces(claim.UserId)
	if err != nil {
		return nil, err
	}

	return spaces, nil
}

func (a *Server) AuthorizeSpace(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {

	data := &actions.SpaceAuth{}
	if err := ctx.ShouldBindJSON(data); err != nil {
		return nil, err
	}

	token, err := a.ctrl.AuthorizeSpace(claim.UserId, *data)
	if err != nil {
		return nil, err
	}

	return gin.H{"token": token}, nil
}

func (a *Server) DeletePackage(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	packageId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}
	err = a.ctrl.DeletePackage(claim.UserId, packageId)
	if err != nil {
		return nil, err
	}

	return nil, nil

}

func (a *Server) GeneratePackageDevToken(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	packageId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	token, err := a.ctrl.GeneratePackageDevToken(claim.UserId, packageId)
	if err != nil {
		return nil, err
	}

	return gin.H{"token": token}, nil
}

// engine core

func (a *Server) handleSpaceFile() func(ctx *gin.Context) {

	proxyAddrs := map[string]*httputil.ReverseProxy{}

	if DEV_MODE {
		devSpacesEnv := os.Getenv("TURNIX_DEV_SPACES")
		devSpaces := strings.Split(devSpacesEnv, ",")

		for _, pname := range devSpaces {
			nameParts := strings.Split(pname, ":")
			if len(nameParts) != 2 {
				continue
			}

			url, err := url.Parse(fmt.Sprint("http://localhost:", nameParts[1]))
			if err != nil {
				panic(err)
			}
			proxy := httputil.NewSingleHostReverseProxy(url)
			proxyAddrs[nameParts[0]] = proxy
		}
	}

	return func(ctx *gin.Context) {

		spaceKey := ctx.Param("space_key")
		proxy := proxyAddrs[spaceKey]
		if proxy != nil {
			proxy.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}

		a.engine.ServeSpaceFile(ctx)

	}
}

func (a *Server) handleEngineDebugData(ctx *gin.Context) {
	debugData := a.ctrl.GetEngineDebugData()
	httpx.WriteJSON(ctx, debugData, nil)
}

func (a *Server) handleSpaceApi(ctx *gin.Context) {
	a.engine.SpaceApi(ctx)
}

func (a *Server) handlePluginFile() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {}
}

func (a *Server) handlePluginApi(ctx *gin.Context) {}

func (a *Server) handleSpaceInfo(ctx *gin.Context) {
	spaceKey := ctx.Param("space_key")
	spaceInfo, err := a.engine.SpaceInfo(spaceKey)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
	httpx.WriteJSON(ctx, spaceInfo, nil)
}
