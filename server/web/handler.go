package web

import (
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/serlorsx-xln/mahanathi/server/internal/httpx"
)

// Cache policies. Vite puts content-hashed files under assets/, so those can
// be cached forever; everything else must revalidate so a deploy shows up.
const (
	cacheImmutable  = "public, max-age=31536000, immutable"
	cacheRevalidate = "no-cache"
)

// notice is served at every app route when no front-end build is embedded,
// as on a fresh clone. It is plain on purpose: it is a developer message.
const notice = `<!doctype html>
<html lang="en">
<meta charset="utf-8">
<title>mahanathi: web build not embedded</title>
<h1>Web build not embedded</h1>
<p>This binary was built without the front end. Run <code>make build</code> from the repository root, which builds <code>web/</code> and copies it into <code>server/web/dist</code> before compiling. The API under <code>/api/v1/</code> works either way.</p>
`

// Handler serves the embedded front end from memory.
type Handler struct {
	files map[string]*httpx.Payload // keyed by slash path without a leading slash
	index *httpx.Payload            // nil when no build is embedded
}

// NewHandler reads every servable file in fsys into memory and prepares its
// validators and gzip variant. Dotfiles and macOS AppleDouble "._" files are
// skipped, so they can never be served.
func NewHandler(fsys fs.FS) (*Handler, error) {
	h := &Handler{files: map[string]*httpx.Payload{}}
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p != "." && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		body, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		cache := cacheRevalidate
		if strings.HasPrefix(p, "assets/") {
			cache = cacheImmutable
		}
		h.files[p] = httpx.NewPayload(body, contentType(p), cache)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("web: reading embedded files: %w", err)
	}
	h.index = h.files["index.html"]
	return h, nil
}

// Built reports whether a front-end build is embedded.
func (h *Handler) Built() bool { return h.index != nil }

// ServeHTTP serves a file by path. A path without a file extension is a
// client-side route and gets index.html; a missing file with an extension is
// a 404, so a broken asset link fails loudly instead of returning HTML.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if p, ok := h.files[name]; ok {
		p.ServeHTTP(w, r)
		return
	}
	if path.Ext(name) != "" {
		http.NotFound(w, r)
		return
	}
	if h.index == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		if r.Method != http.MethodHead {
			_, _ = fmt.Fprint(w, notice)
		}
		return
	}
	h.index.ServeHTTP(w, r)
}

// contentType picks a MIME type from the extension, with explicit answers for
// the types the build emits so the result does not depend on the host's
// mime.types file (distroless has none).
func contentType(p string) string {
	switch ext := path.Ext(p); ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".js", ".mjs":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".json", ".map":
		return "application/json"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".woff2":
		return "font/woff2"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".wasm":
		return "application/wasm"
	default:
		if t := mime.TypeByExtension(ext); t != "" {
			return t
		}
		return "application/octet-stream"
	}
}
