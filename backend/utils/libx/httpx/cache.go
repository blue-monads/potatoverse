package httpx

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
)

// https://github.com/benbjohnson/hashfs/blob/main/hashfs.go
func WriteWithCache(filename string, w http.ResponseWriter, r *http.Request) {
	var file, err = os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	fi, err := file.Stat()
	if err != nil {
		panic(err)
	}

	if ext := path.Ext(filename); ext != "" {
		w.Header().Set("Content-Type", mime.TypeByExtension(ext))
	}

	w.Header().Set("Cache-Control", `public, max-age=31536000`)

	// Set content length.
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))

	// Flush header and write content.
	w.WriteHeader(http.StatusOK)
	if r.Method != "HEAD" {
		io.Copy(w, file)
	}
}
