package db

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/IfedayoAwe/payment-processing-service/config"
	"github.com/IfedayoAwe/payment-processing-service/db/gen"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var logFatalf = log.Fatalf

type Dependencies struct {
	OpenDB             func(driverName, dataSourceName string) (*sql.DB, error)
	PingDB             func(db *sql.DB) error
	FileExists         func(path string) bool
	NewMigrationDriver func(db *sql.DB) (database.Driver, error)
	NewMigrator        func(path string, drv database.Driver) (Migrator, error)
}

type Migrator interface {
	Version() (uint, bool, error)
	Force(int) error
	Up() error
}

var DefaultDependencies = Dependencies{
	OpenDB: sql.Open,
	PingDB: func(db *sql.DB) error { return db.Ping() },
	FileExists: func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	},
	NewMigrationDriver: func(db *sql.DB) (database.Driver, error) {
		return postgres.WithInstance(db, &postgres.Config{})
	},
	NewMigrator: func(path string, drv database.Driver) (Migrator, error) {
		return migrate.NewWithDatabaseInstance("file://"+path, "postgres", drv)
	},
}

func InitDBWithDeps(cfg config.Config, deps Dependencies) (*gen.Queries, *sql.DB) {
	dbConn := openAndPingDB(cfg.DatabaseURL, deps)
	migrationsPath := findMigrationsPath(deps)

	drv := setupMigrationDriverWithRetry(dbConn, deps)
	migrator := setupMigratorWithRetry(migrationsPath, drv, deps)

	applyMigrations(migrator)

	log.Println("Migrations applied successfully. SQLC client initialized.")
	return gen.New(dbConn), dbConn
}

func InitDB(cfg config.Config) (*gen.Queries, *sql.DB) {
	return InitDBWithDeps(cfg, DefaultDependencies)
}

func openAndPingDB(dsn string, deps Dependencies) *sql.DB {
	const maxAttempts = 10
	const retryDelay = 2 * time.Second

	dbConn, err := deps.OpenDB("postgres", dsn)
	if err != nil {
		logFatalf("db.Open failed: %v", err)
	}

	for i := 1; i <= maxAttempts; i++ {
		if err = deps.PingDB(dbConn); err == nil {
			return dbConn
		}
		log.Printf("Database ping failed (attempt %d): %v; retrying in %s...", i, err, retryDelay)
		time.Sleep(retryDelay)
	}

	logFatalf("Could not connect to database after %d attempts: %v", maxAttempts, err)
	return nil // unreachable but required for compilation
}

func findMigrationsPath(deps Dependencies) string {
	wd, err := os.Getwd()
	if err != nil {
		logFatalf("could not get working directory: %v", err)
	}

	candidate := filepath.Join(wd, "migrations")
	if deps.FileExists(candidate) {
		return candidate
	}

	repoRoot := wd
	for !deps.FileExists(filepath.Join(repoRoot, "go.mod")) && repoRoot != "/" {
		repoRoot = filepath.Dir(repoRoot)
	}
	return filepath.Join(repoRoot, "migrations")
}

func setupMigrationDriverWithRetry(db *sql.DB, deps Dependencies) database.Driver {
	const maxAttempts = 10
	const retryDelay = 2 * time.Second

	var (
		drv database.Driver
		err error
	)

	for i := 1; i <= maxAttempts; i++ {
		drv, err = deps.NewMigrationDriver(db)
		if err == nil {
			return drv
		}
		log.Printf("postgres.WithInstance failed (attempt %d): %v; retrying in %s...", i, err, retryDelay)
		time.Sleep(retryDelay)
	}

	logFatalf("postgres.WithInstance failed after retries: %v", err)
	return nil // unreachable but required for compilation
}

func setupMigratorWithRetry(path string, drv database.Driver, deps Dependencies) Migrator {
	const maxAttempts = 10
	const retryDelay = 2 * time.Second

	var (
		m   Migrator
		err error
	)

	for i := 1; i <= maxAttempts; i++ {
		m, err = deps.NewMigrator(path, drv)
		if err == nil {
			return m
		}
		log.Printf("migrate.NewWithDatabaseInstance failed (attempt %d): %v; retrying in %s...", i, err, retryDelay)
		time.Sleep(retryDelay)
	}

	logFatalf("migrate.NewWithDatabaseInstance failed after retries: %v", err)
	return nil // unreachable but required for compilation
}

func applyMigrations(m Migrator) {
	if m == nil {
		logFatalf("Migration instance is nil")
	}

	if version, dirty, err := m.Version(); err == nil && dirty {
		log.Printf("Detected dirty migration version %d; forcing state to %d", version, version)
		if fErr := m.Force(int(version)); fErr != nil {
			logFatalf("migrate.Force(%d) failed: %v", version, fErr)
		}
	} else if err != nil {
		log.Printf("migrator.Version failed: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logFatalf("migrate.Up failed: %v", err)
	}
}
