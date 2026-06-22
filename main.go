package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rmaisshadiq/critical-prompt-api/config"
	"github.com/rmaisshadiq/critical-prompt-api/routes"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Connect to the database and run migrations
	config.ConnectDatabase()

	// Initialize Gin router
	r := gin.Default()

	// Register all routes
	routes.SetupRoutes(r)

	// Determine the port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server starting on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
