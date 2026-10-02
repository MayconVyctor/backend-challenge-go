package database

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dbUrl string) error {

	log.Println("check database migrations")

	m, err := migrate.New("file://migrations", dbUrl)
	if err != nil {
		return fmt.Errorf("error initializing migrations: %w", err)
	}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("the database is already updated")
			return nil
		}
		return fmt.Errorf("error applying migrations:: %w", err)
	}

	log.Println("migrations successfully applied")
	return nil
}
