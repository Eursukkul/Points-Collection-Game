// Package handler is the HTTP layer: bind/validate requests, call the
// service, shape responses. No business logic lives here.
package handler

import (
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

func internalError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{"code": "INTERNAL", "message": "internal server error"},
	})
}
