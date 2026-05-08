package main

import (
	"log"
	"os"

	"saas_baseon_go/internal/bootstrap"
)

func main() {
	cfg := bootstrap.LoadConfig()
	if err := cfg.ValidateForRuntime(); err != nil {
		log.Fatalf("configuration rejected: %v", err)
	}
	root, err := os.Getwd()
	if err != nil {
		log.Fatalf("resolve repo root: %v", err)
	}
	if err := bootstrap.RunMigrations(cfg.DatabaseDSN, root); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	status, err := bootstrap.MigrationStatusFor(cfg.DatabaseDSN, root)
	if err != nil {
		log.Fatalf("migration status failed: %v", err)
	}
	log.Printf("database migrations applied total=%d applied=%d pending=%d mismatched=%d", status.TotalVersioned, status.Applied, status.Pending, status.Mismatched)
}
