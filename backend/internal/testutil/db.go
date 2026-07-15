// Package testutil provides shared test helpers backed by the local
// docker-compose Postgres — tests run against the real database engine.
package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/domain"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/repository"
	"gorm.io/gorm"
)

// DB connects to the test Postgres and ensures the schema is migrated.
func DB(t *testing.T) *gorm.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://points:points_local@localhost:5432/points_game?sslmode=disable"
	}
	db, err := repository.Connect(url)
	if err != nil {
		t.Skipf("postgres not available (run `docker compose up -d`): %v", err)
	}
	if err := repository.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// NewPlayer inserts a player with the given points; the row (and its plays and
// claims, via ON DELETE CASCADE) is removed on test cleanup.
func NewPlayer(t *testing.T, db *gorm.DB, points int) domain.Player {
	t.Helper()
	repos := repository.NewRepositories(db)
	ctx := context.Background()
	p, err := repos.Player().Create(ctx)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := repos.Player().UpdatePoints(ctx, p.ID, points); err != nil {
		t.Fatalf("set points: %v", err)
	}
	p.Points = points
	t.Cleanup(func() {
		db.Exec("DELETE FROM players WHERE id = ?", p.ID)
	})
	return p
}
