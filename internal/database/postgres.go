package database

import (
	"fmt"
	"github.com/Nishithcs/bank-info/internal/config"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"log"
)

func ConnectPostgres(cfg config.PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	//  Optional: Enable logging
	//db.LogMode(true)

	//  Optional: Auto-migrate the Account model (if needed)
	err = db.AutoMigrate(&account.Account{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate account schema: %v", err)
		return nil, err
	}

	return db, nil
}