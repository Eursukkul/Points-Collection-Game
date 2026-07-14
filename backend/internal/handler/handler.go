// Package handler is the HTTP layer: bind/validate requests, call the
// service, shape responses. No business logic lives here.
package handler

import (
	"errors"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/middleware"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/service"
	"github.com/gofiber/fiber/v2"
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
	summary, err := h.svc.Summary(middleware.PlayerID(c))
	if err != nil {
		return internalError(c)
	}
	return c.JSON(summary)
}

func (h *Handler) play(c *fiber.Ctx) error {
	result, err := h.svc.Play(middleware.PlayerID(c))
	if err != nil {
		return internalError(c)
	}
	return c.JSON(result)
}

type claimRequest struct {
	Checkpoint int `json:"checkpoint"`
}

func (h *Handler) claim(c *fiber.Ctx) error {
	var req claimRequest
	if err := c.BodyParser(&req); err != nil || req.Checkpoint == 0 {
		return respondError(c, fiber.StatusBadRequest, "INVALID_INPUT", `body must be {"checkpoint": <number>}`)
	}

	claim, err := h.svc.Claim(middleware.PlayerID(c), req.Checkpoint)
	switch {
	case errors.Is(err, service.ErrCheckpointUnknown):
		return respondError(c, fiber.StatusBadRequest, "CHECKPOINT_UNKNOWN", "checkpoint must be one of 5000, 7500, 10000")
	case errors.Is(err, service.ErrCheckpointNotReached):
		return respondError(c, fiber.StatusConflict, "CHECKPOINT_NOT_REACHED", "not enough points for this checkpoint")
	case errors.Is(err, service.ErrAlreadyClaimed):
		return respondError(c, fiber.StatusConflict, "ALREADY_CLAIMED", "this checkpoint's reward was already claimed")
	case err != nil:
		return internalError(c)
	default:
		return c.JSON(claim)
	}
}

func (h *Handler) playHistory(c *fiber.Ctx) error {
	plays, err := h.svc.PlayHistory(middleware.PlayerID(c))
	if err != nil {
		return internalError(c)
	}
	return c.JSON(fiber.Map{"items": plays})
}

func (h *Handler) claimHistory(c *fiber.Ctx) error {
	claims, err := h.svc.ClaimHistory(middleware.PlayerID(c))
	if err != nil {
		return internalError(c)
	}
	return c.JSON(fiber.Map{"items": claims})
}

func (h *Handler) reset(c *fiber.Ctx) error {
	if err := h.svc.Reset(middleware.PlayerID(c)); err != nil {
		return internalError(c)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func respondError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": fiber.Map{"code": code, "message": message},
	})
}

func internalError(c *fiber.Ctx) error {
	return respondError(c, fiber.StatusInternalServerError, "INTERNAL", "internal server error")
}
