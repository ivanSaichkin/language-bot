package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"strings"
)

var migrationFiles embed.FS

type Migration struct {
	ID   int
	Name string
	SQL  string
}

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) EnsureMigrationsTable() error {
	query := `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            id SERIAL PRIMARY KEY,
            version INTEGER NOT NULL UNIQUE,
            name VARCHAR(255) NOT NULL,
            applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )
    `
	_, err := m.db.Exec(query)
	return err
}

func (m *Migrator) GetAppliedMigrations() (map[int]bool, error) {
	applied := make(map[int]bool)

	query := `SELECT version FROM schema_migrations ORDER BY version`
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, nil
}

func (m *Migrator) LoadMigrations() ([]Migration, error) {
	var migrations []Migration

	files, err := migrationFiles.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read migration files: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			var version int
			var name string
			_, err := fmt.Sscanf(file.Name(), "%d_%s", &version, &name)
			if err != nil {
				log.Printf("Warning: skipping invalid migration file: %s", file.Name())
				continue
			}

			content, err := migrationFiles.ReadFile(file.Name())
			if err != nil {
				return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
			}

			migrations = append(migrations, Migration{
				ID:   version,
				Name: strings.TrimSuffix(name, ".sql"),
				SQL:  string(content),
			})
		}
	}

	return migrations, nil
}

func (m *Migrator) RunMigrations() error {
	log.Println("🔄 Running database migrations...")

	if err := m.EnsureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	migrations, err := m.LoadMigrations()
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if applied[migration.ID] {
			log.Printf("✅ Migration %d_%s already applied", migration.ID, migration.Name)
			continue
		}

		log.Printf("🔄 Applying migration %d_%s...", migration.ID, migration.Name)

		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(migration.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %d_%s: %w", migration.ID, migration.Name, err)
		}

		insertQuery := `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`
		if _, err := tx.Exec(insertQuery, migration.ID, migration.Name); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d_%s: %w", migration.ID, migration.Name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d_%s: %w", migration.ID, migration.Name, err)
		}

		log.Printf("✅ Successfully applied migration %d_%s", migration.ID, migration.Name)
	}

	log.Println("🎉 All migrations completed successfully!")
	return nil
}

func (m *Migrator) GetDatabaseVersion() (int, error) {
	var version int
	query := `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`
	err := m.db.QueryRow(query).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get database version: %w", err)
	}
	return version, nil
}
