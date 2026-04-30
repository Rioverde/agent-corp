package db

import (
	"context"
	"embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/Rioverde/agent-corp/internal/config"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations applies all embedded Postgres migrations.
func RunMigrations(ctx context.Context, cfg config.DatabaseConfig) error {
	sqlDB, err := goose.OpenDBWithDriver("pgx", cfg.URL)
	if err != nil {
		return fmt.Errorf("open postgres for migrations: %w", err)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.UpContext(ctx, sqlDB, "migrations"); err != nil {
		return fmt.Errorf("run postgres migrations: %w", err)
	}

	return nil
}
