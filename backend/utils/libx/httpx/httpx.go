package httpx

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CtypeJSON = "application/json"
	CtypeJS   = "application/javascript"
	CtypeBin  = "application/octet-stream"
	CtypeCSS  = "text/css"
)

func NotFound(ctx *gin.Context) {
	ctx.Data(http.StatusNotFound, "text/html", []byte(`<h1>Not Found<h1>`))
}

func UnAuthorized(ctx *gin.Context) {
	ctx.Data(http.StatusUnauthorized, "text/html", []byte(`<h1>Unauthorized<h1>`))
}

func WriteFile(file string, data []byte, ctx *gin.Context) {

	ffiles := strings.Split(file, ".")

	ctype := ""
	switch ffiles[len(ffiles)-1] {
	case "js":
		ctype = CtypeJS
	case "css":
		ctype = CtypeCSS
	default:
		ctype = http.DetectContentType(data)
	}

	ctx.Data(http.StatusOK, ctype, data)
}
