package service

import (
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Reset wipes the player's plays, claims, and points in one transaction so
// testers can start over. Other players' data is untouched.
func (s *Service) Reset(playerID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("player_id = ?", playerID).Delete(&model.Play{}).Error; err != nil {
			return err
		}
		if err := tx.Where("player_id = ?", playerID).Delete(&model.Claim{}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Player{}).Where("id = ?", playerID).
			Update("points", 0).Error
	})
}
