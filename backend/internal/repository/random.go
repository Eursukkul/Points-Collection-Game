package repository

import (
	"crypto/rand"
	"math/big"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/domain"
)

// cryptoRandomizer picks a score with crypto/rand — unpredictable, so a client
// can't anticipate or replay outcomes.
type cryptoRandomizer struct{}

// NewCryptoRandomizer returns the production Randomizer.
func NewCryptoRandomizer() domain.Randomizer { return cryptoRandomizer{} }

func (cryptoRandomizer) Score() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(domain.Scores))))
	if err != nil {
		return 0, err
	}
	return domain.Scores[n.Int64()], nil
}
