// Package handler is the HTTP layer: bind/validate requests, call the
// service, shape responses. No business logic lives here.
package handler

import (
	"errors"
	"net/http"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/middleware"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(api *gin.RouterGroup) {
	api.GET("/me", h.getMe)
	api.POST("/game/play", h.play)
	api.POST("/claims", h.claim)
	api.GET("/history/plays", h.playHistory)
	api.GET("/history/claims", h.claimHistory)
	api.POST("/reset", h.reset)
}

func (h *Handler) getMe(c *gin.Context) {
	summary, err := h.svc.Summary(middleware.PlayerID(c))
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handler) play(c *gin.Context) {
	result, err := h.svc.Play(middleware.PlayerID(c))
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, result)
}

type claimRequest struct {
	Checkpoint int `json:"checkpoint" binding:"required"`
}

func (h *Handler) claim(c *gin.Context) {
	var req claimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_INPUT", "body must be {\"checkpoint\": <number>}")
		return
	}

	claim, err := h.svc.Claim(middleware.PlayerID(c), req.Checkpoint)
	switch {
	case errors.Is(err, service.ErrCheckpointUnknown):
		respondError(c, http.StatusBadRequest, "CHECKPOINT_UNKNOWN", "checkpoint must be one of 5000, 7500, 10000")
	case errors.Is(err, service.ErrCheckpointNotReached):
		respondError(c, http.StatusConflict, "CHECKPOINT_NOT_REACHED", "not enough points for this checkpoint")
	case errors.Is(err, service.ErrAlreadyClaimed):
		respondError(c, http.StatusConflict, "ALREADY_CLAIMED", "this checkpoint's reward was already claimed")
	case err != nil:
		internalError(c)
	default:
		c.JSON(http.StatusOK, claim)
	}
}

func (h *Handler) playHistory(c *gin.Context) {
	plays, err := h.svc.PlayHistory(middleware.PlayerID(c))
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": plays})
}

func (h *Handler) claimHistory(c *gin.Context) {
	claims, err := h.svc.ClaimHistory(middleware.PlayerID(c))
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": claims})
}

func (h *Handler) reset(c *gin.Context) {
	if err := h.svc.Reset(middleware.PlayerID(c)); err != nil {
		internalError(c)
		return
	}
	c.Status(http.StatusNoContent)
}

func respondError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{"code": code, "message": message},
	})
}

func internalError(c *gin.Context) {
	respondError(c, http.StatusInternalServerError, "INTERNAL", "internal server error")
}
