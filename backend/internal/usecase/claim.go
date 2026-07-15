package usecase

import (
	"context"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/domain"
	"github.com/google/uuid"
)

// historyLimit bounds history responses; the UI shows recent items only.
const historyLimit = 50

// Claim awards the reward for a reached checkpoint. The player row is locked so
// the points check can't race a concurrent reset; the repository's uniqueness
// guard is the final defense against double claims.
func (u *UseCase) Claim(ctx context.Context, playerID uuid.UUID, threshold int) (domain.Claim, error) {
	cp, ok := domain.FindCheckpoint(threshold)
	if !ok {
		return domain.Claim{}, domain.ErrCheckpointUnknown
	}

	var claim domain.Claim
	err := u.tx.WithinTx(ctx, func(r domain.Repositories) error {
		player, err := r.Player().GetForUpdate(ctx, playerID)
		if err != nil {
			return err
		}
		if player.Points < cp.Threshold {
			return domain.ErrCheckpointNotReached
		}
		created, err := r.Claim().Create(ctx, domain.Claim{
			PlayerID:   playerID,
			Checkpoint: cp.Threshold,
			RewardName: cp.RewardName,
		})
		if err != nil {
			return err
		}
		claim = created
		return nil
	})
	// claim is the zero value on any error path (it is only assigned on success),
	// so returning it directly is correct; the handler branches on err via errors.Is.
	return claim, err
}

// ClaimHistory returns the player's claimed rewards (newest first).
func (u *UseCase) ClaimHistory(ctx context.Context, playerID uuid.UUID) ([]domain.Claim, error) {
	return u.repos.Claim().ListByPlayer(ctx, playerID, historyLimit)
}
