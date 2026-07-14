package database

import (
	"fmt"
	"time"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/checkpoint"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect opens a Postgres connection, verifies it with a ping, and sizes the
// pool for a small managed Postgres (e.g. Railway) so a burst of concurrent
// requests can't exhaust the server's connection ceiling.
func Connect(databaseURL string) (*gorm.DB, error) {
	// TranslateError maps driver errors to gorm sentinels (e.g. unique
	// violations → gorm.ErrDuplicatedKey) so services don't import pgconn.
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{TranslateError: true})
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
	sqlDB.SetMaxOpenConns(15)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}

// Migrate creates/updates the schema and applies the points ceiling CHECK
// constraint derived from checkpoint.MaxPoints. AutoMigrate can't alter an
// existing CHECK, so it's dropped and re-added explicitly — keeping the ceiling
// single-sourced and correct even when MaxPoints changes.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.Player{}, &model.Play{}, &model.Claim{}); err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(
		`ALTER TABLE players
		   DROP CONSTRAINT IF EXISTS chk_players_points,
		   ADD CONSTRAINT chk_players_points CHECK (points >= 0 AND points <= %d)`,
		checkpoint.MaxPoints)).Error
}
