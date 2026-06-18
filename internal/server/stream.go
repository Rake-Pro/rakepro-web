package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

// streamMonitor polls the rakecast status endpoint in the background and caches
// whether the stream is currently live. The cached value is read on every page
// render, so a render never blocks on (or is coupled to) rakecast availability.
// It fails closed: any error polling the endpoint marks the stream offline.
type streamMonitor struct {
	statusURL string
	interval  time.Duration
	client    *http.Client
	log       zerolog.Logger
	online    atomic.Bool
}

func newStreamMonitor(statusURL string, interval time.Duration, log zerolog.Logger) *streamMonitor {
	return &streamMonitor{
		statusURL: statusURL,
		interval:  interval,
		// Short timeout: the upstream check must never stall the poll loop.
		client: &http.Client{Timeout: 4 * time.Second},
		log:    log.With().Str("component", "stream-monitor").Logger(),
	}
}

// Online reports the last-known live status of the stream.
func (m *streamMonitor) Online() bool { return m.online.Load() }

// run polls immediately, then on every interval tick until ctx is cancelled.
func (m *streamMonitor) run(ctx context.Context) {
	m.poll(ctx)

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.poll(ctx)
		}
	}
}

// poll fetches the status endpoint once and updates the cached flag. Errors are
// logged at debug level and treated as offline.
func (m *streamMonitor) poll(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.statusURL, nil)
	if err != nil {
		m.set(false, err)
		return
	}

	resp, err := m.client.Do(req)
	if err != nil {
		m.set(false, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.log.Debug().Int("status", resp.StatusCode).Msg("stream status non-200; treating as offline")
		m.online.Store(false)
		return
	}

	var body struct {
		Online bool `json:"online"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		m.set(false, err)
		return
	}

	prev := m.online.Swap(body.Online)
	if prev != body.Online {
		m.log.Info().Bool("online", body.Online).Msg("stream status changed")
	}
}

func (m *streamMonitor) set(online bool, err error) {
	if err != nil {
		m.log.Debug().Err(err).Msg("stream status poll failed; treating as offline")
	}
	m.online.Store(online)
}
