// Package server wires together routing, middleware, templates, and the
// lifecycle of the HTTP server.
package server

import (
	"context"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/rakepro/rakepro-web/internal/config"
	"github.com/rakepro/rakepro-web/web"
)

// version is overridden at build time via -ldflags. See the Makefile.
var version = "dev"

// Server holds the dependencies shared across handlers.
type Server struct {
	cfg    config.Config
	log    zerolog.Logger
	tmpl   *template.Template
	http   *http.Server
	start  time.Time
	stream *streamMonitor
}

// New constructs a Server, parsing the embedded templates up front so a
// malformed template fails fast at startup rather than on first request.
func New(cfg config.Config, log zerolog.Logger) (*Server, error) {
	tmpl, err := template.ParseFS(web.Templates, "templates/*.html")
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg:   cfg,
		log:   log,
		tmpl:  tmpl,
		start: time.Now(),
	}

	// Only monitor the stream when a status endpoint is configured.
	if cfg.StreamStatusURL != "" {
		s.stream = newStreamMonitor(cfg.StreamStatusURL, cfg.StreamPollInterval, log)
	}

	s.http = &http.Server{
		Addr:         cfg.Addr,
		Handler:      s.routes(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		ErrorLog:     nil,
	}

	return s, nil
}

// routes builds the http.Handler with all routes and global middleware.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Static assets served straight from the embedded FS, rooted at /static/.
	staticFS, _ := fs.Sub(web.Static, "static")
	fileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	// Liveness and readiness probes for Docker healthchecks and k8s.
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /readyz", s.handleHealth)

	// Homepage. The trailing-slash root pattern also catches 404s, which we
	// handle explicitly to avoid serving the homepage for unknown paths.
	mux.HandleFunc("GET /", s.handleHome)

	return requestLogger(s.log)(securityHeaders(mux))
}

// securityHeaders sets baseline security headers on every response. The CSP
// allowlist mirrors the page's real asset origins: everything is same-origin
// except the favicon/logo assets loaded from cdn.rake.pro; there are no inline
// scripts or styles, no XHR, and the page is never framed.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy",
			"default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self' https://cdn.rake.pro; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// handleHome renders the homepage template. Unknown paths under "/" yield 404.
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.handleNotFound(w, r)
		return
	}

	data := map[string]any{
		"Year":         time.Now().Year(),
		"Version":      version,
		"StreamURL":    s.cfg.StreamURL,
		"StreamOnline": s.stream != nil && s.stream.Online(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		s.log.Error().Err(err).Msg("render homepage")
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("404 - not found"))
}

// handleHealth reports process liveness and uptime as JSON.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	uptime := time.Since(s.start).Round(time.Second)
	_, _ = w.Write([]byte(`{"status":"ok","version":"` + version + `","uptime":"` + uptime.String() + `"}`))
}

// Run starts the server and blocks until ctx is cancelled, then shuts down
// gracefully within the configured ShutdownTimeout.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	if s.stream != nil {
		go s.stream.run(ctx)
	}

	go func() {
		s.log.Info().
			Str("addr", s.cfg.Addr).
			Str("env", s.cfg.Env).
			Str("version", version).
			Msg("server listening")
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.log.Info().Msg("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		s.log.Error().Err(err).Msg("graceful shutdown failed; forcing close")
		return s.http.Close()
	}

	s.log.Info().Msg("server stopped cleanly")
	return nil
}
