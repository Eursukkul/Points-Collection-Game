package service

import (
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/google/uuid"
)

// historyLimit bounds history responses; the UI shows recent items only.
const historyLimit = 50

func (s *Service) PlayHistory(playerID uuid.UUID) ([]model.Play, error) {
	plays := make([]model.Play, 0)
	err := s.db.Where("player_id = ?", playerID).
		Order("created_at DESC").Limit(historyLimit).
		Find(&plays).Error
	return plays, err
}

func (s *Service) ClaimHistory(playerID uuid.UUID) ([]model.Claim, error) {
	claims := make([]model.Claim, 0)
	err := s.db.Where("player_id = ?", playerID).
		Order("created_at DESC").Limit(historyLimit).
		Find(&claims).Error
	return claims, err
}
