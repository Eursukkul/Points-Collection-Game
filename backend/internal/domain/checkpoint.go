package domain

// MaxPoints is the accumulated points ceiling.
const MaxPoints = 10000

// Scores are the possible outcomes of a single play.
var Scores = []int{300, 500, 1000, 3000}

// Checkpoint is a reward milestone.
type Checkpoint struct {
	Threshold  int
	RewardName string
}

// Checkpoints are the reward milestones in ascending order.
var Checkpoints = []Checkpoint{
	{Threshold: 5000, RewardName: "รางวัล A"},
	{Threshold: 7500, RewardName: "รางวัล B"},
	{Threshold: 10000, RewardName: "รางวัล C"},
}

// FindCheckpoint returns the checkpoint with the given threshold.
func FindCheckpoint(threshold int) (Checkpoint, bool) {
	for _, cp := range Checkpoints {
		if cp.Threshold == threshold {
			return cp, true
		}
	}
	return Checkpoint{}, false
}

// Clamp caps a points total at the ceiling.
func Clamp(points int) int {
	if points > MaxPoints {
		return MaxPoints
	}
	return points
}
