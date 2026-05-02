package db

import (
	"path/filepath"
	"testing"
)

func TestOpenAppliesStartupPragmas(t *testing.T) {
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

	var synchronous int
	if err := database.QueryRow("PRAGMA synchronous;").Scan(&synchronous); err != nil {
		t.Fatalf("PRAGMA synchronous query failed: %v", err)
	}
	if synchronous != 1 {
		t.Fatalf("synchronous = %d, want %d (NORMAL)", synchronous, 1)
	}

	var foreignKeys int
	if err := database.QueryRow("PRAGMA foreign_keys;").Scan(&foreignKeys); err != nil {
		t.Fatalf("PRAGMA foreign_keys query failed: %v", err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign_keys = %d, want %d", foreignKeys, 1)
	}
}
