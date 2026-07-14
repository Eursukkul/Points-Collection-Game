package main

import (
	"log"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/config"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/database"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/server"
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

	app := server.New(db, cfg)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
