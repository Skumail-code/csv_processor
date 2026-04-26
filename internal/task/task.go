package task

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeProcessCSV = "process:csv"
)

type ProcessCSVPayload struct {
	JobID    string `json:"job_id"`
	FilePath string `json:"file_path"`
	FileHash string `json:"file_hash"`
}

func NewProcessCSVTask(jobID uuid.UUID, filePath, fileHash string) (*asynq.Task, error) {
	payload := ProcessCSVPayload{
		JobID:    jobID.String(),
		FilePath: filePath,
		FileHash: fileHash,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeProcessCSV, payloadBytes), nil
}
