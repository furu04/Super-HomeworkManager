package main

import (
	"flag"
	"log"
	"time"

	"homework-manager/internal/config"
	"homework-manager/internal/database"
	"homework-manager/internal/router"
	"homework-manager/internal/service"
)

func main() {
	configPath := flag.String("config", "", "Path to config.ini file (default: config.ini in current directory)")
	flag.Parse()

	cfg := config.Load(*configPath)

	log.Printf("Connecting to database (driver: %s)", cfg.Database.Driver)
	if err := database.Connect(cfg.Database, cfg.Debug); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	recurringService := service.NewRecurringAssignmentService()
	if err := recurringService.GenerateNextAssignments(); err != nil {
		log.Printf("recurring generation error on startup: %v", err)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("recurring scheduler panic: %v", r)
			}
		}()
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := recurringService.GenerateNextAssignments(); err != nil {
				log.Printf("recurring generation error: %v", err)
			}
		}
	}()

	r := router.Setup(cfg)

	log.Printf("Server starting on http://localhost:%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
