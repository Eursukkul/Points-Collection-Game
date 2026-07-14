package middleware

import (
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/apierr"
	"github.com/gofiber/fiber/v2"
)

// CSRFGuard rejects state-changing requests whose Origin isn't the allowed
// frontend. The player cookie is SameSite=None in production, so the browser
// sends it on cross-site requests — an Origin check is what stops a page on
// another site from silently POSTing /reset or /game/play with the victim's
// cookie. Requests with no Origin header (curl, Apidog, server-to-server) carry
// no ambient browser credentials and are allowed so API testing still works.
func CSRFGuard(allowedOrigin string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if isSafeMethod(c.Method()) {
			return c.Next()
		}
		if origin := c.Get(fiber.HeaderOrigin); origin != "" && origin != allowedOrigin {
			return apierr.Respond(c, fiber.StatusForbidden, "FORBIDDEN_ORIGIN", "cross-origin request rejected")
		}
		return c.Next()
	}
}
