package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Migration struct {
	Version string
	Name    string
	SQL     string
}

type Migrator struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Migrator {
	return &Migrator{pool: pool}
}

func (m *Migrator) Up(ctx context.Context) error {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return err
	}

	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return err
	}

	appliedSet := make(map[string]bool)
	for _, v := range applied {
		appliedSet[v] = true
	}

	for _, mig := range migrations {
		if appliedSet[mig.Version] {
			log.Info().Str("version", mig.Version).Str("name", mig.Name).Msg("Migration already applied")
			continue
		}

		if err := m.applyMigration(ctx, mig); err != nil {
			return err
		}
	}

	log.Info().Msg("All migrations applied successfully")
	return nil
}

func (m *Migrator) ensureMigrationsTable(ctx context.Context) error {
	sql := `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_schema_migrations_applied_at ON schema_migrations(applied_at);`

	_, err := m.pool.Exec(ctx, sql)
	return err
}

func (m *Migrator) getAppliedVersions(ctx context.Context) ([]string, error) {
	rows, err := m.pool.Query(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}

	return versions, rows.Err()
}

func (m *Migrator) applyMigration(ctx context.Context, mig Migration) error {
	log.Info().Str("version", mig.Version).Str("name", mig.Name).Msg("Applying migration")

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, mig.SQL); err != nil {
		return fmt.Errorf("failed to execute migration %s: %w", mig.Version, err)
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, $3)",
		mig.Version, mig.Name, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to record migration %s: %w", mig.Version, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	log.Info().Str("version", mig.Version).Str("name", mig.Name).Msg("Migration applied successfully")
	return nil
}

func (m *Migrator) Status(ctx context.Context) error {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return err
	}

	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return err
	}

	appliedSet := make(map[string]bool)
	for _, v := range applied {
		appliedSet[v] = true
	}

	log.Info().Msg("Migration status:")
	for _, mig := range migrations {
		status := "pending"
		if appliedSet[mig.Version] {
			status = "applied"
		}
		log.Info().Str("version", mig.Version).Str("name", mig.Name).Str("status", status).Msg("-")
	}

	return nil
}

func loadMigrations() ([]Migration, error) {
	migrationsDir := "migrations"

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		altDir := filepath.Join("..", "..", "migrations")
		entries, err = os.ReadDir(altDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read migrations directory: %w", err)
		}
		migrationsDir = altDir
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		version, migrationName := parseFileName(name)
		if version == "" {
			continue
		}

		sqlBytes, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    migrationName,
			SQL:     string(sqlBytes),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func parseFileName(name string) (version, migrationName string) {
	parts := strings.SplitN(strings.TrimSuffix(name, ".sql"), "_", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
