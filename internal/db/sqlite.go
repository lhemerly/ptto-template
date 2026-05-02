package db

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	values := url.Values{}
	values.Add("_pragma", "busy_timeout(5000)")
	values.Add("_pragma", "journal_mode(WAL)")
	values.Add("_pragma", "synchronous(NORMAL)")
	values.Add("_pragma", "foreign_keys(ON)")

	dsn := fmt.Sprintf("%s?%s", path, values.Encode())

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite failed with dsn %q: %w", dsn, err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	display_name TEXT NOT NULL,
	webauthn_id BLOB NOT NULL UNIQUE,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);`,
		`CREATE TABLE IF NOT EXISTS credentials (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	credential_id BLOB NOT NULL UNIQUE,
	public_key BLOB NOT NULL,
	aaguid BLOB NOT NULL,
	sign_count INTEGER NOT NULL DEFAULT 0,
	transports TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);`,
		`CREATE TABLE IF NOT EXISTS sessions (
	id TEXT PRIMARY KEY,
	user_id INTEGER NOT NULL,
	credential_id BLOB NOT NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	expires_at TEXT NOT NULL,
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY(credential_id) REFERENCES credentials(credential_id) ON DELETE CASCADE
);`,
		`CREATE TABLE IF NOT EXISTS registration_states (
	id TEXT PRIMARY KEY,
	user_id INTEGER NOT NULL,
	session_data BLOB NOT NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	expires_at TEXT NOT NULL,
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);`,
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("apply schema statement: %w", err)
		}
	}
	return tx.Commit()
}