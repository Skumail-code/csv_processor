package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID             uuid.UUID
	FileHash       string
	Status         string
	TotalRows      int
	ProcessedRows  int
	InvalidRows    int
	OutputFilePath sql.NullString
	ErrorMessage   sql.NullString
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
