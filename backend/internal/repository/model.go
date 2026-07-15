package repository

import (
	"time"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/domain"
	"github.com/google/uuid"
)

// The gorm persistence models. They carry the ORM tags so the domain entities
// stay framework-agnostic; mapping happens in this package.

type playerModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Points    int       `gorm:"not null;default:0"`
	CreatedAt time.Time
}

func (playerModel) TableName() string { return "players" }

type playModel struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PlayerID  uuid.UUID   `gorm:"type:uuid;not null;index:idx_plays_player_created,priority:1"`
	Player    playerModel `gorm:"constraint:OnDelete:CASCADE"`
	Score     int         `gorm:"not null"`
	CreatedAt time.Time   `gorm:"index:idx_plays_player_created,priority:2,sort:desc"`
}

func (playModel) TableName() string { return "plays" }

type claimModel struct {
	ID         uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PlayerID   uuid.UUID   `gorm:"type:uuid;not null;uniqueIndex:idx_claims_player_checkpoint"`
	Player     playerModel `gorm:"constraint:OnDelete:CASCADE"`
	Checkpoint int         `gorm:"not null;uniqueIndex:idx_claims_player_checkpoint"`
	RewardName string      `gorm:"not null"`
	CreatedAt  time.Time
}

func (claimModel) TableName() string { return "claims" }

func (m playerModel) toDomain() domain.Player {
	return domain.Player{ID: m.ID, Points: m.Points, CreatedAt: m.CreatedAt}
}

func (m playModel) toDomain() domain.Play {
	return domain.Play{ID: m.ID, PlayerID: m.PlayerID, Score: m.Score, CreatedAt: m.CreatedAt}
}

func (m claimModel) toDomain() domain.Claim {
	return domain.Claim{
		ID:         m.ID,
		PlayerID:   m.PlayerID,
		Checkpoint: m.Checkpoint,
		RewardName: m.RewardName,
		CreatedAt:  m.CreatedAt,
	}
}
