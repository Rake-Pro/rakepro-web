package server

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

// testMonitor returns a monitor for statusURL whose transport sends every
// request to srv, whatever host the URL names.
func testMonitor(t *testing.T, srv *httptest.Server, statusURL string) *streamMonitor {
	t.Helper()
	m := newStreamMonitor(statusURL, time.Hour, zerolog.Nop())
	addr := srv.Listener.Addr().String()
	m.client.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		},
	}
	return m
}

func TestStreamMonitorOnline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"online":true}`))
	}))
	defer srv.Close()

	m := testMonitor(t, srv, "http://status.example.com/api/status")
	m.poll(context.Background())
	if !m.Online() {
		t.Fatal("online = false, want true")
	}
}

func TestStreamMonitorStatusBodyLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"online":true,"pad":"` + strings.Repeat("a", 2*maxStatusBody) + `"}`))
	}))
	defer srv.Close()

	m := testMonitor(t, srv, "http://status.example.com/api/status")
	m.online.Store(true)
	m.poll(context.Background())
	if m.Online() {
		t.Fatal("status body larger than the limit was decoded")
	}
}

func TestStreamMonitorRedirectSameHost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/old" {
			http.Redirect(w, r, "http://status.example.com:8080/api/status", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte(`{"online":true}`))
	}))
	defer srv.Close()

	m := testMonitor(t, srv, "http://status.example.com/old")
	m.poll(context.Background())
	if !m.Online() {
		t.Fatal("same-host redirect was not followed")
	}
}

func TestStreamMonitorRedirectOtherHostRefused(t *testing.T) {
	var otherHits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "status.example.com" {
			http.Redirect(w, r, "http://other.example.net/api/status", http.StatusFound)
			return
		}
		otherHits++
		_, _ = w.Write([]byte(`{"online":true}`))
	}))
	defer srv.Close()

	m := testMonitor(t, srv, "http://status.example.com/api/status")
	m.poll(context.Background())
	if otherHits != 0 || m.Online() {
		t.Fatalf("redirect to another host followed: hits=%d online=%t", otherHits, m.Online())
	}
}

func TestStreamMonitorRedirectLimit(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		http.Redirect(w, r, "/loop", http.StatusFound)
	}))
	defer srv.Close()

	m := testMonitor(t, srv, "http://status.example.com/api/status")
	m.poll(context.Background())
	if want := maxStatusRedirects + 1; hits != want || m.Online() {
		t.Fatalf("hits=%d online=%t, want hits=%d online=false", hits, m.Online(), want)
	}
}
