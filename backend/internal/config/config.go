package config

import (
	"os"
	"strings"
)

type Config struct {
	DatabaseURL    string
	Port           string
	FrontendOrigin string
	// CookieSecure is derived from FrontendOrigin: an https frontend means a
	// cross-site (Vercel↔Railway) deployment, which needs Secure + SameSite=None.
	// Deriving it removes the COOKIE_SECURE footgun where a wrong/missing value
	// silently broke the cross-site cookie in production.
	CookieSecure bool
}

// Load reads configuration from environment variables.
// Defaults match docker-compose.yml for zero-config local development;
// production values are injected by the deployment platform.
func Load() Config {
	origin := getEnv("FRONTEND_ORIGIN", "http://localhost:3000")
	return Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://points:points_local@localhost:5432/points_game?sslmode=disable"),
		Port:           getEnv("PORT", "8080"),
		FrontendOrigin: origin,
		CookieSecure:   strings.HasPrefix(origin, "https://"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
