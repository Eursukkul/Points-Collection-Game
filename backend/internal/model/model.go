package model

import (
	"time"

	"github.com/google/uuid"
)

// Player is an anonymous player identified by an httpOnly cookie.
type Player struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Points    int       `gorm:"not null;default:0;check:points >= 0 AND points <= 10000" json:"points"`
	CreatedAt time.Time `json:"createdAt"`
}

// Play records a single game round and the score it produced (before cap clamping).
type Play struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PlayerID  uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`
	Player    Player    `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	Score     int       `gorm:"not null" json:"score"`
	CreatedAt time.Time `json:"createdAt"`
}

// Claim records a reward claimed at a checkpoint.
// The unique index makes claiming idempotent per player+checkpoint.
type Claim struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PlayerID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_claims_player_checkpoint" json:"-"`
	Player     Player    `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	Checkpoint int       `gorm:"not null;uniqueIndex:idx_claims_player_checkpoint" json:"checkpoint"`
	RewardName string    `gorm:"not null" json:"rewardName"`
	CreatedAt  time.Time `json:"createdAt"`
}
