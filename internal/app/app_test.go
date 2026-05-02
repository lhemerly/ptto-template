package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestNewWithDBPathAndRouter(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data.sqlite")
	application, err := NewWithDBPath(dbPath)
	if err != nil {
		t.Fatalf("NewWithDBPath() error = %v", err)
	}
	t.Cleanup(func() { _ = application.Close() })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	application.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(body) != "ptto-template core engine is up\n" {
		t.Fatalf("body = %q", string(body))
	}
}
