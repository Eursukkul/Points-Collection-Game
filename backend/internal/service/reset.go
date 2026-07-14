package service

import (
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Reset wipes the player's plays, claims, and points in one transaction so
// testers can start over. It locks the player row FIRST (SELECT ... FOR UPDATE)
// so a concurrent Play/Claim can't interleave and leave, e.g., a surviving
// claim row against a zeroed player. Other players' data is untouched.
func (s *Service) Reset(playerID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var player model.Player
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&player, "id = ?", playerID).Error; err != nil {
			return err
		}
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
