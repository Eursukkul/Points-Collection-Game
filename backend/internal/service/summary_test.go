package service

import (
	"testing"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/testutil"
)

func TestSummary_CheckpointStates(t *testing.T) {
	db := testutil.DB(t)
	svc := New(db)
	player := testutil.NewPlayer(t, db, 7500)

	if err := db.Create(&model.Claim{
		PlayerID:   player.ID,
		Checkpoint: 5000,
		RewardName: "รางวัล A",
	}).Error; err != nil {
		t.Fatalf("seed claim: %v", err)
	}

	sum, err := svc.Summary(player.ID)
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}

	if sum.Points != 7500 || sum.MaxPoints != 10000 {
		t.Errorf("points = %d/%d, want 7500/10000", sum.Points, sum.MaxPoints)
	}
	if len(sum.Checkpoints) != 3 {
		t.Fatalf("checkpoints = %d, want 3", len(sum.Checkpoints))
	}

	want := []struct {
		checkpoint int
		reached    bool
		claimed    bool
	}{
		{5000, true, true},
		{7500, true, false},
		{10000, false, false},
	}
	for i, w := range want {
		got := sum.Checkpoints[i]
		if got.Checkpoint != w.checkpoint || got.Reached != w.reached || got.Claimed != w.claimed {
			t.Errorf("checkpoint[%d] = %+v, want %+v", i, got, w)
		}
	}
}
