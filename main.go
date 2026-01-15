package main

import (
	"log"

	"github.com/DucLUT/goodstuff/config"
	"github.com/DucLUT/goodstuff/models"
	"github.com/DucLUT/goodstuff/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	config.Load()

	// Set Gin mode
	gin.SetMode(config.AppConfig.GinMode)

	// Initialize database
	config.InitDatabase()

	// Auto-migrate models
	config.AutoMigrate(
		&models.User{},
		&models.Worker{},
		&models.ServiceCategory{},
		&models.Service{},
		&models.Booking{},
		&models.Review{},
	)

	// Setup router
	r := routes.SetupRouter()

	// Start server
	log.Printf("Server starting on port %s", config.AppConfig.Port)
	if err := r.Run(":" + config.AppConfig.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
