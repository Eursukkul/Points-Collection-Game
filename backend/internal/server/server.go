// Package server assembles the Fiber app — middleware, routes, CORS — so
// main and tests boot the identical stack.
package server

import (
	"time"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/config"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/handler"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/middleware"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"gorm.io/gorm"
)

func New(db *gorm.DB, cfg config.Config) *fiber.App {
	app := fiber.New(fiber.Config{AppName: "points-game"})

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
	api.Use(middleware.EnsurePlayer(db, cfg.CookieSecure))
	handler.New(service.New(db)).Register(api)

	return app
}
