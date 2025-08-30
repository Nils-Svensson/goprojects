package server

import (
	"database/sql"
	"fmt"

	"goprojects/findings"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS findings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		namespace TEXT,
		resource TEXT,
		kind TEXT,
		container TEXT,
		issue TEXT,
		suggestion TEXT,
		subjects TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return db, nil
}

func InsertFinding(db *sql.DB, f findings.Finding) error {
	_, err := db.Exec(`
		INSERT INTO findings (namespace, resource, kind, container, issue, suggestion, subjects)
		VALUES (?, ?, ?, ?, ?, ?)`,
		f.Namespace, f.Resource, f.Kind, f.Container, f.Issue, f.Suggestion, f.Subjects,
	)
	return err
}
