// Package web will embed the built front end (web/dist, copied here at build
// time) so one binary serves both the app and the API. go:embed cannot reach
// outside the module directory, which is why the build copies the files in
// rather than embedding ../web/dist directly.
package web
