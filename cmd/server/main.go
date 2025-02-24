package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hra42/7x42/internal/database"
	"github.com/hra42/7x42/internal/server"
)

func main() {
	// Initialize database connection
	dbConfig := &database.Config{
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		Port:     os.Getenv("DB_PORT"),
		SSLMode:  "disable",
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize server
	app := server.New(&server.Config{
		DB: db,
	})

	// Setup graceful shutdown
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
	log.Println("Server starting on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatal("Server error:", err)
	}
}
