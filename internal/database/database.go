package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"

	"github.com/alperkirkus/fintech-backend/internal/config"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type DB struct {
	*sql.DB
}

func Connect(databaseConfig config.DBConfig) (*DB, error) {
	sqlDatabase, err := sql.Open("pgx", databaseConfig.DSN())
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	sqlDatabase.SetMaxOpenConns(databaseConfig.MaxOpenConns)
	sqlDatabase.SetMaxIdleConns(databaseConfig.MaxIdleConns)
	sqlDatabase.SetConnMaxLifetime(databaseConfig.ConnMaxLifetime)

	pingContext, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()

	if err := sqlDatabase.PingContext(pingContext); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	log.Info().Str("host", databaseConfig.Host).Str("db", databaseConfig.Name).Msg("database connected")

	databaseInstance := &DB{sqlDatabase}

	if err := databaseInstance.runMigrations(); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return databaseInstance, nil
}

func (databaseInstance *DB) runMigrations() error {
	migrationSource, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	migrationDriver, err := postgres.WithInstance(databaseInstance.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", migrationSource, "postgres", migrationDriver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	version, dirty, err := migrator.Version()
	if err != nil {
		log.Warn().Err(err).Msg("could not read migration version")
	} else {
		log.Info().Uint("version", version).Bool("dirty", dirty).Msg("migrations applied")
	}

	return nil
}

func (databaseInstance *DB) Close() error {
	log.Info().Msg("closing database connection pool")
	return databaseInstance.DB.Close()
}
