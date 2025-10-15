package main

import (
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"

	"go-api-kbt/migrations"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	source, err := iofs.New(migrations.Files, ".")
	if err != nil {
		log.Fatalf("failed to create iofs source: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}

	cmd := os.Args[1]

	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("failed to apply migrations: %v", err)
		}
		log.Println("migrations applied successfully")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("failed to rollback migrations: %v", err)
		}
		log.Println("migrations rolled back successfully")
	case "goto":
		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("invalid version: %v", err)
		}
		if err := m.Migrate(uint(version)); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("failed to migrate to version %d: %v", version, err)
		}
		log.Printf("migrated to version %d successfully", version)
	case "version":
		version, dirty, err := m.Version()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("failed to get version: %v", err)
		}
		log.Printf("version: %d, dirty: %t", version, dirty)
	case "force":
		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("invalid version: %v", err)
		}
		if err := m.Force(version); err != nil {
			log.Fatalf("failed to force version %d: %v", version, err)
		}
		log.Printf("forced version to %d successfully", version)
	default:
		log.Fatalf("unknown command: %s", cmd)
	}
}
