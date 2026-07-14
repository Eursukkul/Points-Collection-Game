package service

import (
	"errors"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/checkpoint"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrCheckpointUnknown    = errors.New("unknown checkpoint")
	ErrCheckpointNotReached = errors.New("checkpoint not reached")
	ErrAlreadyClaimed       = errors.New("checkpoint already claimed")
)

// Claim awards the reward for a reached checkpoint. The player row is locked
// so the points check can't race a concurrent reset; the UNIQUE
// (player_id, checkpoint) index is the final guard against double claims.
func (s *Service) Claim(playerID uuid.UUID, threshold int) (model.Claim, error) {
	cp, ok := checkpoint.Find(threshold)
	if !ok {
		return model.Claim{}, ErrCheckpointUnknown
	}

	var claim model.Claim
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var player model.Player
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&player, "id = ?", playerID).Error; err != nil {
			return err
		}
		if player.Points < cp.Threshold {
			return ErrCheckpointNotReached
		}

		claim = model.Claim{
			PlayerID:   playerID,
			Checkpoint: cp.Threshold,
			RewardName: cp.RewardName,
		}
		if err := tx.Create(&claim).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrAlreadyClaimed
			}
			return err
		}
		return nil
	})
	return claim, err
}
