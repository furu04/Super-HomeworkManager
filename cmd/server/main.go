package main

import (
	"flag"
	"log"

	"homework-manager/internal/config"
	"homework-manager/internal/database"
	"homework-manager/internal/router"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to config.ini file (default: config.ini in current directory)")
	flag.Parse()

	// Load configuration
	cfg := config.Load(*configPath)

	// Connect to database
	log.Printf("Connecting to database (driver: %s)", cfg.Database.Driver)
	if err := database.Connect(cfg.Database, cfg.Debug); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Setup router
	r := router.Setup(cfg)

	// Start server
	log.Printf("Server starting on http://localhost:%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
