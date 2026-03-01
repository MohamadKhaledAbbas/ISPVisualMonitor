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

// NewConnection creates a new database connection with retry logic
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
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Verify connection with retry logic
	maxRetries := 30
	retryInterval := 2 * time.Second
	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err != nil {
			log.Printf("Database connection attempt %d/%d failed: %v (retrying in %s)", i+1, maxRetries, err, retryInterval)
			time.Sleep(retryInterval)
			continue
		}
		log.Println("Database connection established and verified")
		return &DB{db}, nil
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts", maxRetries)
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
