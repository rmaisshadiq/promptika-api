package config

import (
	"fmt"
	"log"
	"os"

	"github.com/rmaisshadiq/critical-prompt-api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance used throughout the application.
var DB *gorm.DB

// ConnectDatabase initializes the GORM connection to Neon PostgreSQL
// and runs auto-migration for all models.
func ConnectDatabase() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.PromptLog{},
		&models.Report{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	DB = db
	fmt.Println("✅ Database connected and migrated successfully")
}
