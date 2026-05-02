package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestShowcaseRoutes(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data.sqlite")
	application, err := NewWithDBPath(dbPath)
	if err != nil {
		t.Fatalf("NewWithDBPath() error = %v", err)
	}
	t.Cleanup(func() { _ = application.Close() })

	t.Run("hero page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		application.Router().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "Rendered in") {
			t.Fatalf("expected render timing in body, got %q", body)
		}
		if !strings.Contains(body, `hx-post="/latency-ping"`) {
			t.Fatalf("expected latency ping button, got %q", body)
		}
		if !strings.Contains(body, `hx-trigger="load, every 5s"`) {
			t.Fatalf("expected resource monitor polling, got %q", body)
		}
	})

	t.Run("latency ping", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/latency-ping", nil)
		rr := httptest.NewRecorder()
		application.Router().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		body, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if !strings.Contains(string(body), "SQLite time:") {
			t.Fatalf("unexpected body = %q", string(body))
		}
	})

	t.Run("resource monitor", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resource-monitor", nil)
		rr := httptest.NewRecorder()
		application.Router().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "MB") {
			t.Fatalf("unexpected body = %q", body)
		}
	})
}
