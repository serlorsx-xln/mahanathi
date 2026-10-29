package httpx

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAcceptsGzip(t *testing.T) {
	cases := []struct {
		header string
		want   bool
	}{
		{"", false},
		{"gzip", true},
		{"GZIP", true},
		{"x-gzip", true},
		{"br", false},
		{"br, gzip", true},
		{"gzip;q=0", false},
		{"gzip; q=0.0", false},
		{"gzip;q=0.001", true},
		{"*", true},
		{"*;q=0", false},
		{"br, *", true},
		{"gzip;q=0, *", false},
		{"identity", false},
		{"deflate, gzip;q=0.5", true},
	}
	for _, c := range cases {
		if got := AcceptsGzip(c.header); got != c.want {
			t.Errorf("AcceptsGzip(%q) = %v, want %v", c.header, got, c.want)
		}
	}
}

func TestETagMatches(t *testing.T) {
	cases := []struct {
		inm, etag string
		want      bool
	}{
		{"", `"a"`, false},
		{`"a"`, `"a"`, true},
		{`W/"a"`, `"a"`, true},
		{`"b", "a"`, `"a"`, true},
		{`"b"`, `"a"`, false},
		{"*", `"a"`, true},
		{`"a-gz"`, `"a"`, false},
	}
	for _, c := range cases {
		if got := ETagMatches(c.inm, c.etag); got != c.want {
			t.Errorf("ETagMatches(%q, %q) = %v, want %v", c.inm, c.etag, got, c.want)
		}
	}
}

func serve(p *Payload, method string, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/x", nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)
	return rec
}

func TestPayloadVariantsAndValidators(t *testing.T) {
	body := []byte(strings.Repeat(`{"station":"วัดท่าเจดีย์"},`, 100))
	p := NewPayload(body, "application/json", "public, max-age=300")
	if p.Gzip == nil {
		t.Fatal("a repetitive 3 KB body was not compressed")
	}
	if p.ETag == p.GzipETag {
		t.Fatal("identity and gzip variants share an ETag")
	}

	plain := serve(p, http.MethodGet, nil)
	if plain.Code != 200 || !bytes.Equal(plain.Body.Bytes(), body) {
		t.Fatalf("identity GET = %d", plain.Code)
	}
	for k, want := range map[string]string{
		"Content-Type":  "application/json",
		"Cache-Control": "public, max-age=300",
		"Vary":          "Accept-Encoding",
		"ETag":          p.ETag,
	} {
		if got := plain.Header().Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}

	gz := serve(p, http.MethodGet, map[string]string{"Accept-Encoding": "gzip"})
	if gz.Header().Get("Content-Encoding") != "gzip" || gz.Header().Get("ETag") != p.GzipETag {
		t.Fatalf("gzip GET headers = %v", gz.Header())
	}
	zr, err := gzip.NewReader(gz.Body)
	if err != nil {
		t.Fatal(err)
	}
	inflated, _ := io.ReadAll(zr)
	if !bytes.Equal(inflated, body) {
		t.Fatal("gzip variant does not inflate to the body")
	}

	// The validator must match the variant the client would get.
	notMod := serve(p, http.MethodGet, map[string]string{"Accept-Encoding": "gzip", "If-None-Match": p.GzipETag})
	if notMod.Code != http.StatusNotModified || notMod.Body.Len() != 0 {
		t.Errorf("matching If-None-Match = %d", notMod.Code)
	}
	if notMod.Header().Get("Content-Encoding") != "" {
		t.Error("304 carries Content-Encoding")
	}
	cross := serve(p, http.MethodGet, map[string]string{"If-None-Match": p.GzipETag})
	if cross.Code != http.StatusOK {
		t.Errorf("gzip ETag revalidated an identity request: %d", cross.Code)
	}

	head := serve(p, http.MethodHead, map[string]string{"Accept-Encoding": "gzip"})
	if head.Body.Len() != 0 || head.Header().Get("Content-Length") == "" {
		t.Errorf("HEAD body %d bytes, Content-Length %q", head.Body.Len(), head.Header().Get("Content-Length"))
	}
}

func TestSmallOrIncompressibleBodiesStayIdentity(t *testing.T) {
	small := NewPayload([]byte(`{"ok":true}`), "application/json", "")
	if small.Gzip != nil {
		t.Error("an 11-byte body was compressed")
	}
	rec := serve(small, http.MethodGet, map[string]string{"Accept-Encoding": "gzip"})
	if rec.Header().Get("Content-Encoding") != "" || rec.Header().Get("Vary") != "" {
		t.Errorf("identity-only payload sent %v", rec.Header())
	}
	if rec.Header().Get("Cache-Control") != "" {
		t.Error("empty CacheControl still set the header")
	}

	// Pseudo-random bytes do not compress by a tenth.
	noise := make([]byte, 4096)
	x := uint32(2463534242)
	for i := range noise {
		x ^= x << 13
		x ^= x >> 17
		x ^= x << 5
		noise[i] = byte(x)
	}
	if NewPayload(noise, "application/octet-stream", "").Gzip != nil {
		t.Error("kept a gzip variant that saves less than a tenth")
	}
}
