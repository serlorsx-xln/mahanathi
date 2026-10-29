// Package web embeds the built front end (web/dist, copied into
// server/web/dist by `make build` and by the Dockerfile) and serves it, so
// one binary carries both the app and the API.
//
// go:embed cannot reach outside the module directory, which is why the build
// copies the files in rather than embedding ../web/dist directly. A fresh
// clone has only dist/.gitkeep, so `go build` and `go test` still work
// without a front-end build; the handler then serves a short notice instead
// of the app.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embedded embed.FS

// Dist returns the embedded build, rooted at dist/.
func Dist() fs.FS {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		// fs.Sub fails only for an invalid path, and "dist" is valid.
		panic(err)
	}
	return sub
}
