package middleware

import (
	"context"
	"time"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/apierr"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	cookieName   = "player_id"
	ctxPlayerKey = "playerID"
	cookieMaxAge = 365 * 24 * time.Hour
)

// PlayerResolver maps a cookie to a player, creating one only when allowed.
// Implemented by the use case (dependency inversion — middleware declares its need).
type PlayerResolver interface {
	ResolvePlayer(ctx context.Context, cookieID uuid.UUID, hasCookie, allowCreate bool) (uuid.UUID, bool, error)
}

// EnsurePlayer identifies the player from an httpOnly cookie. On a state-changing
// request without a valid player it bootstraps one (and sets the cookie); on a
// safe (read-only) request it does NOT create a row — that would let cookieless
// traffic (bots, probes) grow the players table. Read handlers treat uuid.Nil
// as empty state.
func EnsurePlayer(resolver PlayerResolver, secureCookie bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cookieID, hasCookie := parseCookie(c.Cookies(cookieName))
		allowCreate := !isSafeMethod(c.Method())

		id, created, err := resolver.ResolvePlayer(c.UserContext(), cookieID, hasCookie, allowCreate)
		if err != nil {
			return apierr.Internal(c)
		}
		if created {
			setPlayerCookie(c, id, secureCookie)
		}
		c.Locals(ctxPlayerKey, id)
		return c.Next()
	}
}

// PlayerID returns the player's ID set by EnsurePlayer, or uuid.Nil when a safe
// request arrived without a valid player cookie.
func PlayerID(c *fiber.Ctx) uuid.UUID {
	id, _ := c.Locals(ctxPlayerKey).(uuid.UUID)
	return id
}

func parseCookie(raw string) (uuid.UUID, bool) {
	if raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
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
