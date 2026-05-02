package app

import (
	"encoding/json"
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


	t.Run("tutorial page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tutorial", nil)
		rr := httptest.NewRecorder()
		application.Router().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "ptto Onboarding Tutorial") {
			t.Fatalf("expected tutorial heading in body, got %q", body)
		}
		if !strings.Contains(body, "ptto init") || !strings.Contains(body, "ptto deploy") {
			t.Fatalf("expected tutorial commands in body, got %q", body)
		}
	})

	t.Run("tutorial method check", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/tutorial", nil)
		rr := httptest.NewRecorder()
		application.Router().ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
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

	t.Run("webauthn start method check", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/webauthn/register/start", nil)
		rr := httptest.NewRecorder()
		application.Router().ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("webauthn start returns json and cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webauthn/register/start", nil)
		rr := httptest.NewRecorder()
		application.Router().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		var payload map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if payload["status"] != "ok" {
			t.Fatalf("unexpected status payload: %v", payload["status"])
		}
		if _, ok := payload["publicKey"]; !ok {
			t.Fatalf("missing publicKey in payload: %v", payload)
		}
		found := false
		for _, c := range rr.Result().Cookies() {
			if c.Name == registerCookieName && c.Value != "" {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing registration cookie")
		}
	})

	t.Run("webauthn finish invalid payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webauthn/register/finish", strings.NewReader(`{"bad":true}`))
		rr := httptest.NewRecorder()
		application.Router().ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})
}
