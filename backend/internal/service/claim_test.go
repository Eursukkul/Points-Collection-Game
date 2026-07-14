package service

import (
	"errors"
	"testing"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/testutil"
)

func TestClaim_AwardsReward(t *testing.T) {
	db := testutil.DB(t)
	svc := New(db)
	player := testutil.NewPlayer(t, db, 5000)

	claim, err := svc.Claim(player.ID, 5000)
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if claim.RewardName != "รางวัล A" {
		t.Errorf("RewardName = %q, want รางวัล A", claim.RewardName)
	}

	var count int64
	db.Model(&model.Claim{}).Where("player_id = ?", player.ID).Count(&count)
	if count != 1 {
		t.Errorf("claim rows = %d, want 1", count)
	}
}

func TestClaim_BelowThresholdRejected(t *testing.T) {
	db := testutil.DB(t)
	svc := New(db)
	player := testutil.NewPlayer(t, db, 4999)

	_, err := svc.Claim(player.ID, 5000)
	if !errors.Is(err, ErrCheckpointNotReached) {
		t.Fatalf("err = %v, want ErrCheckpointNotReached", err)
	}

	var count int64
	db.Model(&model.Claim{}).Where("player_id = ?", player.ID).Count(&count)
	if count != 0 {
		t.Errorf("claim rows = %d, want 0", count)
	}
}

func TestClaim_DuplicateRejected(t *testing.T) {
	db := testutil.DB(t)
	svc := New(db)
	player := testutil.NewPlayer(t, db, 10000)

	if _, err := svc.Claim(player.ID, 7500); err != nil {
		t.Fatalf("first Claim: %v", err)
	}
	_, err := svc.Claim(player.ID, 7500)
	if !errors.Is(err, ErrAlreadyClaimed) {
		t.Fatalf("err = %v, want ErrAlreadyClaimed", err)
	}

	var count int64
	db.Model(&model.Claim{}).Where("player_id = ?", player.ID).Count(&count)
	if count != 1 {
		t.Errorf("claim rows = %d, want exactly 1", count)
	}
}

func TestClaim_UnknownCheckpointRejected(t *testing.T) {
	db := testutil.DB(t)
	svc := New(db)
	player := testutil.NewPlayer(t, db, 10000)

	if _, err := svc.Claim(player.ID, 6000); !errors.Is(err, ErrCheckpointUnknown) {
		t.Fatalf("err = %v, want ErrCheckpointUnknown", err)
	}
}
