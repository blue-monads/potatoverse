package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/blue-monads/turnix/backend/utils/qq"
	"golang.org/x/net/webdav"
)

type WebdavServer struct {
	Host  string
	Port  int
	FsDir string
}

func NewWebdavServer(host string, port int, fsDir string) *WebdavServer {
	return &WebdavServer{
		Host:  host,
		Port:  port,
		FsDir: fsDir,
	}
}

func (s *WebdavServer) Listen() {
	lock := webdav.NewMemLS()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		qq.Println("req/1")

		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			qq.Println("req/2")
			return
		}
		if username != "flydav" || password != "flydav12345" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			qq.Println("req/3")
			return
		}

		davHandler := &webdav.Handler{
			Prefix:     "/",
			FileSystem: buildDirName(s.FsDir, username),
			LockSystem: lock,
		}

		qq.Println("req/4")

		davHandler.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		qq.Println("Failed to listen and serve: %v", err)
		return
	}

	qq.Println("WebdavServer/end")

}

func buildDirName(fsDir, subFsDir string) webdav.Dir {
	qq.Println("buildDirName/1", fsDir, subFsDir)

	if subFsDir == "" {
		qq.Println("buildDirName/2")
		return webdav.Dir(fsDir)
	}

	qq.Println("buildDirName/3")

	dir := filepath.Join(fsDir, subFsDir)

	qq.Println("buildDirName/4")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		qq.Println("buildDirName/3")

		os.MkdirAll(dir, 0755)
		qq.Println("buildDirName/4")
	}

	qq.Println("buildDirName/5")

	return webdav.Dir(dir)
}
