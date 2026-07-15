package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/domain"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/repository"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/testutil"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/usecase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fixedRand struct{ v int }

func (f fixedRand) Score() (int, error) { return f.v, nil }

func newUC(db *gorm.DB, rand domain.Randomizer) *usecase.UseCase {
	return usecase.New(repository.NewRepositories(db), repository.NewTxManager(db), rand)
}

var ctx = context.Background()

func TestPlay_ClampsAtCeiling(t *testing.T) {
	db := testutil.DB(t)
	uc := newUC(db, fixedRand{3000})
	player := testutil.NewPlayer(t, db, 9800)

	res, err := uc.Play(ctx, player.ID)
	if err != nil {
		t.Fatalf("Play: %v", err)
	}
	if res.Score != 3000 || res.PointsAdded != 200 || res.TotalPoints != domain.MaxPoints {
		t.Errorf("got %+v, want score 3000 / added 200 / total %d", res, domain.MaxPoints)
	}

	repos := repository.NewRepositories(db)
	stored, _ := repos.Player().Get(ctx, player.ID)
	if stored.Points != domain.MaxPoints {
		t.Errorf("stored points = %d, want %d", stored.Points, domain.MaxPoints)
	}
	plays, _ := repos.Play().ListByPlayer(ctx, player.ID, 0)
	if len(plays) != 1 || plays[0].Score != 3000 {
		t.Errorf("play history should keep raw score 3000, got %+v", plays)
	}
}

func TestPlay_AtCeilingGainsZero(t *testing.T) {
	db := testutil.DB(t)
	uc := newUC(db, fixedRand{500})
	player := testutil.NewPlayer(t, db, domain.MaxPoints)

	res, err := uc.Play(ctx, player.ID)
	if err != nil {
		t.Fatalf("Play: %v", err)
	}
	if res.PointsAdded != 0 || res.TotalPoints != domain.MaxPoints {
		t.Errorf("got %+v, want added 0 / total %d", res, domain.MaxPoints)
	}
	plays, _ := repository.NewRepositories(db).Play().ListByPlayer(ctx, player.ID, 0)
	if len(plays) != 1 {
		t.Errorf("play count = %d, want 1", len(plays))
	}
}

func TestPlay_RandomScoreMembership(t *testing.T) {
	db := testutil.DB(t)
	uc := newUC(db, repository.NewCryptoRandomizer())
	player := testutil.NewPlayer(t, db, 0)

	valid := map[int]bool{}
	for _, s := range domain.Scores {
		valid[s] = true
	}
	seen := map[int]bool{}
	for i := 0; i < 40; i++ {
		res, err := uc.Play(ctx, player.ID)
		if err != nil {
			t.Fatalf("Play #%d: %v", i+1, err)
		}
		if !valid[res.Score] {
			t.Fatalf("score %d outside %v", res.Score, domain.Scores)
		}
		seen[res.Score] = true
	}
	if len(seen) < 2 {
		t.Errorf("expected variety, saw %v", seen)
	}
	stored, _ := repository.NewRepositories(db).Player().Get(ctx, player.ID)
	if stored.Points != domain.MaxPoints {
		t.Errorf("points = %d, want clamped %d", stored.Points, domain.MaxPoints)
	}
}

func TestPlay_ConcurrentPlaysSerialize(t *testing.T) {
	db := testutil.DB(t)
	uc := newUC(db, fixedRand{300})
	player := testutil.NewPlayer(t, db, 0)

	const rounds = 10
	errs := make(chan error, rounds)
	for i := 0; i < rounds; i++ {
		go func() {
			_, err := uc.Play(ctx, player.ID)
			errs <- err
		}()
	}
	for i := 0; i < rounds; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent Play: %v", err)
		}
	}

	repos := repository.NewRepositories(db)
	stored, _ := repos.Player().Get(ctx, player.ID)
	if stored.Points != rounds*300 {
		t.Errorf("points = %d, want %d (lost updates)", stored.Points, rounds*300)
	}
	plays, _ := repos.Play().ListByPlayer(ctx, player.ID, 0)
	if len(plays) != rounds {
		t.Errorf("play rows = %d, want %d", len(plays), rounds)
	}
}

func TestClaim_AwardsReward(t *testing.T) {
	db := testutil.DB(t)
	uc := newUC(db, fixedRand{300})
	player := testutil.NewPlayer(t, db, 5000)

	claim, err := uc.Claim(ctx, player.ID, 5000)
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if claim.RewardName != "รางวัล A" {
		t.Errorf("RewardName = %q, want รางวัล A", claim.RewardName)
	}
	claims, _ := repository.NewRepositories(db).Claim().ListByPlayer(ctx, player.ID, 0)
	if len(claims) != 1 {
		t.Errorf("claim rows = %d, want 1", len(claims))
	}
}

func TestClaim_BelowThresholdRejected(t *testing.T) {
	db := testutil.DB(t)
	uc := newUC(db, fixedRand{300})
	player := testutil.NewPlayer(t, db, 4999)

	if _, err := uc.Claim(ctx, player.ID, 5000); !errors.Is(err, domain.ErrCheckpointNotReached) {
		t.Fatalf("err = %v, want ErrCheckpointNotReached", err)
	}
	claims, _ := repository.NewRepositories(db).Claim().ListByPlayer(ctx, player.ID, 0)
	if len(claims) != 0 {
		t.Errorf("claim rows = %d, want 0", len(claims))
	}
}

