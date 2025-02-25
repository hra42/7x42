package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hra42/7x42/internal/ai"
	"github.com/hra42/7x42/internal/database"
	"github.com/hra42/7x42/internal/server"
)

func main() {
	// Database configuration
	dbConfig := &database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		User:     getEnv("DB_USER", "7x42user"),
		Password: getEnv("DB_PASSWORD", "7x42pass"),
		DBName:   getEnv("DB_NAME", "7x42db"),
		Port:     getEnv("DB_PORT", "5432"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Connect to database
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize AI service
	aiService, err := ai.NewService(db)
	if err != nil {
		log.Fatal("Failed to initialize AI service:", err)
	}

	// Initialize server
	app := server.New(&server.Config{
		DB:        db,
		AIService: aiService,
	})

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		if err := app.Shutdown(); err != nil {
			log.Fatal("Error during shutdown:", err)
		}
	}()

	// Start server
	port := getEnv("PORT", "8080")
	log.Println("Server starting on :" + port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Server error:", err)
	}
}

// Helper function to get environment variable with fallback
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
