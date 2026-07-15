// Package usecase holds the application business rules. It orchestrates the
// domain entities and depends only on domain ports (interfaces) — never on a
// concrete database or web framework.
package usecase

import "github.com/Eursukkul/Points-Collection-Game/backend/internal/domain"

type UseCase struct {
	repos domain.Repositories // non-transactional, for reads
	tx    domain.TxManager    // for atomic writes
	rand  domain.Randomizer
}

func New(repos domain.Repositories, tx domain.TxManager, rand domain.Randomizer) *UseCase {
	return &UseCase{repos: repos, tx: tx, rand: rand}
}
