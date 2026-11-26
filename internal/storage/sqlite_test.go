
package storage

import (
	"testing"
)

func TestNewSQLiteDB(t *testing.T) {
	dbPath := t.TempDir() + "/platformpilot.db"

	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize SQLite: %v", err)
	}
	defer db.Close()

	var tableName string

	err = db.QueryRow(`
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
		AND name = 'metrics_samples'
	`).Scan(&tableName)

	if err != nil {
		t.Fatalf("metrics_samples table was not created: %v", err)
	}

	if tableName != "metrics_samples" {
		t.Fatalf("unexpected table name: %s", tableName)
	}
}

