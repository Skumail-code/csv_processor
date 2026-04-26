package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"csv-processor/backend/internal/config"

	_ "github.com/lib/pq"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	var db *sql.DB
	var err error

	// Retry logic for database connection
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", cfg.GetDBConnString())
		if err != nil {
			log.Printf("Failed to open database connection (attempt %d/5): %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Configure connection pool
		db.SetMaxOpenConns(cfg.DBMaxOpenConns)
		db.SetMaxIdleConns(cfg.DBMaxIdleConns)
		db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = db.PingContext(ctx)
		cancel()

		if err == nil {
			log.Println("Database connected successfully")
			return &Database{DB: db}, nil
		}

		log.Printf("Failed to ping database (attempt %d/5): %v", i+1, err)
		db.Close()
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after 5 attempts: %w", err)
}

func (d *Database) Close() error {
	return d.DB.Close()
}

func (d *Database) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return d.DB.PingContext(ctx)
}
