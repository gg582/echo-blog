// Package database opens the SQLite database that stores the admin account.
package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const createUsersTable = `CREATE TABLE IF NOT EXISTS blog_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);`

// Open opens the SQLite database at path and creates the users table if it
// does not exist yet.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open database %s: %w", path, err)
	}
	if _, err := db.Exec(createUsersTable); err != nil {
		db.Close()
		return nil, fmt.Errorf("create users table: %w", err)
	}
	return db, nil
}
