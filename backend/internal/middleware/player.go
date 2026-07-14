package middleware

import (
	"errors"
	"net/http"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	cookieName   = "player_id"
	ctxPlayerKey = "playerID"
	cookieMaxAge = 60 * 60 * 24 * 365 // 1 year
)

// EnsurePlayer identifies the player from an httpOnly cookie, bootstrapping a
// new player (and setting the cookie) when the cookie is missing or invalid.
// The player ID is server-generated only — a tampered cookie that doesn't match
// an existing player row gets a fresh player instead.
func EnsurePlayer(db *gorm.DB, secureCookie bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if raw, err := c.Cookie(cookieName); err == nil {
			if id, parseErr := uuid.Parse(raw); parseErr == nil {
				var player model.Player
				err := db.Select("id").First(&player, "id = ?", id).Error
				switch {
				case err == nil:
					c.Set(ctxPlayerKey, id)
					c.Next()
					return
				case !errors.Is(err, gorm.ErrRecordNotFound):
					abortInternal(c)
					return
				}
			}
		}

		player := model.Player{}
		if err := db.Create(&player).Error; err != nil {
			abortInternal(c)
			return
		}
		setPlayerCookie(c, player.ID, secureCookie)
		c.Set(ctxPlayerKey, player.ID)
		c.Next()
	}
}

// PlayerID returns the authenticated player's ID set by EnsurePlayer.
func PlayerID(c *gin.Context) uuid.UUID {
	return c.MustGet(ctxPlayerKey).(uuid.UUID)
}

func setPlayerCookie(c *gin.Context, id uuid.UUID, secure bool) {
	// Cross-site FE↔BE (Vercel↔Railway) needs SameSite=None which requires Secure.
	// Local dev is same-site over http, so Lax without Secure.
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    id.String(),
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func abortInternal(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{"code": "INTERNAL", "message": "internal server error"},
	})
}
