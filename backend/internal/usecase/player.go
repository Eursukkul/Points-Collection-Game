package usecase

import (
	"context"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/domain"
	"github.com/google/uuid"
)

type CheckpointStatus struct {
	Checkpoint int
	RewardName string
	Reached    bool
	Claimed    bool
}

type Summary struct {
	Points      int
	MaxPoints   int
	Checkpoints []CheckpointStatus
}

// ResolvePlayer maps a cookie to a player. When the cookie is missing or points
// to a non-existent player, it creates one only if allowCreate is set (safe
// read requests don't create rows). Returns the resolved id (uuid.Nil when no
// player and none created) and whether a new player was created.
func (u *UseCase) ResolvePlayer(ctx context.Context, cookieID uuid.UUID, hasCookie, allowCreate bool) (uuid.UUID, bool, error) {
	if hasCookie {
		exists, err := u.repos.Player().Exists(ctx, cookieID)
		if err != nil {
			return uuid.Nil, false, err
		}
		if exists {
			return cookieID, false, nil
		}
	}
	if !allowCreate {
		return uuid.Nil, false, nil
	}
	player, err := u.repos.Player().Create(ctx)
	if err != nil {
		return uuid.Nil, false, err
	}
	return player.ID, true, nil
}

// EmptySummary is the zero-state summary for a visitor with no player yet.
func (u *UseCase) EmptySummary() Summary {
	return buildSummary(0, nil)
}

// Summary returns the player's points and per-checkpoint reached/claimed state.
func (u *UseCase) Summary(ctx context.Context, playerID uuid.UUID) (Summary, error) {
	player, err := u.repos.Player().Get(ctx, playerID)
	if err != nil {
		return Summary{}, err
	}
	claims, err := u.repos.Claim().ListByPlayer(ctx, playerID, 0)
	if err != nil {
		return Summary{}, err
	}
	claimed := make(map[int]bool, len(claims))
	for _, c := range claims {
		claimed[c.Checkpoint] = true
	}
	return buildSummary(player.Points, claimed), nil
}

func buildSummary(points int, claimed map[int]bool) Summary {
	statuses := make([]CheckpointStatus, 0, len(domain.Checkpoints))
	for _, cp := range domain.Checkpoints {
		statuses = append(statuses, CheckpointStatus{
			Checkpoint: cp.Threshold,
			RewardName: cp.RewardName,
			Reached:    points >= cp.Threshold,
			Claimed:    claimed[cp.Threshold],
		})
	}
	return Summary{Points: points, MaxPoints: domain.MaxPoints, Checkpoints: statuses}
}

// Reset wipes the player's plays, claims, and points atomically, locking the
// player row first so a concurrent Play/Claim can't interleave.
func (u *UseCase) Reset(ctx context.Context, playerID uuid.UUID) error {
	return u.tx.WithinTx(ctx, func(r domain.Repositories) error {
		if _, err := r.Player().GetForUpdate(ctx, playerID); err != nil {
			return err
		}
		if err := r.Play().DeleteByPlayer(ctx, playerID); err != nil {
			return err
		}
		if err := r.Claim().DeleteByPlayer(ctx, playerID); err != nil {
			return err
		}
		return r.Player().UpdatePoints(ctx, playerID, 0)
	})
}
