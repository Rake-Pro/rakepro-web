package server

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// NewLogger builds a zerolog.Logger. In development it uses a colorized,
// human-readable console writer; otherwise it emits structured JSON suitable
// for log shippers in Docker/k8s. An unparseable level falls back to info.
func NewLogger(env, level string) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	var l zerolog.Logger
	if env == "development" {
		w := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		l = zerolog.New(w)
	} else {
		l = zerolog.New(os.Stdout)
	}

	return l.Level(lvl).With().Timestamp().Str("service", "rakepro-web").Logger()
}

// statusRecorder wraps http.ResponseWriter to capture the status code and the
// number of bytes written for access logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// requestLogger is middleware that emits one structured access-log line per
// request, including method, path, status, latency, and client address.
func requestLogger(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r)

			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", rec.status).
				Int("bytes", rec.bytes).
				Dur("latency", time.Since(start)).
				Str("remote", clientIP(r)).
				Str("user_agent", truncate(r.UserAgent(), maxLoggedHeader)).
				Msg("request")
		})
	}
}

// clientIP returns the X-Forwarded-For chain as received (cut to
// maxLoggedHeader), else the peer address. Behind several proxies the
// right-most entry is the nearest proxy, not the client; which entries can be
// trusted depends on how each proxy in front writes the header.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return truncate(xff, maxLoggedHeader)
	}
	return r.RemoteAddr
}

// maxLoggedHeader caps client-controlled header values copied into log lines.
const maxLoggedHeader = 256

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
