package service

import (
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/checkpoint"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/google/uuid"
)

type CheckpointStatus struct {
	Checkpoint int    `json:"checkpoint"`
	RewardName string `json:"rewardName"`
	Reached    bool   `json:"reached"`
	Claimed    bool   `json:"claimed"`
}

type Summary struct {
	Points      int                `json:"points"`
	MaxPoints   int                `json:"maxPoints"`
	Checkpoints []CheckpointStatus `json:"checkpoints"`
}

// EmptySummary is the zero-state summary for a player that doesn't exist yet
// (a fresh visitor with no cookie) — no database row is created.
func EmptySummary() Summary {
	statuses := make([]CheckpointStatus, 0, len(checkpoint.All))
	for _, cp := range checkpoint.All {
		statuses = append(statuses, CheckpointStatus{
			Checkpoint: cp.Threshold,
			RewardName: cp.RewardName,
		})
	}
	return Summary{Points: 0, MaxPoints: checkpoint.MaxPoints, Checkpoints: statuses}
}

// Summary returns the player's points and per-checkpoint reached/claimed state.
func (s *Service) Summary(playerID uuid.UUID) (Summary, error) {
	var player model.Player
	if err := s.db.First(&player, "id = ?", playerID).Error; err != nil {
		return Summary{}, err
	}

	var claims []model.Claim
	if err := s.db.Where("player_id = ?", playerID).Find(&claims).Error; err != nil {
		return Summary{}, err
	}
	claimed := make(map[int]bool, len(claims))
	for _, cl := range claims {
		claimed[cl.Checkpoint] = true
	}

	statuses := make([]CheckpointStatus, 0, len(checkpoint.All))
	for _, cp := range checkpoint.All {
		statuses = append(statuses, CheckpointStatus{
			Checkpoint: cp.Threshold,
			RewardName: cp.RewardName,
			Reached:    player.Points >= cp.Threshold,
			Claimed:    claimed[cp.Threshold],
		})
	}

	return Summary{
		Points:      player.Points,
		MaxPoints:   checkpoint.MaxPoints,
		Checkpoints: statuses,
	}, nil
}
