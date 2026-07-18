package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// DB wraps a sql.DB with the application context.
type DB struct {
	*sql.DB
}

// New opens a connection pool to the Postgres database at dsn.
func New(dsn string) (*DB, error) {
	pool, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return &DB{DB: pool}, nil
}
