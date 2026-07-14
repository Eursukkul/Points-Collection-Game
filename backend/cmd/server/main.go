package main

import (
	"log"
	"net/http"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/config"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/database"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/handler"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/middleware"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	r := gin.Default()
	// Behind Railway's proxy we don't rely on client IPs; disable to avoid trusting all.
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Fatalf("set trusted proxies: %v", err)
	}

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	api.Use(middleware.EnsurePlayer(db, cfg.CookieSecure))
	handler.New(service.New(db)).Register(api)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
