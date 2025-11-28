package server

import (
	"errors"
	"fmt"
	"io"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/services/signer"
	xutils "github.com/blue-monads/turnix/backend/utils"
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

	ipackage, err := a.ctrl.InstallPackageByUrl(claim.UserId, req.URL)
	if err != nil {
		return nil, err
	}

	return ipackage, nil

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
	ipackage, err := a.ctrl.InstallPackageByFile(claim.UserId, tempFile.Name())
	if err != nil {
		return nil, err
	}

	return ipackage, nil
}

type InstallPackageEmbedRequest struct {
	Name     string `json:"name"`
	RepoSlug string `json:"repo_slug,omitempty"`
}

func (a *Server) InstallPackageEmbed(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	var req InstallPackageEmbedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	ipackage, err := a.ctrl.InstallPackageEmbed(claim.UserId, req.Name, req.RepoSlug)
	if err != nil {
		return nil, err
	}

	return ipackage, nil
}

func (a *Server) ListEPackages(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	repoSlug := ctx.Query("repo")
	epackages, err := a.ctrl.ListEPackages(repoSlug)
	if err != nil {
		return nil, err
	}

	return epackages, nil
}

func (a *Server) ListRepos(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	repos, err := a.ctrl.ListRepos()
	if err != nil {
		return nil, err
	}

	return repos, nil
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
		// TURNIX_DEV_SPACES="space1:8080,space2:8081"
		devSpacesEnv := os.Getenv("TURNIX_DEV_SPACES")
		devSpaces := strings.SplitSeq(devSpacesEnv, ",")

		for pname := range devSpaces {
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

func (a *Server) handleDeriveHost(ctx *gin.Context) {
	nskey := ctx.Param("nskey")
	hostName := ctx.Query("host_name")
	spaceIdStr := ctx.Query("space_id")
	spaceId, _ := strconv.ParseInt(spaceIdStr, 10, 64)

	if hostName == "" {
		hostName = ctx.Request.Host
	}

	if spaceId == 0 {
		spaceInfo, err := a.engine.SpaceInfo(nskey, hostName)
		if err != nil {
			httpx.WriteErr(ctx, err)
			return
		}
		spaceId = spaceInfo.ID
	}

	execHost := xutils.BuildExecHost(hostName, spaceId, a.opt.Hosts, a.opt.ServerKey)
	httpx.WriteJSON(ctx, gin.H{
		"host":     execHost,
		"space_id": spaceId,
	}, nil)

}

func (a *Server) handleSpaceInfo(ctx *gin.Context) {
	spaceKey := ctx.Param("space_key")
	hostName := ctx.Query("host_name")
	spaceInfo, err := a.engine.SpaceInfo(spaceKey, hostName)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
	httpx.WriteJSON(ctx, spaceInfo, nil)
}

func (a *Server) PushPackage(ctx *gin.Context) {
	// Get the dev token from Authorization header
	token := ctx.GetHeader("Authorization")
	if token == "" {
		httpx.WriteErr(ctx, errors.New("missing authorization token"))
		return
	}

	// Parse the package dev token
	claim, err := a.signer.ParsePackageDev(token)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	recreateArtifacts := ctx.Query("recreate_artifacts") == "true"

	// Create temp file for the uploaded zip
	tempFile, err := os.CreateTemp("", "turnix-package-push-*.zip")
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
	defer os.Remove(tempFile.Name())

	// Copy the uploaded file to temp file
	_, err = io.Copy(tempFile, ctx.Request.Body)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	// Close the temp file before passing to UpgradePackage
	tempFile.Close()

	// Upgrade the package
	packageId, err := a.ctrl.UpgradePackage(claim.UserId, tempFile.Name(), claim.InstallPackageId, recreateArtifacts)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteJSON(ctx, gin.H{
		"package_version_id": packageId,
		"message":            "package upgraded successfully",
	}, nil)
}

func (a *Server) handleCapabilities(ctx *gin.Context) {
	a.engine.ServeCapability(ctx)
}

func (a *Server) handleCapabilitiesRoot(ctx *gin.Context) {
	a.engine.ServeCapabilityRoot(ctx)
}
