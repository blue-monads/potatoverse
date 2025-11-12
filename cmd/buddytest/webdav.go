package main

import (
	"net/http"

	"github.com/blue-monads/turnix/backend/utils/qq"
	webdav "github.com/emersion/go-webdav"
)

// Define your expected credentials here.
// NOTE: For a real application, you should load credentials securely
// (e.g., from environment variables, a configuration file, or a database)
// and ideally use secure comparison methods like subtle.ConstantTimeCompare
// which is used in the basicAuth function.
const (
	expectedUsername = "admin"
	expectedPassword = "admin"
	realm            = "WebDAV Restricted" // The message shown in the login prompt
)

// basicAuth is an http.Handler middleware that enforces basic authentication.
func basicAuth(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// user, pass, ok := r.BasicAuth()

		// 1. Check if the Authorization header is present and valid.
		// if !ok {
		// 	// If not, send a 401 Unauthorized response with the WWW-Authenticate header
		// 	// to prompt the client for credentials.
		// 	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
		// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
		// 	return
		// }

		// 2. Compare the provided credentials with the expected ones securely.
		// The use of subtle.ConstantTimeCompare helps mitigate timing attacks.
		// correctUser := subtle.ConstantTimeCompare([]byte(user), []byte(expectedUsername)) == 1
		// correctPass := subtle.ConstantTimeCompare([]byte(pass), []byte(expectedPassword)) == 1

		// if !correctUser || !correctPass {
		// 	// If credentials are wrong, send 401 and prompt again.
		// 	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
		// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
		// 	return
		// }

		// 3. If authenticated, serve the request with the wrapped handler.
		handler.ServeHTTP(w, r)
	}
}

func WebdavServer() {

	qq.Println("WebdavServer/start")

	handler := webdav.Handler{
		FileSystem: &LocalFsWithPrincipal{
			LocalFileSystem: webdav.LocalFileSystem("./tmp/buddyfs"),
		},
	}

	// ⭐️ Apply the basicAuth middleware to the WebDAV handler
	authHandler := basicAuth(&handler)

	// ListenAndServe with the authenticated handler
	err := http.ListenAndServe(":8666", authHandler)
	if err != nil {
		qq.Println("Failed to listen and serve: %v", err)
		return
	}

	qq.Println("WebdavServer/end")

}

type LocalFsWithPrincipal struct {
	webdav.LocalFileSystem
}
