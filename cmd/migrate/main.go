package main

import (
	"log"
	"os"

	"saas_baseon_go/internal/bootstrap"
)

func main() {
	cfg := bootstrap.LoadConfig()
	root, err := os.Getwd()
	if err != nil {
		log.Fatalf("resolve repo root: %v", err)
	}
	if err := bootstrap.RunMigrations(cfg.DatabaseDSN, root); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("database migrations applied")
}
