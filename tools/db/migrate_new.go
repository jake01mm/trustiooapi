package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"trusioo_api/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Parse command line arguments
	var (
		action    = flag.String("action", "up", "Migration action: up, down, drop, version, force, create")
		steps     = flag.Int("steps", 0, "Number of migration steps (0 = all)")
		version   = flag.Int("version", 0, "Target migration version")
		name      = flag.String("name", "", "Migration name for create action")
		force     = flag.Int("force", 0, "Force migration to specific version (dangerous)")
	)
	flag.Parse()

	switch *action {
	case "create":
		if *name == "" {
			log.Fatal("Migration name is required for create action. Use -name flag")
		}
		createMigration(*name)
	case "up", "down", "drop", "version", "force":
		runMigration(*action, *steps, *version, *force)
	default:
		fmt.Println("Available actions:")
		fmt.Println("  up      - Run pending migrations")
		fmt.Println("  down    - Rollback migrations")
		fmt.Println("  drop    - Drop all tables and reset")
		fmt.Println("  version - Show current migration version")
		fmt.Println("  force   - Force migration to specific version (dangerous)")
		fmt.Println("  create  - Create new migration files")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run migrate_new.go -action=up")
		fmt.Println("  go run migrate_new.go -action=down -steps=1")
		fmt.Println("  go run migrate_new.go -action=create -name=\"add_user_preferences\"")
		fmt.Println("  go run migrate_new.go -action=version")
		os.Exit(1)
	}
}

func createMigration(name string) {
	// Get next migration number
	migrationsDir := "../../migrations"
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	nextVersion := len(files) + 1
	timestamp := time.Now().Format("20060102150405")
	
	// Create migration files
	upFile := fmt.Sprintf("%s/%06d_%s_%s.up.sql", migrationsDir, nextVersion, timestamp, name)
	downFile := fmt.Sprintf("%s/%06d_%s_%s.down.sql", migrationsDir, nextVersion, timestamp, name)

	// Create up migration file
	upTemplate := fmt.Sprintf(`-- Migration: %s (UP)
-- Created: %s
-- Description: Add your migration logic here

-- Example:
-- ALTER TABLE users ADD COLUMN new_field VARCHAR(100) DEFAULT '';
-- CREATE INDEX IF NOT EXISTS idx_users_new_field ON users(new_field);

-- Your migration SQL goes here...

`, name, time.Now().Format("2006-01-02 15:04:05"))

	err = os.WriteFile(upFile, []byte(upTemplate), 0644)
	if err != nil {
		log.Fatalf("Failed to create up migration file: %v", err)
	}

	// Create down migration file
	downTemplate := fmt.Sprintf(`-- Migration: %s (DOWN/ROLLBACK)
-- Created: %s
-- Description: Rollback changes from the corresponding up migration

-- Example:
-- DROP INDEX IF EXISTS idx_users_new_field;
-- ALTER TABLE users DROP COLUMN IF EXISTS new_field;

-- Your rollback SQL goes here...

`, name, time.Now().Format("2006-01-02 15:04:05"))

	err = os.WriteFile(downFile, []byte(downTemplate), 0644)
	if err != nil {
		log.Fatalf("Failed to create down migration file: %v", err)
	}

	fmt.Printf("✅ Migration files created:\n")
	fmt.Printf("   📄 %s\n", upFile)
	fmt.Printf("   📄 %s\n", downFile)
	fmt.Printf("\n💡 Edit these files and then run: go run migrate_new.go -action=up\n")
}

func runMigration(action string, steps, version, forceVersion int) {
	cfg := config.AppConfig.Database

	// Create database connection string
	dbURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=require",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create postgres driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create postgres driver: %v", err)
	}

	// Create migrate instance
	migrationsPath := "file://../../migrations"
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	switch action {
	case "up":
		fmt.Println("🚀 Running pending migrations...")
		if steps > 0 {
			err = m.Steps(steps)
		} else {
			err = m.Up()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration failed: %v", err)
		}
		if err == migrate.ErrNoChange {
			fmt.Println("📋 No pending migrations")
		} else {
			fmt.Println("✅ Migrations completed successfully")
		}

	case "down":
		fmt.Printf("⬇️  Rolling back migrations (steps: %d)...\n", steps)
		if steps > 0 {
			err = m.Steps(-steps)
		} else {
			err = m.Down()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Rollback failed: %v", err)
		}
		if err == migrate.ErrNoChange {
			fmt.Println("📋 No migrations to rollback")
		} else {
			fmt.Println("✅ Rollback completed successfully")
		}

	case "version":
		currentVersion, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		fmt.Printf("📊 Current migration version: %d\n", currentVersion)
		if dirty {
			fmt.Println("⚠️  Database is in dirty state - migration may have failed")
		}

	case "force":
		if forceVersion == 0 {
			log.Fatal("Force version is required. Use -force flag")
		}
		fmt.Printf("⚡ Forcing migration to version %d...\n", forceVersion)
		err = m.Force(forceVersion)
		if err != nil {
			log.Fatalf("Force migration failed: %v", err)
		}
		fmt.Println("✅ Forced migration completed")

	case "drop":
		fmt.Println("🗑️  Dropping all tables...")
		err = m.Drop()
		if err != nil {
			log.Fatalf("Drop failed: %v", err)
		}
		fmt.Println("✅ All tables dropped successfully")

	default:
		log.Fatalf("Unknown action: %s", action)
	}

	// Show final status
	currentVersion, dirty, err := m.Version()
	if err == nil {
		fmt.Printf("📈 Final database version: %d", currentVersion)
		if dirty {
			fmt.Println(" (dirty)")
		} else {
			fmt.Println(" (clean)")
		}
	}
}