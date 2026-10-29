package web

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
)

const indexHTML = `<!doctype html><html lang="th"><head><title>มหานที</title></head><body><div id="root"></div><script type="module" src="/assets/index-abc123.js"></script></body></html>`

func builtFS() fstest.MapFS {
	return fstest.MapFS{
		".gitkeep":                 {Data: nil},
		"._index.html":             {Data: []byte("AppleDouble junk")},
		"index.html":               {Data: []byte(indexHTML)},
		"assets/index-abc123.js":   {Data: []byte(strings.Repeat("console.log('มหานที');\n", 200))},
		"assets/index-abc123.css":  {Data: []byte("body{margin:0}")},
		"assets/._index-abc123.js": {Data: []byte("junk")},
		"favicon.svg":              {Data: []byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`)},
	}
}

func get(t *testing.T, h http.Handler, method, target string, hdr map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func mustHandler(t *testing.T, fsys fstest.MapFS) *Handler {
	t.Helper()
	h, err := NewHandler(fsys)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	return h
}

func TestFreshCloneServesNotice(t *testing.T) {
	h := mustHandler(t, fstest.MapFS{".gitkeep": {Data: nil}})
	if h.Built() {
		t.Fatal("Built() = true for a tree with only .gitkeep")
	}
	rec := get(t, h, http.MethodGet, "/", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(rec.Body.String(), "make build") {
		t.Errorf("notice does not say how to build the front end: %q", rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("notice Cache-Control = %q, want no-store so a later build is not hidden", cc)
	}
	if rec := get(t, h, http.MethodGet, "/assets/app.js", nil); rec.Code != http.StatusNotFound {
		t.Errorf("GET /assets/app.js status = %d, want 404", rec.Code)
	}
}

func TestTheRealEmbedLoads(t *testing.T) {
	// Whatever is embedded (only .gitkeep on a fresh clone, a full build
	// after `make build`), the handler must construct without error.
	if _, err := NewHandler(Dist()); err != nil {
		t.Fatalf("NewHandler(Dist()): %v", err)
	}
}

func TestServesIndexWithRevalidation(t *testing.T) {
	h := mustHandler(t, builtFS())
	if !h.Built() {
		t.Fatal("Built() = false with index.html present")
	}
	rec := get(t, h, http.MethodGet, "/", nil)
	if rec.Code != http.StatusOK || rec.Body.String() != indexHTML {
		t.Fatalf("GET / = %d %q", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
		t.Errorf("index Cache-Control = %q, want no-cache", got)
	}
	etag := rec.Header().Get("ETag")
	if !strings.HasPrefix(etag, `"`) || len(etag) < 10 {
		t.Fatalf("ETag = %q, want a quoted strong validator", etag)
	}

	rec = get(t, h, http.MethodGet, "/", map[string]string{"If-None-Match": etag})
	if rec.Code != http.StatusNotModified || rec.Body.Len() != 0 {
		t.Errorf("conditional GET = %d with %d body bytes, want 304 and none", rec.Code, rec.Body.Len())
	}
	rec = get(t, h, http.MethodGet, "/index.html", nil)
	if rec.Code != http.StatusOK || rec.Body.String() != indexHTML {
		t.Errorf("GET /index.html = %d", rec.Code)
	}
}

func TestHashedAssetsAreImmutable(t *testing.T) {
	h := mustHandler(t, builtFS())
	rec := get(t, h, http.MethodGet, "/assets/index-abc123.js", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Errorf("Cache-Control = %q", got)
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/javascript") {
		t.Errorf("Content-Type = %q, want text/javascript", got)
	}
	// Files outside assets/ are not content-hashed, so they revalidate.
	rec = get(t, h, http.MethodGet, "/favicon.svg", nil)
	if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
		t.Errorf("favicon Cache-Control = %q, want no-cache", got)
	}
}

func TestGzipWhenAccepted(t *testing.T) {
	h := mustHandler(t, builtFS())
	plain := get(t, h, http.MethodGet, "/assets/index-abc123.js", nil)
	if plain.Header().Get("Content-Encoding") != "" {
		t.Fatal("gzip sent without Accept-Encoding")
	}
	if got := plain.Header().Get("Vary"); got != "Accept-Encoding" {
		t.Errorf("Vary = %q on the identity response", got)
	}

	rec := get(t, h, http.MethodGet, "/assets/index-abc123.js", map[string]string{"Accept-Encoding": "br, gzip;q=0.8"})
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", rec.Header().Get("Content-Encoding"))
	}
	if got := rec.Header().Get("Vary"); got != "Accept-Encoding" {
		t.Errorf("Vary = %q", got)
	}
	zr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(body, plain.Body.Bytes()) {
		t.Error("gzip body does not inflate to the identity body")
	}
	if rec.Header().Get("ETag") == plain.Header().Get("ETag") {
		t.Error("gzip and identity responses share an ETag; they are different bytes")
	}

	// gzip;q=0 is a refusal.
	rec = get(t, h, http.MethodGet, "/assets/index-abc123.js", map[string]string{"Accept-Encoding": "gzip;q=0"})
	if rec.Header().Get("Content-Encoding") != "" {
		t.Error("gzip sent although the client refused it with q=0")
	}

	// A tiny file is not worth compressing.
	rec = get(t, h, http.MethodGet, "/assets/index-abc123.css", map[string]string{"Accept-Encoding": "gzip"})
	if rec.Header().Get("Content-Encoding") != "" {
		t.Error("gzip used for a 14-byte file")
	}
}

func TestClientRoutesFallBackToIndex(t *testing.T) {
	h := mustHandler(t, builtFS())
	for _, p := range []string{"/station/737", "/about"} {
		rec := get(t, h, http.MethodGet, p, nil)
		if rec.Code != http.StatusOK || rec.Body.String() != indexHTML {
			t.Errorf("GET %s = %d, want index.html", p, rec.Code)
		}
	}
	for _, p := range []string{"/missing.png", "/assets/gone-123.js"} {
		if rec := get(t, h, http.MethodGet, p, nil); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404 (a file path is never the app)", p, rec.Code)
		}
	}
}

func TestHiddenAndSideFilesAreNeverServed(t *testing.T) {
	h := mustHandler(t, builtFS())
	for _, p := range []string{"/.gitkeep", "/._index.html", "/assets/._index-abc123.js", "/assets/../.gitkeep"} {
		if rec := get(t, h, http.MethodGet, p, nil); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", p, rec.Code)
		}
	}
}

func TestMethods(t *testing.T) {
	h := mustHandler(t, builtFS())
	rec := get(t, h, http.MethodHead, "/", nil)
	if rec.Code != http.StatusOK || rec.Body.Len() != 0 {
		t.Errorf("HEAD / = %d with %d body bytes", rec.Code, rec.Body.Len())
	}
	if got, want := rec.Header().Get("Content-Length"), strconv.Itoa(len(indexHTML)); got != want {
		t.Errorf("HEAD Content-Length = %q, want %q", got, want)
	}
	rec = get(t, h, http.MethodPost, "/", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST / = %d, want 405", rec.Code)
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Errorf("Allow = %q", got)
	}
}
