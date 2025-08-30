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
	{"#fc6767", "#ec008c"}, // Pink to red
	{"#00b09b", "#96c93d"}, // Green to lime
	{"#f7b733", "#fc4a1a"}, // Orange to red
	{"#fdbb2d", "#22c1c3"}, // Yellow to cyan
	{"#ee9ca7", "#ffdde1"}, // Light pink
	{"#43e97b", "#38f9d7"}, // Mint to turquoise
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
