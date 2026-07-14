package middleware

import (
	"errors"
	"time"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	cookieName   = "player_id"
	ctxPlayerKey = "playerID"
	cookieMaxAge = 365 * 24 * time.Hour
)

// EnsurePlayer identifies the player from an httpOnly cookie, bootstrapping a
// new player (and setting the cookie) when the cookie is missing or invalid.
// The player ID is server-generated only — a tampered cookie that doesn't match
// an existing player row gets a fresh player instead.
func EnsurePlayer(db *gorm.DB, secureCookie bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if raw := c.Cookies(cookieName); raw != "" {
			if id, parseErr := uuid.Parse(raw); parseErr == nil {
				var player model.Player
				err := db.Select("id").First(&player, "id = ?", id).Error
				switch {
				case err == nil:
					c.Locals(ctxPlayerKey, id)
					return c.Next()
				case !errors.Is(err, gorm.ErrRecordNotFound):
					return internalError(c)
				}
			}
		}

		player := model.Player{}
		if err := db.Create(&player).Error; err != nil {
			return internalError(c)
		}
		setPlayerCookie(c, player.ID, secureCookie)
		c.Locals(ctxPlayerKey, player.ID)
		return c.Next()
	}
}

// PlayerID returns the authenticated player's ID set by EnsurePlayer.
func PlayerID(c *fiber.Ctx) uuid.UUID {
	return c.Locals(ctxPlayerKey).(uuid.UUID)
}

func setPlayerCookie(c *fiber.Ctx, id uuid.UUID, secure bool) {
	// Cross-site FE↔BE (Vercel↔Railway) needs SameSite=None which requires Secure.
	// Local dev is same-site over http, so Lax without Secure.
	sameSite := fiber.CookieSameSiteLaxMode
	if secure {
		sameSite = fiber.CookieSameSiteNoneMode
	}
	c.Cookie(&fiber.Cookie{
		Name:     cookieName,
		Value:    id.String(),
		Path:     "/",
		Expires:  time.Now().Add(cookieMaxAge),
		HTTPOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func internalError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": fiber.Map{"code": "INTERNAL", "message": "internal server error"},
	})
}
