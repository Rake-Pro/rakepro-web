package server

import (
	"net/http"
	"net/http/httptest"
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
