package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"csv-processor/backend/internal/repository"
)

type DownloadHandler struct {
	jobRepo *repository.JobRepository
}

func NewDownloadHandler(jobRepo *repository.JobRepository) *DownloadHandler {
	return &DownloadHandler{
		jobRepo: jobRepo,
	}
}

func (h *DownloadHandler) Download(c *gin.Context) {
	jobIDStr := c.Param("jobId")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}

	// Get job details
	job, err := h.jobRepo.GetJob(jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	if job.Status != "done" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job not completed yet"})
		return
	}

	if !job.OutputFilePath.Valid || job.OutputFilePath.String == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "output file not found"})
		return
	}

	// Check if file exists
	if _, err := os.Stat(job.OutputFilePath.String); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "output file missing"})
		return
	}

	// Serve file for download
	c.FileAttachment(job.OutputFilePath.String, "processed_transactions.csv")
}

func (h *DownloadHandler) DownloadDamagedRows(c *gin.Context) {
	jobIDStr := c.Param("jobId")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}

	// Get damaged rows path
	damagedRowsPath, err := h.jobRepo.GetDamagedRowsPath(jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if !damagedRowsPath.Valid || damagedRowsPath.String == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no damaged rows file available"})
		return
	}

	// Check if file exists
	if _, err := os.Stat(damagedRowsPath.String); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "damaged rows file missing"})
		return
	}

	// Serve file for download
	c.FileAttachment(damagedRowsPath.String, "damaged_rows.csv")
}
