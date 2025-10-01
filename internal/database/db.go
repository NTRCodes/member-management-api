package database

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

// DB wraps sql.DB
type DB struct {
	*sql.DB
}

// NewDB creates a database connection with optimized connection pooling
func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool for high performance
	db.SetMaxOpenConns(25)        // Maximum number of open connections
	db.SetMaxIdleConns(10)        // Maximum number of idle connections
	db.SetConnMaxLifetime(300)    // Maximum connection lifetime (5 minutes)
	db.SetConnMaxIdleTime(60)     // Maximum idle time (1 minute)

	return &DB{db}, nil
}

// HealthCheck tests if database is reachable
func (db *DB) HealthCheck(ctx context.Context) error {
	// TODO: What should go here?
	// Hint: Use PingContext with the provided context
	return db.PingContext(ctx)
}
