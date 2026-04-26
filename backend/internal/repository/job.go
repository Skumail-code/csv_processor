package repository

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"os"
	"time"

	"csv-processor/backend/internal/model"

	"github.com/google/uuid"
)

type JobRepository struct {
	db *sql.DB
}

func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Create(job *model.Job) error {
	query := `
        INSERT INTO jobs (id, file_hash, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	now := time.Now()
	_, err := r.db.Exec(query, job.ID, job.FileHash, job.Status, now, now)
	return err
}

func (r *JobRepository) FindByHash(hash string) (*model.Job, error) {
	query := `SELECT id, status FROM jobs WHERE file_hash = $1`
	row := r.db.QueryRow(query, hash)

	var job model.Job
	err := row.Scan(&job.ID, &job.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &job, err
}

func ComputeFileHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (r *JobRepository) UpdateStatus(id uuid.UUID, status string) error {
	query := `UPDATE jobs SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *JobRepository) UpdateStatusWithError(id uuid.UUID, status string, errMsg string) error {
	query := `UPDATE jobs SET status = $1, error_message = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(query, status, errMsg, id)
	return err
}

func (r *JobRepository) MarkAsDone(id uuid.UUID) error {
	query := `UPDATE jobs SET status = 'done', updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *JobRepository) GetJob(id uuid.UUID) (*model.Job, error) {
	query := `SELECT id, status, total_rows, processed_rows, invalid_rows, output_file_path, error_message, created_at, updated_at 
              FROM jobs WHERE id = $1`

	var job model.Job
	err := r.db.QueryRow(query, id).Scan(
		&job.ID, &job.Status, &job.TotalRows, &job.ProcessedRows,
		&job.InvalidRows, &job.OutputFilePath, &job.ErrorMessage,
		&job.CreatedAt, &job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &job, err
}

func (r *JobRepository) UpdateProgress(id uuid.UUID, processedRows, invalidRows, totalRows int) error {
	query := `UPDATE jobs 
              SET processed_rows = $1, invalid_rows = $2, total_rows = $3, updated_at = NOW() 
              WHERE id = $4`
	_, err := r.db.Exec(query, processedRows, invalidRows, totalRows, id)
	return err
}

func (r *JobRepository) UpdateOutputPath(id uuid.UUID, outputPath string) error {
	query := `UPDATE jobs SET output_file_path = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, outputPath, id)
	return err
}

func (r *JobRepository) GetOutputPath(id uuid.UUID) (string, error) {
	var outputPath string
	query := `SELECT output_file_path FROM jobs WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&outputPath)
	return outputPath, err
}

func (r *JobRepository) UpdateDamagedRowsPath(id uuid.UUID, damagedRowsPath string) error {
	query := `UPDATE jobs SET damaged_rows_path = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, damagedRowsPath, id)
	return err
}

func (r *JobRepository) GetDamagedRowsPath(id uuid.UUID) (sql.NullString, error) {
	var path sql.NullString
	query := `SELECT damaged_rows_path FROM jobs WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&path)
	return path, err
}
