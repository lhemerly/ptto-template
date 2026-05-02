package db

import (
	"path/filepath"
	"testing"
)

func TestOpenCreatesSQLiteFileAndUsesWAL(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data.sqlite")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	var mode string
	if err := database.QueryRow("PRAGMA journal_mode;").Scan(&mode); err != nil {
		t.Fatalf("PRAGMA journal_mode query failed: %v", err)
	}
	if mode != "wal" {
		t.Fatalf("journal_mode = %q, want %q", mode, "wal")
	}

	var busyTimeout int
	if err := database.QueryRow("PRAGMA busy_timeout;").Scan(&busyTimeout); err != nil {
		t.Fatalf("PRAGMA busy_timeout query failed: %v", err)
	}
	if busyTimeout != 5000 {
		t.Fatalf("busy_timeout = %d, want %d", busyTimeout, 5000)
	}
}
