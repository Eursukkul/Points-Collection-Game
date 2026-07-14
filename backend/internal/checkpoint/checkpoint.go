// Package checkpoint holds the game's fixed business constants.
// Three fixed checkpoints don't justify a database table.
package checkpoint

// MaxPoints is the accumulated points ceiling.
const MaxPoints = 10000

// Scores are the possible outcomes of a single play.
var Scores = []int{300, 500, 1000, 3000}

type Checkpoint struct {
	Threshold  int
	RewardName string
}

// All checkpoints in ascending order.
var All = []Checkpoint{
	{Threshold: 5000, RewardName: "รางวัล A"},
	{Threshold: 7500, RewardName: "รางวัล B"},
	{Threshold: 10000, RewardName: "รางวัล C"},
}

// Find returns the checkpoint with the given threshold.
func Find(threshold int) (Checkpoint, bool) {
	for _, cp := range All {
		if cp.Threshold == threshold {
			return cp, true
		}
	}
	return Checkpoint{}, false
}
