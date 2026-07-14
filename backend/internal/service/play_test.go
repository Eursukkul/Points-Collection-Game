package service

import (
	"testing"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/checkpoint"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/testutil"
)

func fixedScore(n int) func() (int, error) {
	return func() (int, error) { return n, nil }
}

func TestPlay_ClampsAtCeiling(t *testing.T) {
	db := testutil.DB(t)
	svc := New(db)
	svc.randScore = fixedScore(3000)
	player := testutil.NewPlayer(t, db, 9800)

	res, err := svc.Play(player.ID)
	if err != nil {
		t.Fatalf("Play: %v", err)
	}

	if res.Score != 3000 {
		t.Errorf("Score = %d, want 3000", res.Score)
	}
	if res.PointsAdded != 200 {
		t.Errorf("PointsAdded = %d, want 200 (clamped)", res.PointsAdded)
	}
	if res.TotalPoints != checkpoint.MaxPoints {
		t.Errorf("TotalPoints = %d, want %d", res.TotalPoints, checkpoint.MaxPoints)
	}

	var stored model.Player
	if err := db.First(&stored, "id = ?", player.ID).Error; err != nil {
		t.Fatalf("reload player: %v", err)
	}
	if stored.Points != checkpoint.MaxPoints {
		t.Errorf("stored points = %d, want %d", stored.Points, checkpoint.MaxPoints)
	}

	var play model.Play
	if err := db.First(&play, "player_id = ?", player.ID).Error; err != nil {
		t.Fatalf("load play: %v", err)
	}
	if play.Score != 3000 {
		t.Errorf("play history keeps raw score: got %d, want 3000", play.Score)
	}
}

func TestPlay_AtCeilingGainsZero(t *testing.T) {
	db := testutil.DB(t)
	svc := New(db)
	svc.randScore = fixedScore(500)
	player := testutil.NewPlayer(t, db, checkpoint.MaxPoints)

	res, err := svc.Play(player.ID)
	if err != nil {
		t.Fatalf("Play: %v", err)
	}

	if res.PointsAdded != 0 {
		t.Errorf("PointsAdded = %d, want 0", res.PointsAdded)
	}
	if res.TotalPoints != checkpoint.MaxPoints {
		t.Errorf("TotalPoints = %d, want %d", res.TotalPoints, checkpoint.MaxPoints)
	}

	// Playing at the ceiling still records history.
	var plays int64
	db.Model(&model.Play{}).Where("player_id = ?", player.ID).Count(&plays)
	if plays != 1 {
		t.Errorf("play count = %d, want 1", plays)
	}
}

func TestPlay_RandomScoreMembership(t *testing.T) {
	db := testutil.DB(t)
	svc := New(db) // real crypto/rand
	player := testutil.NewPlayer(t, db, 0)

	valid := make(map[int]bool, len(checkpoint.Scores))
	for _, s := range checkpoint.Scores {
		valid[s] = true
	}

	seen := make(map[int]bool)
	const rounds = 40
	for i := 0; i < rounds; i++ {
		res, err := svc.Play(player.ID)
		if err != nil {
			t.Fatalf("Play #%d: %v", i+1, err)
		}
		if !valid[res.Score] {
			t.Fatalf("Play #%d returned score %d outside %v", i+1, res.Score, checkpoint.Scores)
		}
		seen[res.Score] = true
	}

	if len(seen) < 2 {
		t.Errorf("expected variety across %d rounds, saw only %v", rounds, seen)
	}

	// 40 rounds × min score 300 far exceeds the ceiling — total must be clamped.
	var stored model.Player
	if err := db.First(&stored, "id = ?", player.ID).Error; err != nil {
		t.Fatalf("reload player: %v", err)
	}
	if stored.Points != checkpoint.MaxPoints {
		t.Errorf("stored points = %d, want clamped %d", stored.Points, checkpoint.MaxPoints)
	}
}

func TestPlay_ConcurrentPlaysSerialize(t *testing.T) {
	db := testutil.DB(t)
	svc := New(db)
	svc.randScore = fixedScore(300)
	player := testutil.NewPlayer(t, db, 0)

	const rounds = 10
	errs := make(chan error, rounds)
	for i := 0; i < rounds; i++ {
		go func() {
			_, err := svc.Play(player.ID)
			errs <- err
		}()
	}
	for i := 0; i < rounds; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent Play: %v", err)
		}
	}

	// Without SELECT ... FOR UPDATE, racing read-modify-writes would lose updates.
	var stored model.Player
	if err := db.First(&stored, "id = ?", player.ID).Error; err != nil {
		t.Fatalf("reload player: %v", err)
	}
	if want := rounds * 300; stored.Points != want {
		t.Errorf("points = %d, want %d (lost updates)", stored.Points, want)
	}
	var plays int64
	db.Model(&model.Play{}).Where("player_id = ?", player.ID).Count(&plays)
	if plays != rounds {
		t.Errorf("play rows = %d, want %d", plays, rounds)
	}
}
