// Package service holds the business logic between HTTP handlers and the database.
package service

import (
	"crypto/rand"
	"math/big"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/checkpoint"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
	// randScore is the score randomizer — a field so tests can make outcomes deterministic.
	randScore func() (int, error)
}

func New(db *gorm.DB) *Service {
	return &Service{db: db, randScore: cryptoRandScore}
}

// cryptoRandScore picks a score with crypto/rand — unpredictable even if a
// client tries to time or replay requests.
func cryptoRandScore() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(checkpoint.Scores))))
	if err != nil {
		return 0, err
	}
	return checkpoint.Scores[n.Int64()], nil
}
