package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/rakepro/rakepro-web/internal/config"
)

func TestStaticDirectoryPathsNotFound(t *testing.T) {
	s, err := New(config.Config{Addr: ":0"}, zerolog.Nop())
	if err != nil {
		t.Fatal(err)
	}
	h := s.routes()

	cases := map[string]int{
		"/":                     http.StatusOK,
		"/static/css/style.css": http.StatusOK,
		"/static/":              http.StatusNotFound,
		"/static/img/":          http.StatusNotFound,
		"/static/js/":           http.StatusNotFound,
		"/static/css/":          http.StatusNotFound,
	}
	for path, want := range cases {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://example.com"+path, nil))
		if rec.Code != want {
			t.Errorf("GET %s = %d, want %d", path, rec.Code, want)
		}
	}
}

func TestNotFoundContentType(t *testing.T) {
	s, err := New(config.Config{Addr: ":0"}, zerolog.Nop())
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	s.routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope", nil))
	if rec.Code != http.StatusNotFound || rec.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Fatalf("got %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
}

func TestMaxHeaderBytes(t *testing.T) {
	s, err := New(config.Config{Addr: ":0"}, zerolog.Nop())
	if err != nil {
		t.Fatal(err)
	}
	if s.http.MaxHeaderBytes != 16<<10 {
		t.Fatalf("MaxHeaderBytes = %d", s.http.MaxHeaderBytes)
	}
}

func TestClientIPAndLogTruncation(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "203.0.113.9:4444"
	if got := clientIP(r); got != "203.0.113.9:4444" {
		t.Fatalf("no XFF: %q", got)
	}
	r.Header.Set("X-Forwarded-For", "127.0.0.1, spoofed, 198.51.100.7")
	if got := clientIP(r); got != "127.0.0.1, spoofed, 198.51.100.7" {
		t.Fatalf("XFF: %q", got)
	}
	r.Header.Set("X-Forwarded-For", strings.Repeat("a", 5000))
	if got := clientIP(r); len(got) != maxLoggedHeader {
		t.Fatalf("XFF not truncated: %d", len(got))
	}

	var buf bytes.Buffer
	h := requestLogger(zerolog.New(&buf))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	r.Header.Set("User-Agent", strings.Repeat("u", 5000))
	h.ServeHTTP(httptest.NewRecorder(), r)
	if buf.Len() > 2000 {
		t.Fatalf("log line too large: %d", buf.Len())
	}
}
