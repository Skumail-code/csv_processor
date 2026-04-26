package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"csv-processor/internal/model"
	"csv-processor/internal/repository"
	"csv-processor/internal/task"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type UploadHandler struct {
	jobRepo     *repository.JobRepository
	asynqClient *asynq.Client
}

func NewUploadHandler(jobRepo *repository.JobRepository, asynqClient *asynq.Client) *UploadHandler {
	return &UploadHandler{
		jobRepo:     jobRepo,
		asynqClient: asynqClient,
	}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	// 1. Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}

	// Check extension
	if ext := filepath.Ext(file.Filename); ext != ".csv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only CSV files allowed"})
		return
	}

	// 2. Save temp file to shared uploads directory
	uploadsDir := "/tmp/uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create uploads directory"})
		return
	}

	tmpPath := filepath.Join(uploadsDir, file.Filename)
	if err := c.SaveUploadedFile(file, tmpPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	// 3. Compute hash
	hash, err := repository.ComputeFileHash(tmpPath)
	if err != nil {
		os.Remove(tmpPath) // Clean up on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash file"})
		return
	}

	// 4. Check for existing job (dedup)
	existing, err := h.jobRepo.FindByHash(hash)
	if err != nil {
		os.Remove(tmpPath) // Clean up on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if existing != nil {
		// Return existing job - clean up duplicate file
		os.Remove(tmpPath)
		c.JSON(http.StatusOK, gin.H{
			"jobId":  existing.ID,
			"status": existing.Status,
		})
		return
	}

	// 5. Create new job
	jobID := uuid.New()
	job := &model.Job{
		ID:       jobID,
		FileHash: hash,
		Status:   "queued",
	}

	if err := h.jobRepo.Create(job); err != nil {
		os.Remove(tmpPath) // Clean up on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job"})
		return
	}

	task, err := task.NewProcessCSVTask(jobID, tmpPath, hash)
	if err != nil {
		os.Remove(tmpPath) // Clean up on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	info, err := h.asynqClient.Enqueue(task)
	if err != nil {
		os.Remove(tmpPath) // Clean up on error
		// Also clean up the job record
		h.jobRepo.UpdateStatus(jobID, "failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue job"})
		return
	}
	// Note: File is NOT deleted here - worker will process it and clean up

	c.JSON(http.StatusOK, gin.H{
		"jobId":   job.ID,
		"status":  job.Status,
		"task_id": info.ID,
	})
}
