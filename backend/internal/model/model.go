package model

import (
	"time"

	"github.com/google/uuid"
)

// Player is an anonymous player identified by an httpOnly cookie.
// The points ceiling CHECK constraint is applied in database.Migrate from
// checkpoint.MaxPoints (single source of truth) rather than a literal here.
type Player struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Points    int       `gorm:"not null;default:0" json:"points"`
	CreatedAt time.Time `json:"createdAt"`
}

// Play records a single game round and the score it produced (before cap clamping).
// The composite index (player_id, created_at DESC) serves PlayHistory's
// "WHERE player_id = ? ORDER BY created_at DESC LIMIT 50" as an index range scan.
type Play struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PlayerID  uuid.UUID `gorm:"type:uuid;not null;index:idx_plays_player_created,priority:1" json:"-"`
	Player    Player    `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	Score     int       `gorm:"not null" json:"score"`
	CreatedAt time.Time `gorm:"index:idx_plays_player_created,priority:2,sort:desc" json:"createdAt"`
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
