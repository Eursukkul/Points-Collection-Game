package domain

import (
	"context"

	"github.com/google/uuid"
)

// PlayerRepository persists players. GetForUpdate takes a row lock so
// concurrent writes serialize instead of losing updates.
type PlayerRepository interface {
	Create(ctx context.Context) (Player, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	Get(ctx context.Context, id uuid.UUID) (Player, error)
	GetForUpdate(ctx context.Context, id uuid.UUID) (Player, error)
	UpdatePoints(ctx context.Context, id uuid.UUID, points int) error
}

// PlayRepository persists play records.
type PlayRepository interface {
	Create(ctx context.Context, playerID uuid.UUID, score int) (Play, error)
	ListByPlayer(ctx context.Context, playerID uuid.UUID, limit int) ([]Play, error)
	DeleteByPlayer(ctx context.Context, playerID uuid.UUID) error
}

// ClaimRepository persists reward claims. Create returns ErrAlreadyClaimed when
// the (player, checkpoint) uniqueness is violated.
type ClaimRepository interface {
	Create(ctx context.Context, claim Claim) (Claim, error)
	ListByPlayer(ctx context.Context, playerID uuid.UUID, limit int) ([]Claim, error)
	DeleteByPlayer(ctx context.Context, playerID uuid.UUID) error
}

// Repositories bundles the repositories, whether backed by the base connection
// (reads) or a transaction (writes).
type Repositories interface {
	Player() PlayerRepository
	Play() PlayRepository
	Claim() ClaimRepository
}

// TxManager runs fn inside a database transaction, exposing transactional
// repositories. A returned error rolls the transaction back.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(Repositories) error) error
}

// Randomizer picks a play score. Abstracted so the source of randomness is a
// plug-in (crypto/rand in production, fixed value in tests).
type Randomizer interface {
	Score() (int, error)
}
