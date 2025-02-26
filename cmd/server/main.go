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
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting 7x42 application...")

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
	log.Println("Connecting to database...")
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connection established")

	// Run migrations
	log.Println("Running database migrations...")
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	log.Println("Database migrations completed")

	// Initialize AI service directly with configuration
	log.Println("Initializing AI service...")
	aiConfig := ai.Config{
		OpenRouterKey: getEnv("OPENROUTER_API_KEY", ""),
		Model:         getEnv("OPENROUTER_MODEL", "google/gemini-2.0-flash-001"),
		DB:            db,
	}

	aiService, err := ai.NewServiceWithConfig(aiConfig)
	if err != nil {
		log.Fatal("Failed to initialize AI service:", err)
	}

	// Initialize server
	log.Println("Initializing server...")
	app := server.New(&server.Config{
		DB:        db,
		AIService: aiService,
	})
	log.Println("Server initialized")

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Graceful shutdown initiated...")

		// Attempt to shut down gracefully
		if err := app.Shutdown(); err != nil {
			log.Fatal("Error during shutdown:", err)
		}

		log.Println("Shutdown completed successfully")
		os.Exit(0)
	}()

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Server error:", err)
	}
}

// getEnv retrieves environment variables with fallback values
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
