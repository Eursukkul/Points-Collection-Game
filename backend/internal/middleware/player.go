package middleware

import (
	"errors"
	"time"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/apierr"
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

// EnsurePlayer identifies the player from an httpOnly cookie. On a state-changing
// request without a valid player it bootstraps one (and sets the cookie); on a
// safe (read-only) request it does NOT create a row — that would let cookieless
// traffic (bots, uptime probes) grow the players table without bound. Read
// handlers treat a nil player as empty state.
//
// The player ID is server-generated only — a tampered cookie that doesn't match
// an existing player row is ignored.
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
					return apierr.Internal(c)
				}
			}
		}

		if isSafeMethod(c.Method()) {
			c.Locals(ctxPlayerKey, uuid.Nil)
			return c.Next()
		}

		player := model.Player{}
		if err := db.Create(&player).Error; err != nil {
			return apierr.Internal(c)
		}
		setPlayerCookie(c, player.ID, secureCookie)
		c.Locals(ctxPlayerKey, player.ID)
		return c.Next()
	}
}

// PlayerID returns the player's ID set by EnsurePlayer, or uuid.Nil when a safe
// request arrived without a valid player cookie.
func PlayerID(c *fiber.Ctx) uuid.UUID {
	id, _ := c.Locals(ctxPlayerKey).(uuid.UUID)
	return id
}

func isSafeMethod(method string) bool {
	return method == fiber.MethodGet || method == fiber.MethodHead || method == fiber.MethodOptions
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
