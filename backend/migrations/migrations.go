package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"log"

	"csv-processor/backend/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var migrationFiles embed.FS

func Run(db *sql.DB, cfg *config.Config) error {
	d, err := iofs.New(migrationFiles, ".")
	if err != nil {
		return err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return err
	}

	if dirty {
		log.Println("⚠️  Migration is in dirty state, please fix manually")
	} else if errors.Is(err, migrate.ErrNilVersion) {
		log.Println("No migrations applied yet")
	} else {
		log.Printf("Migrations applied successfully, current version: %d", version)
	}

	return nil
}
