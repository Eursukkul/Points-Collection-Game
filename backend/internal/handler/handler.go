// Package handler is the HTTP adapter: it binds/validates requests, calls the
// use case through an interface, and maps results to transport DTOs.
package handler

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/apierr"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/domain"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/middleware"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GameUseCase is the port the HTTP adapter needs. The concrete *usecase.UseCase
// satisfies it (dependency inversion — the adapter declares its dependency).
type GameUseCase interface {
	Summary(ctx context.Context, playerID uuid.UUID) (usecase.Summary, error)
	EmptySummary() usecase.Summary
	Play(ctx context.Context, playerID uuid.UUID) (usecase.PlayResult, error)
	Claim(ctx context.Context, playerID uuid.UUID, checkpoint int) (domain.Claim, error)
	PlayHistory(ctx context.Context, playerID uuid.UUID) ([]domain.Play, error)
	ClaimHistory(ctx context.Context, playerID uuid.UUID) ([]domain.Claim, error)
	Reset(ctx context.Context, playerID uuid.UUID) error
}

type Handler struct {
	uc GameUseCase
}

func New(uc GameUseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Register(api fiber.Router) {
	api.Get("/me", h.getMe)
	api.Post("/game/play", h.play)
	api.Post("/claims", h.claim)
	api.Get("/history/plays", h.playHistory)
	api.Get("/history/claims", h.claimHistory)
	api.Post("/reset", h.reset)
}

func (h *Handler) getMe(c *fiber.Ctx) error {
	id := middleware.PlayerID(c)
	if id == uuid.Nil {
		return c.JSON(toSummaryDTO(h.uc.EmptySummary()))
	}
	summary, err := h.uc.Summary(c.Context(), id)
	if err != nil {
		return apierr.Internal(c)
	}
	return c.JSON(toSummaryDTO(summary))
}

func (h *Handler) play(c *fiber.Ctx) error {
	result, err := h.uc.Play(c.Context(), middleware.PlayerID(c))
	if err != nil {
		return apierr.Internal(c)
	}
	return c.JSON(toPlayResultDTO(result))
}

type claimRequest struct {
	Checkpoint int `json:"checkpoint"`
}

func (h *Handler) claim(c *fiber.Ctx) error {
	// Parse the body directly rather than fiber's BodyParser, which rejects a
	// well-formed JSON body when Content-Type isn't set (breaks curl/Apidog).
	var req claimRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return apierr.Respond(c, fiber.StatusBadRequest, "INVALID_INPUT", `body must be {"checkpoint": <number>}`)
	}

	// The use case is the sole authority on valid checkpoints; an out-of-range
	// value (including 0) comes back as ErrCheckpointUnknown.
	claim, err := h.uc.Claim(c.Context(), middleware.PlayerID(c), req.Checkpoint)
	switch {
	case errors.Is(err, domain.ErrCheckpointUnknown):
		return apierr.Respond(c, fiber.StatusBadRequest, "CHECKPOINT_UNKNOWN", "checkpoint must be one of 5000, 7500, 10000")
	case errors.Is(err, domain.ErrCheckpointNotReached):
		return apierr.Respond(c, fiber.StatusConflict, "CHECKPOINT_NOT_REACHED", "not enough points for this checkpoint")
	case errors.Is(err, domain.ErrAlreadyClaimed):
		return apierr.Respond(c, fiber.StatusConflict, "ALREADY_CLAIMED", "this checkpoint's reward was already claimed")
	case err != nil:
		return apierr.Internal(c)
	default:
		return c.JSON(toClaimDTO(claim))
	}
}

func (h *Handler) playHistory(c *fiber.Ctx) error {
	id := middleware.PlayerID(c)
	if id == uuid.Nil {
		return c.JSON(fiber.Map{"items": []playDTO{}})
	}
	plays, err := h.uc.PlayHistory(c.Context(), id)
	if err != nil {
		return apierr.Internal(c)
	}
	return c.JSON(fiber.Map{"items": toPlayDTOs(plays)})
}

func (h *Handler) claimHistory(c *fiber.Ctx) error {
	id := middleware.PlayerID(c)
	if id == uuid.Nil {
		return c.JSON(fiber.Map{"items": []claimDTO{}})
	}
	claims, err := h.uc.ClaimHistory(c.Context(), id)
	if err != nil {
		return apierr.Internal(c)
	}
	return c.JSON(fiber.Map{"items": toClaimDTOs(claims)})
}

func (h *Handler) reset(c *fiber.Ctx) error {
	if err := h.uc.Reset(c.Context(), middleware.PlayerID(c)); err != nil {
		return apierr.Internal(c)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
