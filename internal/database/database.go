package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/config"
	_ "github.com/lib/pq"
)

// DB wraps the database connection
type DB struct {
	*sql.DB
}

// NewConnection creates a new database connection
func NewConnection(cfg config.DatabaseConfig) (*DB, error) {
	connStr := cfg.ConnectionString()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MinConns)
	db.SetConnMaxLifetime(time.Hour)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection pool configured")

	return &DB{db}, nil
}

// SetTenantContext sets the tenant context for row-level security
func (db *DB) SetTenantContext(tenantID string) error {
	_, err := db.Exec("SET app.current_tenant = $1", tenantID)
	return err
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
