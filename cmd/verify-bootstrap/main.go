package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"saas_baseon_go/internal/bootstrap"
)

func main() {
	cfg := bootstrap.LoadConfig()
	if err := cfg.ValidateForRuntime(); err != nil {
		log.Fatalf("configuration rejected: %v", err)
	}
	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	result, err := bootstrap.VerifyBootstrapData(db)
	if err != nil {
		log.Fatalf("verify bootstrap data: %v", err)
	}
	log.Println("\n" + result.Summary())
	if !result.Passed() {
		log.Fatal("bootstrap verification failed")
	}
}
