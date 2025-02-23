package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hra42/7x42/internal/server"
)

func main() {
	// Initialize server
	app := server.New()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		_ = <-c
		log.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// Start server
	if err := app.Listen(":8080"); err != nil {
		log.Fatal("Server error:", err)
	}
}
