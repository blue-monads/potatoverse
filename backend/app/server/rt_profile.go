package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var GradientSeed = 71

func hash(s string) uint32 {
	h := uint32(GradientSeed)
	for i := 0; i < len(s); i++ {
		h = h*31 + uint32(s[i])
	}
	return h
}

type Gradient struct {
	Start string
	End   string
}

var gradients = []Gradient{
	{"#d53a9d", "#cbcaa5"},
	{"#74ebd5", "#edde5d"},
	{"#6a11cb", "#2575fc"},
	{"#fdbb2d", "#22c1c3"},
	{"#ee9ca7", "#ffdde1"},
	{"#43e97b", "#38f9d7"},
	{"#ff9a9e", "#fad0c4"},
	{"#ffb347", "#ffcc33"},
	{"#4568dc", "#b06ab3"},
	{"#659999", "#f4791f"},
	{"#40e0d0", "#ff8c00"},
	{"#4ca1af", "#c4e0e5"},
	{"#00b09b", "#96c93d"},
	{"#f4791f", "#660000"},
	{"#fc6767", "#ec008c"},
	{"#f7b733", "#fc4a1a"},
	{"#e1eec3", "#f05053"},
	{"#a8c0ff", "#3f2b96"},
	{"#e55d87", "#5fc3e4"},
	{"#00d2ff", "#3a7bd5"},
	{"#a770ef", "#fdb99b"},
	{"#d53a9d", "#00d2ff"},
	{"#334d50", "#cbcaa5"},
	{"#74ebd5", "#acb6e5"},
	{"#f09819", "#edde5d"},
	{"#56ab2f", "#a8e063"},
	{"#ffafbd", "#ffc3a0"},
	{"#34e89e", "#0f3443"},
	{"#76b852", "#edc26f"},
	{"#00c6ff", "#0072ff"},
	{"#4b6cb7", "#182848"},
	{"#f4791f", "#fdb99b"},
	{"#f7b733", "#ec008c"},
	{"#f7b733", "#fc4a1a"},
	{"#a770ef", "#660000"},
}

func (a *Server) ListGradients(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gradients)
}

func (a *Server) userSvgProfileIconById(ctx *gin.Context) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	// fixme => reterive only name not all user fields
	user, err := a.ctrl.GetUser(int64(idInt))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	a.userSvgProfileIconWithName(ctx, user.Name)
}

func (a *Server) userSvgProfileIcon(ctx *gin.Context) {
	name := ctx.Param("name")
	a.userSvgProfileIconWithName(ctx, name)
}

func (a *Server) userSvgProfileIconWithName(ctx *gin.Context, name string) {

	ctx.Header("Cache-Control", "public, max-age=86400")

	splits := strings.Fields(name)
	initials := ""
	if len(splits) > 0 {
		initials += strings.ToUpper(splits[0][0:1])
		if len(splits) > 1 {
			initials += strings.ToUpper(splits[1][0:1])
		}
	} else if len(name) > 1 {
		initials += strings.ToUpper(name[0:1])
		initials += strings.ToUpper(name[1:2])
	} else {
		initials += "NA"
	}

	nameHash := hash(name)
	gradientIndex := int(nameHash) % len(gradients)
	selectedGradient := gradients[gradientIndex]

	svgIcon := fmt.Sprintf(`
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
        <defs>
            <linearGradient id="gradient" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
                <stop offset="0%%" style="stop-color:%s;stop-opacity:1" />
                <stop offset="100%%" style="stop-color:%s;stop-opacity:1" />
            </linearGradient>
        </defs>
        <circle cx="50" cy="50" r="50" fill="url(#gradient)" />
        <text x="50%%" y="50%%" text-anchor="middle" dominant-baseline="middle" font-family="Arial, sans-serif" font-size="40" font-weight="bold" fill="white">%s</text>
    </svg>      
    `, selectedGradient.Start, selectedGradient.End, initials)

	ctx.Data(http.StatusOK, "image/svg+xml", []byte(svgIcon))
}
