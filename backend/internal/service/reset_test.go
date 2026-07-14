package service

import (
	"testing"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/testutil"
)

func TestReset_WipesOnlyThisPlayer(t *testing.T) {
	db := testutil.DB(t)
	svc := New(db)
	svc.randScore = fixedScore(3000)

	player := testutil.NewPlayer(t, db, 0)
	other := testutil.NewPlayer(t, db, 0)

	// player: 2 plays (6000 points) + 1 claim; other: 1 play (3000 points).
	for _, id := range []model.Player{player, player, other} {
		if _, err := svc.Play(id.ID); err != nil {
			t.Fatalf("seed play: %v", err)
		}
	}
	if _, err := svc.Claim(player.ID, 5000); err != nil {
		t.Fatalf("seed claim: %v", err)
	}

	if err := svc.Reset(player.ID); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	var stored model.Player
	if err := db.First(&stored, "id = ?", player.ID).Error; err != nil {
		t.Fatalf("reload player: %v", err)
	}
	if stored.Points != 0 {
		t.Errorf("points = %d, want 0", stored.Points)
	}
	var plays, claims int64
	db.Model(&model.Play{}).Where("player_id = ?", player.ID).Count(&plays)
	db.Model(&model.Claim{}).Where("player_id = ?", player.ID).Count(&claims)
	if plays != 0 || claims != 0 {
		t.Errorf("plays = %d, claims = %d, want 0/0", plays, claims)
	}

	// The other player's data must be untouched.
	var otherStored model.Player
	if err := db.First(&otherStored, "id = ?", other.ID).Error; err != nil {
		t.Fatalf("reload other: %v", err)
	}
	if otherStored.Points != 3000 {
		t.Errorf("other points = %d, want 3000", otherStored.Points)
	}
	var otherPlays int64
	db.Model(&model.Play{}).Where("player_id = ?", other.ID).Count(&otherPlays)
	if otherPlays != 1 {
		t.Errorf("other plays = %d, want 1", otherPlays)
	}
}
