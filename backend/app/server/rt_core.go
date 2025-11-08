package server

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
)

func (s *Server) getGlobalJS(ctx *gin.Context) {
	httpx.WriteFile("global.js", []byte(s.opt.GlobalJS), ctx)
}

func (s *Server) buildGlobalJS() error {
	finalJS := bytes.Buffer{}

	finalJS.WriteString(s.opt.GlobalJS)

	siteAttr := map[string]string{
		"site_name":  s.opt.SiteName,
		"site_hosts": strings.Join(s.opt.Hosts, ","),
	}

	siteAttrJSON, err := json.Marshal(siteAttr)
	if err != nil {
		return err
	}

	finalJS.WriteString("\n // siteAttr \n")
	finalJS.Write([]byte(`window.__potato_attrs__ = `))
	finalJS.Write(siteAttrJSON)
	finalJS.Write([]byte(`;`))

	s.opt.GlobalJS = finalJS.String()

	return nil

}
