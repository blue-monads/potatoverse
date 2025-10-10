package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func hash(s string) uint32 {
	h := uint32(0)
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
	{"#6a11cb", "#2575fc"}, // Blue to purple
	{"#fdbb2d", "#22c1c3"}, // Yellow to cyan
	{"#ee9ca7", "#ffdde1"}, // Light pink to off-white
	{"#43e97b", "#38f9d7"}, // Mint to turquoise
	{"#ff9a9e", "#fad0c4"}, // Light pink to light red-orange
	{"#ffb347", "#ffcc33"}, // Orange to yellow
	{"#4568dc", "#b06ab3"}, // Blue to mauve
	{"#659999", "#f4791f"}, // Green-gray to orange
	{"#40e0d0", "#ff8c00"}, // Turquoise to dark orange
	{"#4ca1af", "#c4e0e5"}, // Blue to light blue
	{"#00b09b", "#96c93d"}, // Green to lime
	{"#f4791f", "#660000"}, // Orange to dark red
	{"#fc6767", "#ec008c"}, // Pink to dark pink
	{"#f7b733", "#fc4a1a"}, // Orange to red
	{"#e1eec3", "#f05053"}, // Pale green to red
	{"#a8c0ff", "#3f2b96"}, // Light blue to dark purple
	{"#e55d87", "#5fc3e4"}, // Pink to light blue
	{"#00d2ff", "#3a7bd5"}, // Sky blue to deep blue
	{"#a770ef", "#fdb99b"}, // Violet to peach
	{"#d53a9d", "#00d2ff"}, // Magenta to blue
	{"#334d50", "#cbcaa5"}, // Dark green to light yellow
	{"#74ebd5", "#acb6e5"}, // Aqua to light purple
	{"#f09819", "#edde5d"}, // Orange to light yellow
	{"#56ab2f", "#a8e063"}, // Forest green to light green
	{"#ffafbd", "#ffc3a0"}, // Light pink to peach
	{"#34e89e", "#0f3443"}, // Mint green to dark blue
	{"#76b852", "#8dc26f"}, // Olive green to bright green
	{"#00c6ff", "#0072ff"}, // Cyan to deep blue
	{"#4b6cb7", "#182848"}, // Royal blue to very dark blue
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
