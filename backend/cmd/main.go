package main

import (
	"log"
	"net/http"

	"csv-processor/backend/cmd/api"
	"csv-processor/backend/internal/config"
	database "csv-processor/backend/internal/db"

	"csv-processor/backend/migrations"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("Starting CSV Processor in %s mode", cfg.Environment)
	log.Printf("Connecting to database at %s:%d", cfg.DBHost, cfg.DBPort)

	// Validate required config
	if cfg.DBPassword == "" && cfg.Environment == "production" {
		log.Fatal("DB_PASSWORD is required in production environment")
	}

	// Connect to database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := migrations.Run(db.DB, cfg); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	log.Println("Migrations completed successfully")

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	r := setupRouter(db, cfg)

	// Start server
	log.Printf("Server starting on port %s in %s mode", cfg.ServerPort, cfg.Environment)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRouter(db *database.Database, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Health check endpoint with database check
	r.GET("/health", func(c *gin.Context) {
		if err := db.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"ok":     false,
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ok":          true,
			"status":      "healthy",
			"environment": cfg.Environment,
		})
	})

	// Setup API routes
	api.SetupRoutes(r, db, cfg)

	return r
}
