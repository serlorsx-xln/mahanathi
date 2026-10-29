// Package httpx holds the small HTTP helpers that the API and the static file
// server share: prepared payloads with a strong ETag and a precompressed gzip
// variant, and Accept-Encoding negotiation.
//
// Payloads are built once (when data changes, or at start-up for embedded
// files) and then served many times, so compression and hashing never run
// per request.
package httpx

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
)

// minGzipSize is the smallest body worth compressing. Below it, the gzip
// header and the extra round of negotiation cost more than they save.
const minGzipSize = 1024

// Payload is a response body prepared for repeated serving.
type Payload struct {
	Body         []byte
	Gzip         []byte // nil when compression would not help
	ETag         string // strong validator of Body
	GzipETag     string // strong validator of Gzip; differs because the bytes differ
	ContentType  string
	CacheControl string
}

// NewPayload hashes body and, when it is large enough to benefit, compresses
// it at the best gzip level.
func NewPayload(body []byte, contentType, cacheControl string) *Payload {
	p := &Payload{
		Body:         body,
		ContentType:  contentType,
		CacheControl: cacheControl,
	}
	sum := sha256.Sum256(body)
	tag := base64.RawURLEncoding.EncodeToString(sum[:12])
	p.ETag = `"` + tag + `"`

	if len(body) >= minGzipSize {
		var buf bytes.Buffer
		zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression) // the level is valid
		_, _ = zw.Write(body)                                    // writes to a bytes.Buffer cannot fail
		_ = zw.Close()
		// Keep the variant only if it saves at least a tenth.
		if buf.Len() < len(body)*9/10 {
			p.Gzip = buf.Bytes()
			p.GzipETag = `"` + tag + `-gz"`
		}
	}
	return p
}

// ServeHTTP writes the payload, choosing gzip when the client accepts it and
// answering 304 when the client's validator matches the chosen variant.
// HEAD gets the headers only.
func (p *Payload) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	body, etag := p.Body, p.ETag
	if p.Gzip != nil {
		h.Set("Vary", "Accept-Encoding")
		if AcceptsGzip(r.Header.Get("Accept-Encoding")) {
			body, etag = p.Gzip, p.GzipETag
			h.Set("Content-Encoding", "gzip")
		}
	}
	h.Set("ETag", etag)
	if p.CacheControl != "" {
		h.Set("Cache-Control", p.CacheControl)
	}

	if ETagMatches(r.Header.Get("If-None-Match"), etag) {
		// A 304 carries the validators but no body or Content-Type.
		h.Del("Content-Encoding")
		w.WriteHeader(http.StatusNotModified)
		return
	}

	h.Set("Content-Type", p.ContentType)
	h.Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}
}

// AcceptsGzip reports whether an Accept-Encoding header value allows gzip.
// It honours q=0 as a refusal, and accepts "*" unless gzip itself is refused.
func AcceptsGzip(header string) bool {
	star := false
	for _, part := range strings.Split(header, ",") {
		coding, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		coding = strings.ToLower(strings.TrimSpace(coding))
		if coding != "gzip" && coding != "x-gzip" && coding != "*" {
			continue
		}
		q := 1.0
		for _, param := range strings.Split(params, ";") {
			k, v, ok := strings.Cut(strings.TrimSpace(param), "=")
			if ok && strings.EqualFold(strings.TrimSpace(k), "q") {
				if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
					q = f
				}
			}
		}
		if coding == "*" {
			star = q > 0
			continue
		}
		return q > 0
	}
	return star
}

// ETagMatches implements the weak comparison that RFC 9110 specifies for
// If-None-Match: a "W/" prefix is ignored on either side, and "*" matches
// any current representation.
func ETagMatches(ifNoneMatch, etag string) bool {
	if ifNoneMatch == "" {
		return false
	}
	want := strings.TrimPrefix(etag, "W/")
	for _, candidate := range strings.Split(ifNoneMatch, ",") {
		c := strings.TrimSpace(candidate)
		if c == "*" || strings.TrimPrefix(c, "W/") == want {
			return true
		}
	}
	return false
}
