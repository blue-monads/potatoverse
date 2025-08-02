package blua

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/upper/db/v4"
	lua "github.com/yuin/gopher-lua"
)

type ExecHandle struct {
	baseFolder    *os.Root
	db            db.Session
	pid           int64 // project id
	activeRequest *gin.Context
}

type Blua struct {
	handle *ExecHandle
	L      *lua.LState
}

type Options struct {
	ProjectID   int64
	DB          db.Session
	BaseFolder  string
	SocurceCode string
}

func NewBlua(opts Options) (*Blua, error) {
	L := lua.NewState()
	err := L.DoString(opts.SocurceCode)
	if err != nil {
		return nil, err
	}

	rt, err := os.OpenRoot(opts.BaseFolder)
	if err != nil {
		return nil, fmt.Errorf("could not open root folder %s", opts.BaseFolder)
	}

	b := &Blua{
		handle: &ExecHandle{
			baseFolder: rt,
			db:         opts.DB,
			pid:        opts.ProjectID,
		},
		L: L,
	}

	return b, nil
}

func (b *Blua) Exec(ctx *gin.Context) error {
	b.handle.activeRequest = ctx

	err := b.L.CallByParam(lua.P{
		Fn:      b.L.GetGlobal("exec"),
		NRet:    0,
		Protect: true,
	})

	return err
}

func (b *Blua) Close() {
	b.handle = nil
	b.L.Close()
}

func (b *Blua) bind() error {
	tb := b.L.NewTable()

	// Set functions in the table using the scoped baseFolder
	b.L.SetFuncs(tb, map[string]lua.LGFunction{
		"exists":   b.exists,
		"read":     b.read,
		"write":    b.write,
		"mkdir":    b.mkdir,
		"remove":   b.remove,
		"dirname":  b.dirname,
		"basename": b.basename,
		"realpath": b.fnRealpath,
		"glob":     b.glob,
	})

	// Register the table as a global in the Lua state

	reqTb := b.L.NewTable()
	b.L.SetFuncs(reqTb, map[string]lua.LGFunction{
		"abort":               b.reqAbort,
		"abortWithStatus":     b.reqAbortWithStatus,
		"abortWithStatusJSON": b.reqAbortWithStatusJSON,
		"addParam":            b.reqAddParam,
		"clientIP":            b.reqClientIP,
		"contentType":         b.reqContentType,
		"cookie":              b.reqCookie,
		"data":                b.reqData,
		"defaultQuery":        b.reqDefaultQuery,
		"defaultPostForm":     b.reqDefaultPostForm,
		"fullPath":            b.reqFullPath,
		"getHeader":           b.reqGetHeader,
		"getQuery":            b.reqGetQuery,
		"getPostForm":         b.reqGetPostForm,
		"param":               b.reqParam,
		"redirect":            b.reqRedirect,
		"remoteIP":            b.reqRemoteIP,
		"json":                b.reqJSON,
		"html":                b.reqHTML,
		"string":              b.reqString,
		"setCookie":           b.reqSetCookie,
		"status":              b.reqStatus,
		"header":              b.reqHeader,
		"bindJSON":            b.reqBindJSON,
		"bindHeader":          b.reqBindHeader,
		"bindQuery":           b.reqBindQuery,
		"getRawData":          b.reqGetRawData,
		"formFile":            b.reqFormFile,
		"getQueryMap":         b.reqGetQueryMap,
		"getQueryArray":       b.reqGetQueryArray,
		"getPostFormMap":      b.reqGetPostFormMap,
		"getPostFormArray":    b.reqGetPostFormArray,
		"ssEvent":             b.reqSSEvent,
	})

	b.L.SetGlobal("ctx", reqTb)

	// users, selfspace

	return nil
}
