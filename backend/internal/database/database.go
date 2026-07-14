package database

import (
	"fmt"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect opens a Postgres connection and verifies it with a ping.
func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

// Migrate creates/updates the schema for all domain models.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&model.Player{}, &model.Play{}, &model.Claim{})
}
