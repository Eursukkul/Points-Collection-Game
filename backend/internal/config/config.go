package config

import "os"

type Config struct {
	DatabaseURL    string
	Port           string
	FrontendOrigin string
	CookieSecure   bool
}

// Load reads configuration from environment variables.
// Defaults match docker-compose.yml for zero-config local development;
// production values are injected by the deployment platform.
func Load() Config {
	return Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://points:points_local@localhost:5432/points_game?sslmode=disable"),
		Port:           getEnv("PORT", "8080"),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
		CookieSecure:   getEnv("COOKIE_SECURE", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
