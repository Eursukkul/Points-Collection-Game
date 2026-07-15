// Package repository implements the domain ports with Gorm/Postgres. It is the
// only package that imports gorm — the layers above depend on domain interfaces.
package repository

import (
	"context"
	"errors"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// repositories bundles the port implementations over one *gorm.DB handle
// (either the base connection or an open transaction).
type repositories struct {
	db *gorm.DB
}

// NewRepositories returns non-transactional repositories for reads.
func NewRepositories(db *gorm.DB) domain.Repositories { return repositories{db: db} }

func (r repositories) Player() domain.PlayerRepository { return playerRepo{r.db} }
func (r repositories) Play() domain.PlayRepository     { return playRepo{r.db} }
func (r repositories) Claim() domain.ClaimRepository   { return claimRepo{r.db} }

// txManager runs a function inside a Gorm transaction.
type txManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) domain.TxManager { return txManager{db: db} }

func (t txManager) WithinTx(ctx context.Context, fn func(domain.Repositories) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(repositories{db: tx})
	})
}

// --- PlayerRepository ---

type playerRepo struct{ db *gorm.DB }

func (r playerRepo) Create(ctx context.Context) (domain.Player, error) {
	m := playerModel{}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return domain.Player{}, err
	}
	return m.toDomain(), nil
}

func (r playerRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&playerModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r playerRepo) Get(ctx context.Context, id uuid.UUID) (domain.Player, error) {
	return r.first(ctx, id, false)
}

func (r playerRepo) GetForUpdate(ctx context.Context, id uuid.UUID) (domain.Player, error) {
	return r.first(ctx, id, true)
}

func (r playerRepo) first(ctx context.Context, id uuid.UUID, lock bool) (domain.Player, error) {
	q := r.db.WithContext(ctx)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var m playerModel
	if err := q.First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Player{}, domain.ErrPlayerNotFound
		}
		return domain.Player{}, err
	}
	return m.toDomain(), nil
}

func (r playerRepo) UpdatePoints(ctx context.Context, id uuid.UUID, points int) error {
	return r.db.WithContext(ctx).Model(&playerModel{}).Where("id = ?", id).
		Update("points", points).Error
}

// --- PlayRepository ---

type playRepo struct{ db *gorm.DB }

func (r playRepo) Create(ctx context.Context, playerID uuid.UUID, score int) (domain.Play, error) {
	m := playModel{PlayerID: playerID, Score: score}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return domain.Play{}, err
	}
	return m.toDomain(), nil
}

func (r playRepo) ListByPlayer(ctx context.Context, playerID uuid.UUID, limit int) ([]domain.Play, error) {
	var models []playModel
	q := r.db.WithContext(ctx).Where("player_id = ?", playerID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Play, 0, len(models))
	for _, m := range models {
		out = append(out, m.toDomain())
	}
	return out, nil
}

func (r playRepo) DeleteByPlayer(ctx context.Context, playerID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("player_id = ?", playerID).Delete(&playModel{}).Error
}

// --- ClaimRepository ---

type claimRepo struct{ db *gorm.DB }

func (r claimRepo) Create(ctx context.Context, c domain.Claim) (domain.Claim, error) {
	m := claimModel{PlayerID: c.PlayerID, Checkpoint: c.Checkpoint, RewardName: c.RewardName}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.Claim{}, domain.ErrAlreadyClaimed
		}
		return domain.Claim{}, err
	}
	return m.toDomain(), nil
}

func (r claimRepo) ListByPlayer(ctx context.Context, playerID uuid.UUID, limit int) ([]domain.Claim, error) {
	var models []claimModel
	q := r.db.WithContext(ctx).Where("player_id = ?", playerID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Claim, 0, len(models))
	for _, m := range models {
		out = append(out, m.toDomain())
	}
	return out, nil
}

func (r claimRepo) DeleteByPlayer(ctx context.Context, playerID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("player_id = ?", playerID).Delete(&claimModel{}).Error
}
