// Package server is the composition root — it wires repositories, the use case,
// and HTTP adapters into a Fiber app so main and tests boot the identical stack.
package server

import (
	"time"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/config"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/handler"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/middleware"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/repository"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/gorm"
)

func New(db *gorm.DB, cfg config.Config) *fiber.App {
	uc := usecase.New(
		repository.NewRepositories(db),
		repository.NewTxManager(db),
		repository.NewCryptoRandomizer(),
	)

	app := fiber.New(fiber.Config{AppName: "points-game"})

	// Fiber's recover is opt-in (unlike gin.Default): without it a handler panic
	// drops the connection with no response instead of a clean 500.
	app.Use(recover.New())

	// Cookie auth across origins: only the known frontend may send credentials.
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendOrigin,
		AllowMethods:     "GET,POST,OPTIONS",
		AllowHeaders:     "Content-Type",
		AllowCredentials: true,
		MaxAge:           int((12 * time.Hour).Seconds()),
	}))

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api/v1")
	// CSRFGuard runs before EnsurePlayer so a rejected cross-origin request never
	// triggers player creation.
	api.Use(middleware.CSRFGuard(cfg.FrontendOrigin))
	api.Use(middleware.EnsurePlayer(uc, cfg.CookieSecure))
	handler.New(uc).Register(api)

	return app
}
