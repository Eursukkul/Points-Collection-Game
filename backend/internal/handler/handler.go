// Package handler is the HTTP layer: bind/validate requests, call the
// service, shape responses. No business logic lives here.
package handler

import (
	"encoding/json"
	"errors"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/apierr"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/middleware"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
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
		return c.JSON(service.EmptySummary())
	}
	summary, err := h.svc.Summary(id)
	if err != nil {
		return apierr.Internal(c)
	}
	return c.JSON(summary)
}

func (h *Handler) play(c *fiber.Ctx) error {
	result, err := h.svc.Play(middleware.PlayerID(c))
	if err != nil {
		return apierr.Internal(c)
	}
	return c.JSON(result)
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

	// The service (checkpoint.Find) is the sole authority on valid checkpoints;
	// an out-of-range value (including 0) comes back as ErrCheckpointUnknown.
	claim, err := h.svc.Claim(middleware.PlayerID(c), req.Checkpoint)
	switch {
	case errors.Is(err, service.ErrCheckpointUnknown):
		return apierr.Respond(c, fiber.StatusBadRequest, "CHECKPOINT_UNKNOWN", "checkpoint must be one of 5000, 7500, 10000")
	case errors.Is(err, service.ErrCheckpointNotReached):
		return apierr.Respond(c, fiber.StatusConflict, "CHECKPOINT_NOT_REACHED", "not enough points for this checkpoint")
	case errors.Is(err, service.ErrAlreadyClaimed):
		return apierr.Respond(c, fiber.StatusConflict, "ALREADY_CLAIMED", "this checkpoint's reward was already claimed")
	case err != nil:
		return apierr.Internal(c)
	default:
		return c.JSON(claim)
	}
}

func (h *Handler) playHistory(c *fiber.Ctx) error {
	id := middleware.PlayerID(c)
	if id == uuid.Nil {
		return c.JSON(fiber.Map{"items": []any{}})
	}
	plays, err := h.svc.PlayHistory(id)
	if err != nil {
		return apierr.Internal(c)
	}
	return c.JSON(fiber.Map{"items": plays})
}

func (h *Handler) claimHistory(c *fiber.Ctx) error {
	id := middleware.PlayerID(c)
	if id == uuid.Nil {
		return c.JSON(fiber.Map{"items": []any{}})
	}
	claims, err := h.svc.ClaimHistory(id)
	if err != nil {
		return apierr.Internal(c)
	}
	return c.JSON(fiber.Map{"items": claims})
}

func (h *Handler) reset(c *fiber.Ctx) error {
	if err := h.svc.Reset(middleware.PlayerID(c)); err != nil {
		return apierr.Internal(c)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
