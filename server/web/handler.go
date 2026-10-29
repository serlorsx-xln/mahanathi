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

// placeholder is served at every app route when no front-end build is
// embedded, as on a fresh clone or the pre-Phase-3 deployment: it proves the
// pipeline is live by reading the national sentence and the five most severe
// stations from /api/v1/now. It is plain semantic HTML on purpose — Phase 3
// owns every styling decision — and it doubles as the developer notice.
const placeholder = `<!doctype html>
<html lang="th">
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>มหานที</title>
<header>
<h1>มหานที</h1>
<p>ระดับน้ำทั่วประเทศ ภาษาเดียวที่อ่านง่าย</p>
</header>
<main>
<p id="national">กำลังอ่านข้อมูลจากสถานีวัดระดับน้ำ…</p>
<ol id="severe"></ol>
</main>
<footer>
<p>ข้อมูลจากสถาบันสารสนเทศทรัพยากรน้ำ (HII) ผ่าน ThaiWater หน้านี้เป็นหน้าชั่วคราวก่อนเว็บเต็มรูปแบบ ตัวเซิร์ฟเวอร์นี้ยังไม่มี web build (run <code>make build</code>) API อยู่ที่ <code>/api/v1/</code></p>
</footer>
<script>
fetch('/api/v1/now').then(function (r) {
  if (!r.ok) { throw new Error(r.status); }
  return r.json();
}).then(function (d) {
  document.getElementById('national').textContent = d.summary.national;
  var list = document.getElementById('severe');
  d.stations
    .filter(function (s) { return !s.stale && s.suspect.length === 0 && s.level !== null; })
    .sort(function (a, b) { return b.level - a.level || (b.pct || 0) - (a.pct || 0); })
    .slice(0, 5)
    .forEach(function (s) {
      var li = document.createElement('li');
      var where = document.createElement('p');
      where.textContent = s.name + ' — ' + s.province;
      var sentence = document.createElement('p');
      sentence.textContent = s.sentence;
      li.appendChild(where);
      li.appendChild(sentence);
      list.appendChild(li);
    });
}).catch(function () {
  document.getElementById('national').textContent = 'ยังไม่มีข้อมูล ลองใหม่อีกครั้ง';
});
</script>
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
			_, _ = fmt.Fprint(w, placeholder)
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
