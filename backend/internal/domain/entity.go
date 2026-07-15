// Package domain holds the enterprise entities and business rules. It has no
// dependency on any framework, database, or transport — everything else points
// inward to here.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Player is an anonymous player identified by an httpOnly cookie.
type Player struct {
	ID        uuid.UUID
	Points    int
	CreatedAt time.Time
}

// Play records a single game round and the score it produced (before cap clamping).
type Play struct {
	ID        uuid.UUID
	PlayerID  uuid.UUID
	Score     int
	CreatedAt time.Time
}

// Claim records a reward claimed at a checkpoint.
type Claim struct {
	ID         uuid.UUID
	PlayerID   uuid.UUID
	Checkpoint int
	RewardName string
	CreatedAt  time.Time
}
