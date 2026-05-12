package bootstrap

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const baselineMigrationVersion = "baseline_current_schema"

type migrationFile struct {
	Version  string
	Path     string
	Checksum string
}

type MigrationStatus struct {
	TotalVersioned int
	Applied        int
	Pending        int
	Mismatched     int
}

func RunMigrations(dsn, repoRoot string) error {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := ensureSchemaMigrations(db); err != nil {
		return err
	}

	schemaPath := filepath.Join(repoRoot, "internal", "infrastructure", "persistence", "postgres", "schema", "current_schema.sql")
	files, err := loadMigrationFiles(filepath.Join(repoRoot, "internal", "infrastructure", "persistence", "postgres", "migrations"))
	if err != nil {
		return err
	}
	applied, err := appliedMigrations(db)
	if err != nil {
		return err
	}
	if strictMigrationChecksum() {
		if err := validateMigrationChecksums(files, applied); err != nil {
			return err
		}
	}

	if _, ok := applied[baselineMigrationVersion]; !ok {
		empty, err := publicSchemaEmpty(db)
		if err != nil {
			return err
		}
		if empty {
			if err := execSQLFile(db, schemaPath); err != nil {
				return fmt.Errorf("apply baseline schema: %w", err)
			}
		}
		if err := markMigrationApplied(db, baselineMigrationVersion, checksumFile(schemaPath)); err != nil {
			return err
		}
		applied[baselineMigrationVersion] = checksumFile(schemaPath)
		if empty {
			for _, file := range files {
				if err := markMigrationApplied(db, file.Version, file.Checksum); err != nil {
					return err
				}
				applied[file.Version] = file.Checksum
			}
		}
	}

	for _, file := range files {
		if _, ok := applied[file.Version]; ok {
			continue
		}
		if err := execSQLFile(db, file.Path); err != nil {
			return fmt.Errorf("apply migration %s: %w", file.Version, err)
		}
		if err := markMigrationApplied(db, file.Version, file.Checksum); err != nil {
			return err
		}
	}
	return nil
}

func MigrationStatusFor(dsn, repoRoot string) (MigrationStatus, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return MigrationStatus{}, err
	}
	if err := ensureSchemaMigrations(db); err != nil {
		return MigrationStatus{}, err
	}
	files, err := loadMigrationFiles(filepath.Join(repoRoot, "internal", "infrastructure", "persistence", "postgres", "migrations"))
	if err != nil {
		return MigrationStatus{}, err
	}
	applied, err := appliedMigrations(db)
	if err != nil {
		return MigrationStatus{}, err
	}
	status := MigrationStatus{TotalVersioned: len(files)}
	for _, file := range files {
		checksum, ok := applied[file.Version]
		if !ok {
			status.Pending++
			continue
		}
		status.Applied++
		if strictMigrationChecksum() && checksum != "" && checksum != file.Checksum {
			status.Mismatched++
		}
	}
	return status, nil
}

func ensureSchemaMigrations(db *gorm.DB) error {
	return db.Exec(`
CREATE TABLE IF NOT EXISTS public.schema_migrations (
  version varchar(255) PRIMARY KEY,
  checksum varchar(64) NOT NULL,
  applied_at timestamptz NOT NULL
)`).Error
}

func publicSchemaEmpty(db *gorm.DB) (bool, error) {
	var count int64
	err := db.Raw(`
SELECT COUNT(*)
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE'
  AND table_name <> 'schema_migrations'`).Scan(&count).Error
	return count == 0, err
}

func appliedMigrations(db *gorm.DB) (map[string]string, error) {
	rows := []struct {
		Version  string
		Checksum string
	}{}
	if err := db.Raw("SELECT version, checksum FROM public.schema_migrations").Scan(&rows).Error; err != nil {
		return nil, err
	}
	applied := map[string]string{}
	for _, row := range rows {
		applied[row.Version] = row.Checksum
	}
	return applied, nil
}

func markMigrationApplied(db *gorm.DB, version, checksum string) error {
	return db.Exec(`
INSERT INTO public.schema_migrations (version, checksum, applied_at)
VALUES (?, ?, ?)
ON CONFLICT (version) DO NOTHING`, version, checksum, time.Now().UTC()).Error
}

func execSQLFile(db *gorm.DB, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = db.Exec(string(raw)).Error
	_ = db.Exec("SET search_path TO public").Error
	return err
}

func loadMigrationFiles(dir string) ([]migrationFile, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	files := make([]migrationFile, 0, len(matches))
	for _, path := range matches {
		version := migrationVersionFromFile(path)
		if version == "" {
			continue
		}
		files = append(files, migrationFile{Version: version, Path: path, Checksum: checksumFile(path)})
	}
	return files, nil
}

func validateMigrationChecksums(files []migrationFile, applied map[string]string) error {
	for _, file := range files {
		checksum, ok := applied[file.Version]
		if !ok || checksum == "" {
			continue
		}
		if checksum != file.Checksum {
			return fmt.Errorf("migration checksum mismatch for %s: applied=%s local=%s", file.Version, checksum, file.Checksum)
		}
	}
	return nil
}

func strictMigrationChecksum() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("MIGRATION_STRICT_CHECKSUM")))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func migrationVersionFromFile(path string) string {
	name := filepath.Base(path)
	if !strings.HasSuffix(name, ".up.sql") {
		return ""
	}
	return strings.TrimSuffix(name, ".up.sql")
}

func checksumFile(path string) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
