package service

import (
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/checkpoint"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PlayResult struct {
	Score       int `json:"score"`
	PointsAdded int `json:"pointsAdded"`
	TotalPoints int `json:"totalPoints"`
}

// Play draws a random score server-side and credits it to the player,
// clamping accumulated points at checkpoint.MaxPoints. The player row is
// locked (SELECT ... FOR UPDATE) so concurrent plays serialize instead of
// overwriting each other's totals.
func (s *Service) Play(playerID uuid.UUID) (PlayResult, error) {
	score, err := s.randScore()
	if err != nil {
		return PlayResult{}, err
	}

	var result PlayResult
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var player model.Player
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&player, "id = ?", playerID).Error; err != nil {
			return err
		}

		newTotal := player.Points + score
		if newTotal > checkpoint.MaxPoints {
			newTotal = checkpoint.MaxPoints
		}

		if err := tx.Create(&model.Play{PlayerID: playerID, Score: score}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Player{}).Where("id = ?", playerID).
			Update("points", newTotal).Error; err != nil {
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