func TestClaim_DuplicateRejected(t *testing.T) {
	db := testutil.DB(t)
	uc := newUC(db, fixedRand{300})
	player := testutil.NewPlayer(t, db, domain.MaxPoints)

	if _, err := uc.Claim(ctx, player.ID, 7500); err != nil {
		t.Fatalf("first Claim: %v", err)
	}
	if _, err := uc.Claim(ctx, player.ID, 7500); !errors.Is(err, domain.ErrAlreadyClaimed) {
		t.Fatalf("err = %v, want ErrAlreadyClaimed", err)
	}
	claims, _ := repository.NewRepositories(db).Claim().ListByPlayer(ctx, player.ID, 0)
	if len(claims) != 1 {
		t.Errorf("claim rows = %d, want exactly 1", len(claims))
	}
}

func TestClaim_UnknownCheckpointRejected(t *testing.T) {
	db := testutil.DB(t)
	uc := newUC(db, fixedRand{300})
	player := testutil.NewPlayer(t, db, domain.MaxPoints)

	if _, err := uc.Claim(ctx, player.ID, 6000); !errors.Is(err, domain.ErrCheckpointUnknown) {
		t.Fatalf("err = %v, want ErrCheckpointUnknown", err)
	}
}

func TestReset_WipesOnlyThisPlayer(t *testing.T) {
	db := testutil.DB(t)
	uc := newUC(db, fixedRand{3000})
	player := testutil.NewPlayer(t, db, 0)
	other := testutil.NewPlayer(t, db, 0)

	for _, id := range []uuid.UUID{player.ID, player.ID, other.ID} {
		if _, err := uc.Play(ctx, id); err != nil {
			t.Fatalf("seed play: %v", err)
		}
	}
	if _, err := uc.Claim(ctx, player.ID, 5000); err != nil {
		t.Fatalf("seed claim: %v", err)
	}

	if err := uc.Reset(ctx, player.ID); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	repos := repository.NewRepositories(db)
	stored, _ := repos.Player().Get(ctx, player.ID)
	if stored.Points != 0 {
		t.Errorf("points = %d, want 0", stored.Points)
	}
	plays, _ := repos.Play().ListByPlayer(ctx, player.ID, 0)
	claims, _ := repos.Claim().ListByPlayer(ctx, player.ID, 0)
	if len(plays) != 0 || len(claims) != 0 {
		t.Errorf("plays = %d, claims = %d, want 0/0", len(plays), len(claims))
	}

	otherStored, _ := repos.Player().Get(ctx, other.ID)
	if otherStored.Points != 3000 {
		t.Errorf("other points = %d, want 3000", otherStored.Points)
	}
	otherPlays, _ := repos.Play().ListByPlayer(ctx, other.ID, 0)
	if len(otherPlays) != 1 {
		t.Errorf("other plays = %d, want 1", len(otherPlays))
	}
}

func TestSummary_CheckpointStates(t *testing.T) {
	db := testutil.DB(t)
	uc := newUC(db, fixedRand{300})
	player := testutil.NewPlayer(t, db, 7500)

	if _, err := uc.Claim(ctx, player.ID, 5000); err != nil {
		t.Fatalf("seed claim: %v", err)
	}

	sum, err := uc.Summary(ctx, player.ID)
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if sum.Points != 7500 || sum.MaxPoints != domain.MaxPoints {
		t.Errorf("points = %d/%d, want 7500/%d", sum.Points, sum.MaxPoints, domain.MaxPoints)
	}
	want := []struct {
		cp      int
		reached bool
		claimed bool
	}{{5000, true, true}, {7500, true, false}, {10000, false, false}}
	for i, w := range want {
		got := sum.Checkpoints[i]
		if got.Checkpoint != w.cp || got.Reached != w.reached || got.Claimed != w.claimed {
			t.Errorf("checkpoint[%d] = %+v, want %+v", i, got, w)
		}
	}
}

func TestResolvePlayer(t *testing.T) {
	db := testutil.DB(t)
	uc := newUC(db, fixedRand{300})

	// Safe request, no cookie → no player, none created.
	id, created, err := uc.ResolvePlayer(ctx, uuid.Nil, false, false)
	if err != nil || created || id != uuid.Nil {
		t.Fatalf("safe no-cookie: id=%v created=%v err=%v, want Nil/false/nil", id, created, err)
	}

	// State-changing request, no cookie → creates a player.
	id, created, err = uc.ResolvePlayer(ctx, uuid.Nil, false, true)
	if err != nil || !created || id == uuid.Nil {
		t.Fatalf("create: id=%v created=%v err=%v", id, created, err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM players WHERE id = ?", id) })

	// Existing cookie → same player, not recreated.
	same, created, err := uc.ResolvePlayer(ctx, id, true, true)
	if err != nil || created || same != id {
		t.Fatalf("reuse: same=%v created=%v err=%v, want %v/false/nil", same, created, err, id)
	}

	// Unknown cookie + allowCreate → a fresh player, not the tampered id.
	fresh, created, err := uc.ResolvePlayer(ctx, uuid.New(), true, true)
	if err != nil || !created || fresh == id {
		t.Fatalf("tampered: fresh=%v created=%v err=%v", fresh, created, err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM players WHERE id = ?", fresh) })
}
