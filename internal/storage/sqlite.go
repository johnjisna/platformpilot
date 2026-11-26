
package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("database path cannot be empty")
	}

	dir := filepath.Dir(dbPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}

	// SQLite supports concurrent readers but only one writer at a time.
	// A single connection keeps this initial local implementation simple.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to SQLite database: %w", err)
	}

	if err := createSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func createSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS metrics_samples (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		collected_at DATETIME NOT NULL,
		resource_type TEXT NOT NULL,
		resource_name TEXT NOT NULL,
		namespace TEXT,
		cpu_millicores INTEGER NOT NULL,
		memory_bytes INTEGER NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_metrics_resource_time
	ON metrics_samples(resource_type, resource_name, collected_at);

	CREATE INDEX IF NOT EXISTS idx_metrics_namespace_time
	ON metrics_samples(namespace, collected_at);
	`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("create metrics schema: %w", err)
	}

	return nil
}

