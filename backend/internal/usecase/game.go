package usecase

import (
	"context"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/domain"
	"github.com/google/uuid"
)

type PlayResult struct {
	Score       int
	PointsAdded int
	TotalPoints int
}

// Play draws a server-side random score and credits it to the player, clamping
// at the ceiling. The player row is locked for the read-modify-write so
// concurrent plays serialize.
func (u *UseCase) Play(ctx context.Context, playerID uuid.UUID) (PlayResult, error) {
	score, err := u.rand.Score()
	if err != nil {
		return PlayResult{}, err
	}

	var result PlayResult
	err = u.tx.WithinTx(ctx, func(r domain.Repositories) error {
		player, err := r.Player().GetForUpdate(ctx, playerID)
		if err != nil {
			return err
		}
		newTotal := domain.Clamp(player.Points + score)
		if _, err := r.Play().Create(ctx, playerID, score); err != nil {
			return err
		}
		if err := r.Player().UpdatePoints(ctx, playerID, newTotal); err != nil {
			return err
		}
		result = PlayResult{
			Score:       score,
			PointsAdded: newTotal - player.Points,
			TotalPoints: newTotal,
		}
		return nil
	})
	return result, err
}

// PlayHistory returns the player's recent plays (newest first).
func (u *UseCase) PlayHistory(ctx context.Context, playerID uuid.UUID) ([]domain.Play, error) {
	return u.repos.Play().ListByPlayer(ctx, playerID, historyLimit)
}
