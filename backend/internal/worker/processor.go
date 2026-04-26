package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"csv-processor/backend/internal/processor"
	"csv-processor/backend/internal/repository"
	"csv-processor/backend/internal/task"
)

type Processor struct {
	jobRepo *repository.JobRepository
}

func NewProcessor(jobRepo *repository.JobRepository) *Processor {
	return &Processor{
		jobRepo: jobRepo,
	}
}

func (p *Processor) HandleProcessCSVTask(ctx context.Context, t *asynq.Task) error {
	// Parse payload
	var payload task.ProcessCSVPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	jobID, err := uuid.Parse(payload.JobID)
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}

	log.Printf("Processing job %s for file: %s", jobID, payload.FilePath)

	// Update status to processing
	if err := p.jobRepo.UpdateStatus(jobID, "processing"); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Create CSV processor
	csvProcessor := processor.NewCSVProcessor(payload.FilePath)

	// Process CSV with progress updates
	err = csvProcessor.Process(func(processed, valid, invalid int) {
		// Update progress every 100 rows
		if err := p.jobRepo.UpdateProgress(jobID, processed, invalid, processed); err != nil {
			log.Printf("Warning: failed to update progress: %v", err)
		}

		log.Printf("Job %s: Processed %d rows (Valid: %d, Invalid: %d)",
			jobID, processed, valid, invalid)
	})

	if err != nil {
		// Update job as failed
		errorMsg := fmt.Sprintf("Processing failed: %v", err)
		p.jobRepo.UpdateStatusWithError(jobID, "failed", errorMsg)

		// Clean up temp file
		os.Remove(payload.FilePath)

		return fmt.Errorf("processing failed: %w", err)
	}

	// Generate output file
	outputDir := "/tmp/processed"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		errorMsg := fmt.Sprintf("Failed to create output dir: %v", err)
		p.jobRepo.UpdateStatusWithError(jobID, "failed", errorMsg)
		os.Remove(payload.FilePath)
		return fmt.Errorf("%s", errorMsg)
	}

	outputPath, err := csvProcessor.GenerateOutputFile(outputDir)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to generate output file: %v", err)
		p.jobRepo.UpdateStatusWithError(jobID, "failed", errorMsg)
		os.Remove(payload.FilePath)
		return fmt.Errorf("%s", errorMsg)
	}

	// Save output path and final stats
	if err := p.jobRepo.UpdateOutputPath(jobID, outputPath); err != nil {
		log.Printf("Warning: failed to save output path: %v", err)
	}

	// Generate damaged rows file if there are invalid rows
	damagedRowsPath, err := csvProcessor.GenerateInvalidRowsFile(outputDir)
	if err != nil {
		log.Printf("Warning: failed to generate damaged rows file: %v", err)
	} else if damagedRowsPath != "" {
		if err := p.jobRepo.UpdateDamagedRowsPath(jobID, damagedRowsPath); err != nil {
			log.Printf("Warning: failed to save damaged rows path: %v", err)
		}
	}

	summary := csvProcessor.GetSummary()
	if err := p.jobRepo.UpdateProgress(jobID,
		summary["total_rows"].(int),
		summary["invalid_rows"].(int),
		summary["total_rows"].(int)); err != nil {
		log.Printf("Warning: failed to update final progress: %v", err)
	}

	// Mark job as done
	if err := p.jobRepo.MarkAsDone(jobID); err != nil {
		return fmt.Errorf("failed to mark job done: %w", err)
	}

	// Clean up temp file
	os.Remove(payload.FilePath)

	log.Printf("Job %s: Completed! Summary: %+v", jobID, summary)

	return nil
}
