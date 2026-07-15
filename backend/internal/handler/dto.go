package handler

import (
	"time"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/domain"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/usecase"
)

// Transport DTOs. Keeping the json contract here means the domain entities and
// use-case results stay free of transport concerns.

type checkpointDTO struct {
	Checkpoint int    `json:"checkpoint"`
	RewardName string `json:"rewardName"`
	Reached    bool   `json:"reached"`
	Claimed    bool   `json:"claimed"`
}

type summaryDTO struct {
	Points      int             `json:"points"`
	MaxPoints   int             `json:"maxPoints"`
	Checkpoints []checkpointDTO `json:"checkpoints"`
}

type playResultDTO struct {
	Score       int `json:"score"`
	PointsAdded int `json:"pointsAdded"`
	TotalPoints int `json:"totalPoints"`
}

type playDTO struct {
	ID        string    `json:"id"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"createdAt"`
}

type claimDTO struct {
	ID         string    `json:"id"`
	Checkpoint int       `json:"checkpoint"`
	RewardName string    `json:"rewardName"`
	CreatedAt  time.Time `json:"createdAt"`
}

func toSummaryDTO(s usecase.Summary) summaryDTO {
	cps := make([]checkpointDTO, 0, len(s.Checkpoints))
	for _, c := range s.Checkpoints {
		cps = append(cps, checkpointDTO(c))
	}
	return summaryDTO{Points: s.Points, MaxPoints: s.MaxPoints, Checkpoints: cps}
}

func toPlayResultDTO(r usecase.PlayResult) playResultDTO {
	return playResultDTO(r)
}

func toPlayDTOs(plays []domain.Play) []playDTO {
	out := make([]playDTO, 0, len(plays))
	for _, p := range plays {
		out = append(out, playDTO{ID: p.ID.String(), Score: p.Score, CreatedAt: p.CreatedAt})
	}
	return out
}

func toClaimDTO(c domain.Claim) claimDTO {
	return claimDTO{
		ID:         c.ID.String(),
		Checkpoint: c.Checkpoint,
		RewardName: c.RewardName,
		CreatedAt:  c.CreatedAt,
	}
}

func toClaimDTOs(claims []domain.Claim) []claimDTO {
	out := make([]claimDTO, 0, len(claims))
	for _, c := range claims {
		out = append(out, toClaimDTO(c))
	}
	return out
}
