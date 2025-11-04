package engine

import (
	"fmt"
	"strings"

	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes/models"
	"github.com/gin-gonic/gin"
)

/*
EXAMPLE
{
                "router": "dynamic",
                "serve_folder": "public",
                "template_folder": "templates",
                "routes": [
                    {
                        "path": "/",
                        "method": "GET",
                        "type": "static",
                        "file": "index.html"
                    },
                    {
						// in this case :id is a path parameter has no effect, it still serves the same file
                        "path": "/product/:id/edit",
                        "method": "GET",
                        "type": "static",
                        "file": "edit.html"
                    },
                    {
                        "path": "/categories/:id/page.html",
                        "method": "GET",
                        "type": "template",
                        "handler": "get_category_page",
                        "file": "category.template.html"
                    },
                    {
                        "path": "/categories/create",
                        "method": "POST",
                        "type": "api",
                        "handler": "create_category"
                    }
                ]
            }

*/

func (e *Engine) serveDynamicRoute(ctx *gin.Context, indexItem *SpaceRouteIndexItem) {

	// Get the request path and method
	requestPath := ctx.Param("subpath")
	requestMethod := ctx.Request.Method

	qq.Println("@requestPath", requestPath)
	qq.Println("@requestMethod", requestMethod)

	// Find matching route
	matchedRoute, pathParams := e.findMatchingRoute(indexItem, requestPath, requestMethod)

	if matchedRoute == nil {
		httpx.WriteErrString(ctx, "route not found")
		return
	}

	qq.Println("@matchedRoute", matchedRoute)
	if matchedRoute.Type != "static" {
		qq.Println("@pathParams", pathParams)
	}

	// Handle different route types
	switch matchedRoute.Type {
	case "static":
		e.handleStaticRoute(ctx, indexItem, matchedRoute)
	case "template":
		e.handleTemplateRoute(ctx, indexItem, matchedRoute, pathParams)
	case "api":
		e.handleApiRoute(ctx, indexItem, matchedRoute, pathParams)
	default:
		httpx.WriteErrString(ctx, fmt.Sprintf("unsupported route type: %s", matchedRoute.Type))
	}
}

// handleStaticRoute serves static files based on the route configuration
func (e *Engine) handleStaticRoute(ctx *gin.Context, indexItem *SpaceRouteIndexItem, routeMatch *models.PotatoRoute) {
	// Use the file specified in the route, or fall back to the request path
	filePath := routeMatch.File
	if filePath == "" {
		filePath = ctx.Request.URL.Path
	}

	// Build the package file path
	name, path := buildPackageFilePath(filePath, &indexItem.routeOption)

	qq.Println("@static name", name)
	qq.Println("@static path", path)

	pFileOps := e.db.GetPackageFileOps()
	err := pFileOps.StreamFileToHTTP(indexItem.packageVersionId, path, name, ctx.Writer)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
}

func (e *Engine) handleTemplateRoute(ctx *gin.Context, indexItem *SpaceRouteIndexItem, routeMatch *models.PotatoRoute, pathParams map[string]string) {
	if routeMatch.Handler == "" {
		httpx.WriteErrString(ctx, "template handler not specified")
		return
	}
	spaceKey := ctx.Param("space_key")
	for key, value := range pathParams {
		ctx.Set(key, value)
	}

	err := e.runtime.ExecuteHttp(ExecuteOptions{
		PackageName:      spaceKey,
		PackageVersionId: indexItem.packageVersionId,
		InstalledId:      indexItem.installedId,
		SpaceId:          indexItem.spaceId,
		HandlerName:      routeMatch.Handler,
		HttpContext:      ctx,
		Params:           pathParams,
	})

	tmpl, ok := indexItem.compiledTemplates[routeMatch.File]
	if !ok {
		httpx.WriteErrString(ctx, "template not found")
		return
	}

	qq.Println("@template ctx.Keys", pathParams)

	err = tmpl.Execute(ctx.Writer, ctx.Keys)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
}

// handleApiRoute handles API endpoints
func (e *Engine) handleApiRoute(ctx *gin.Context, indexItem *SpaceRouteIndexItem, routeMatch *models.PotatoRoute, pathParams map[string]string) {
	if routeMatch.Handler == "" {
		httpx.WriteErrString(ctx, "API handler not specified")
		return
	}

	// Add path parameters to the context for the handler
	for key, value := range pathParams {
		ctx.Set(key, value)
	}

	// Execute the handler
	spaceKey := ctx.Param("space_key")
	err := e.runtime.ExecuteHttp(ExecuteOptions{
		PackageName:      spaceKey,
		PackageVersionId: indexItem.packageVersionId,
		InstalledId:      indexItem.installedId,
		SpaceId:          indexItem.spaceId,
		HandlerName:      routeMatch.Handler,
		HttpContext:      ctx,
		Params:           pathParams,
	})

	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
}

func (e *Engine) findMatchingRoute(indexItem *SpaceRouteIndexItem, requestPath, requestMethod string) (*models.PotatoRoute, map[string]string) {
	for _, route := range indexItem.routeOption.Routes {
		// Check if method matches
		if route.Method != requestMethod {
			continue
		}

		// Check if path matches and extract parameters
		pathParams, matches := matchPath(route.Path, requestPath)
		if matches {
			return &route, pathParams
		}
	}
	return nil, nil
}

func matchPath(routePattern, requestPath string) (map[string]string, bool) {
	patternSegments := strings.Split(strings.Trim(routePattern, "/"), "/")
	pathSegments := strings.Split(strings.Trim(requestPath, "/"), "/")

	qq.Println("@matchPath/1", patternSegments, pathSegments)

	if len(patternSegments) != len(pathSegments) {
		qq.Println("@matchPath/2", len(patternSegments), len(pathSegments))
		return nil, false
	}

	params := make(map[string]string)

	for i, patternSegment := range patternSegments {
		pathSegment := pathSegments[i]

		// Check if this is a parameter segment (starts with :)
		if strings.HasPrefix(patternSegment, ":") {
			// Extract parameter name (remove the :)
			paramName := patternSegment[1:]
			// Validate parameter name
			if !isValidParamName(paramName) {
				return nil, false
			}
			// Store the parameter value
			params[paramName] = pathSegment
		} else {
			// Static segment - must match exactly
			if patternSegment != pathSegment {
				return nil, false
			}
		}
	}

	return params, true
}

func isValidParamName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// First character must be letter or underscore
	first := name[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Rest of characters must be letter, digit, or underscore
	for i := 1; i < len(name); i++ {
		char := name[i]
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}
