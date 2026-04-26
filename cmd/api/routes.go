package api

import (
	"net/http"

	"csv-processor/internal/config"
	database "csv-processor/internal/db"
	handler "csv-processor/internal/handlers"
	"csv-processor/internal/repository"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

func SetupRoutes(r *gin.Engine, db *database.Database, cfg *config.Config) {
	// Enable CORS for frontend
	r.Use(cors.Default())

	// Initialize repositories
	jobRepo := repository.NewJobRepository(db.DB)

	// Initialize Asynq client
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisAddr})

	// Initialize handlers
	uploadHandler := handler.NewUploadHandler(jobRepo, asynqClient)
	downloadHandler := handler.NewDownloadHandler(jobRepo)

	// Routes (Phase 0 contract)
	r.POST("/upload", uploadHandler.Upload)

	r.GET("/status/:jobId", func(c *gin.Context) {
		jobIDStr := c.Param("jobId")
		jobID, err := uuid.Parse(jobIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
			return
		}

		job, err := jobRepo.GetJob(jobID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if job == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"jobId":         job.ID,
			"status":        job.Status,
			"totalRows":     job.TotalRows,
			"processedRows": job.ProcessedRows,
			"invalidRows":   job.InvalidRows,
			"errorMessage":  job.ErrorMessage,
			"createdAt":     job.CreatedAt,
			"updatedAt":     job.UpdatedAt,
		})
	})

	r.GET("/download/:jobId", downloadHandler.Download)
}
