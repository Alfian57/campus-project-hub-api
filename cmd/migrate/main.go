package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/campus-project-hub/api/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Parse command line arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	action := os.Args[1]
	
	// Parse additional flags
	forceFlag := flag.NewFlagSet("force", flag.ExitOnError)
	forceVersion := forceFlag.Int("version", -1, "Force set version (use with care!)")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Build database URL
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	// Initialize migrate instance
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatalf("❌ Failed to initialize migration: %v", err)
	}
	defer m.Close()

	switch action {
	case "up":
		log.Println("🔄 Running migrations (up)...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("❌ Migration failed: %v", err)
		}
		log.Println("✅ Migration completed successfully!")

	case "down":
		log.Println("🔄 Rolling back last migration...")
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("❌ Rollback failed: %v", err)
		}
		log.Println("✅ Rollback completed successfully!")

	case "fresh":
		log.Println("⚠️  WARNING: Dropping all tables and re-migrating...")
		if err := m.Drop(); err != nil {
			log.Printf("⚠️  Drop warning: %v", err)
		}
		// Re-initialize after drop
		m, err = migrate.New("file://migrations", dbURL)
		if err != nil {
			log.Fatalf("❌ Failed to reinitialize migration: %v", err)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("❌ Migration failed: %v", err)
		}
		log.Println("✅ Fresh migration completed successfully!")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Printf("ℹ️  Version info: %v", err)
		} else {
			log.Printf("📋 Current version: %d (dirty: %v)", version, dirty)
		}

	case "force":
		forceFlag.Parse(os.Args[2:])
		if *forceVersion < 0 {
			log.Fatalf("❌ Please specify version with -version flag")
		}
		if err := m.Force(*forceVersion); err != nil {
			log.Fatalf("❌ Force failed: %v", err)
		}
		log.Printf("✅ Forced version to %d", *forceVersion)

	default:
		log.Printf("❌ Unknown action: %s", action)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`
Campus Project Hub API - Migration Tool

Usage: ./migrate <action> [options]

Actions:
  up        Run all pending migrations
  down      Rollback the last migration
  fresh     Drop all tables and run all migrations (WARNING: destroys data!)
  version   Show current migration version
  force     Force set migration version (use with care!)
            Options: -version=<N>

Examples:
  ./migrate up
  ./migrate down
  ./migrate fresh
  ./migrate version
  ./migrate force -version=3
`)
}
