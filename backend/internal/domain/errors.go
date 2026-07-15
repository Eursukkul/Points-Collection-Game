package domain

import "errors"

var (
	// ErrCheckpointUnknown is returned for a threshold that isn't a checkpoint.
	ErrCheckpointUnknown = errors.New("unknown checkpoint")
	// ErrCheckpointNotReached is returned when the player's points are below the threshold.
	ErrCheckpointNotReached = errors.New("checkpoint not reached")
	// ErrAlreadyClaimed is returned when a checkpoint's reward was already claimed.
	ErrAlreadyClaimed = errors.New("checkpoint already claimed")
	// ErrPlayerNotFound is returned when a player row does not exist.
	ErrPlayerNotFound = errors.New("player not found")
)
