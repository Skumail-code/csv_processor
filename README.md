# CSV Upload Processor

A production-ready web application for uploading, processing, and analyzing CSV transaction files with real-time progress tracking and comprehensive error handling.

## Prerequisites

- **Docker** (v20.10+) - For containerized deployment
- **Docker Compose** (v2.0+) - For orchestrating multi-container setup
- **Git** - For cloning the repository

Optional for local development:
- **Go** (v1.21+) - For backend development
- **Node.js** (v20+) - For frontend development

## Quick Start

```bash
# Clone the repository
git clone https://github.com/Skumail-code/csv_processor.git
cd csv_processor

# Start all services with one command
docker-compose up --build
```

The application will be available at:
- **Frontend UI**: http://localhost:3000
- **API Endpoint**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

## What This Project Achieves

Based on the assignment requirements, this implementation delivers:

### Core Requirements Met

1. **Fast Upload Endpoint** (`POST /upload`)
   - Accepts CSV files and returns `jobId` immediately
   - Files are queued for background processing, never blocking the request
   - SHA256 hash-based deduplication prevents duplicate jobs for identical files

2. **Real-time Status Tracking** (`GET /status/:jobId`)
   - Returns job state: `queued`, `processing`, `done`, or `failed`
   - Live progress updates: rows processed out of total
   - Result summary with valid/invalid row counts
   - Invalid rows are counted and surfaced, never silently dropped

3. **Background Processing**
   - Asynq task queue with Redis for reliable job distribution
   - Separate worker containers for horizontal scaling
   - Job persistence in PostgreSQL for durability

4. **Frontend Experience**
   - Upload files with drag-and-drop
   - Running list of jobs with real-time status
   - Visual progress bars showing processing state
   - **Download processed CSV** (contains only valid rows - damaged rows eliminated)
   - **Download damaged rows CSV separately** - invalid rows with error reasons for easy review and correction
   - Dark/Light mode toggle with glassmorphism UI

5. **Infrastructure**
   - Single-command startup via `docker-compose up`
   - Full stack: API, Worker, PostgreSQL, Redis, Frontend
   - GitHub Actions CI/CD pipeline for lint, test, and build

### Error Handling & Edge Cases

- **Malformed rows**: Captured with specific error messages (invalid date, missing columns, etc.)
- **Worker crashes**: Jobs persist in DB, can be re-queued
- **Duplicate uploads**: SHA256 hash detection prevents re-processing
- **File persistence**: Shared Docker volumes ensure API and worker can access files
- **Queue failures**: Redis persistence + job status tracking in PostgreSQL

### Future Improvements

If I had more time, I would harden the row-validation logic further and make the error handling more exhaustive.

- Tighten validation rules for every field combination and edge case
- Add more test coverage for malformed CSV inputs and tricky row values
- Make validation rules schema-driven instead of hard-coded
- Add clearer reporting for rows that fail multiple validation checks

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                            │
│  ┌─────────────┐                                                │
│  │   Browser   │──▶ Upload CSV → Get jobId immediately          │
│  │  (React)    │──▶ Poll status → See real-time progress        │
│  └─────────────┘──▶ Download processed/damaged files             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         API LAYER                               │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐     │
│  │  Gin Router │───▶│   Handler   │───▶│  Asynq Client   │     │
│  │   (/api)    │    │  (Upload)   │    │  (Enqueue Job)  │     │
│  └─────────────┘    └─────────────┘    └────────┬────────┘     │
│        │                                         │              │
│        │    ┌─────────────┐    ┌─────────────┐   │              │
│        └───▶│  Job Repo   │◀───│   Status    │◀──┘              │
│             │ (PostgreSQL)│    │   Handler   │                   │
│             └─────────────┘    └─────────────┘                   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         QUEUE LAYER                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐       │
│  │    Redis    │◀───│   Asynq     │───▶│  Task Payload   │       │
│  │   (Queue)   │    │  (Broker)   │    │  (Job Metadata) │       │
│  └─────────────┘    └──────┬──────┘    └─────────────────┘       │
└────────────────────────────┼────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                       WORKER LAYER                                │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐       │
│  │   Asynq     │───▶│  Processor  │───▶│   CSV Parser    │       │
│  │   Server    │    │             │    │                 │       │
│  └─────────────┘    └──────┬──────┘    └─────────────────┘       │
│                            │                                    │
│                            ▼                                    │
│              ┌─────────────────────────────┐                     │
│              │     CSV Processing Flow     │                     │
│              │  1. Parse row-by-row        │                     │
│              │  2. Validate each field     │                     │
│              │  3. Categorize valid/invalid│                     │
│              │  4. Write output files      │                     │
│              │  5. Update job progress     │                     │
│              └─────────────┬───────────────┘                     │
│                            │                                    │
│              ┌─────────────┴───────────────┐                   │
│              ▼                             ▼                    │
│  ┌─────────────────────┐    ┌─────────────────────┐             │
│  │  Valid Rows CSV     │    │ Invalid Rows CSV    │             │
│  │  (processed_)       │    │  (damaged_)         │             │
│  └─────────────────────┘    └─────────────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

## API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/upload` | POST | Upload CSV file, returns `{jobId, status, task_id}` |
| `/status/:jobId` | GET | Get job status, progress, and results |
| `/download/:jobId` | GET | Download processed CSV (valid rows only) |
| `/download/:jobId/damaged` | GET | Download CSV with only invalid rows |
| `/health` | GET | Service health check |

### Example: Upload Response
```json
{
  "jobId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "queued",
  "task_id": "task_12345"
}
```

### Example: Status Response
```json
{
  "jobId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "totalRows": 1000,
  "processedRows": 750,
  "invalidRows": 12,
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:31:15Z"
}
```

## CSV Format Requirements

Expected columns (in order):
1. `id` - Integer transaction ID
2. `date` - Date in YYYY-MM-DD format
3. `description` - Non-empty string
4. `amount` - Numeric value (supports ₹ symbol and commas)
5. `category` - Non-empty string (any category value accepted)

## Project Structure

```
csv_processor/
├── backend/                    # Go backend services
│   ├── cmd/
│   │   ├── api/               # HTTP API handlers
│   │   ├── main.go            # API server entry point
│   │   └── worker/            # Background worker entry point
│   ├── internal/
│   │   ├── config/            # Environment configuration
│   │   ├── db/                # Database connection with retry logic
│   │   ├── handlers/          # HTTP request handlers
│   │   │   ├── upload.go      # File upload with deduplication
│   │   │   ├── download.go    # File download (processed + damaged)
│   │   │   └── status.go      # Job status endpoint
│   │   ├── model/             # Data models (Job, Transaction)
│   │   ├── processor/         # CSV parsing and validation logic
│   │   ├── repository/        # Database queries (PostgreSQL)
│   │   ├── task/              # Asynq task definitions
│   │   └── worker/            # Background job processor
│   ├── migrations/            # SQL schema migrations
│   ├── Dockerfile             # API container image
│   ├── Dockerfile.worker      # Worker container image
│   ├── go.mod
│   └── go.sum
├── frontend/                   # React + TypeScript frontend
│   ├── src/
│   │   ├── components/
│   │   │   ├── FileUpload.tsx # Drag-and-drop upload
│   │   │   └── JobCard.tsx    # Job status with progress bar
│   │   ├── hooks/
│   │   │   └── useTheme.ts    # Dark/light mode toggle
│   │   ├── types/
│   │   │   └── index.ts       # TypeScript interfaces
│   │   ├── App.tsx            # Main application
│   │   └── index.css          # Glassmorphism styles
│   ├── Dockerfile             # Frontend container (nginx)
│   ├── nginx.conf             # Reverse proxy config
│   └── package.json
├── .github/workflows/
│   └── ci.yml                 # GitHub Actions CI/CD
├── docker-compose.yml         # Full stack orchestration
└── README.md                  # This file
```

## Design Decisions

### 1. Asynq over Managed Queue Services
**Decision**: Used Asynq (Redis-based) instead of SQS/Cloud Tasks  
**Rationale**: Assignment constraint to "wire pieces together yourself" - demonstrates understanding of queue mechanics, task distribution, and worker coordination without managed service abstraction

### 2. File Deduplication via SHA256
**Decision**: Hash file content on upload, check against existing jobs  
**Rationale**: Prevents wasted compute on identical files; allows immediate return of existing jobId if file already processed

### 3. Separate Damaged Rows Download
**Decision**: Generate two output files (valid + invalid rows)  
**Rationale**: Users can fix invalid data and re-upload without losing the valid portion; clear separation of concerns

### 4. Polling over Websockets
**Decision**: HTTP polling every 2 miliseconds for status updates  
**Rationale**: Simpler infrastructure, works through proxies/firewalls, sufficient for this use case.

### 5. Shared Docker Volumes
**Decision**: API and workers share `/tmp/uploads` and `/tmp/processed`  
**Rationale**: Eliminates need for object storage (S3) in this scale; files are ephemeral after processing

## CI/CD Pipeline

GitHub Actions workflow (`./github/workflows/ci.yml`):

1. **Backend CI**
   - `go vet` - Static analysis
   - `golint` - Style checking
   - `go test -race` - Unit tests with race detection
   - Build API and Worker binaries

2. **Frontend CI**
   - `npm run lint` - ESLint checking
   - `npm run build` - Production build

3. **Docker Build** (on main branch)
   - Build and push API image
   - Build and push Worker image
   - Build and push Frontend image

## What I'd Add With More Time

### 1. Server-Sent Events (SSE) for Real-time Updates
Current implementation polls every 2 seconds. With SSE:
- Server pushes updates immediately when progress changes
- Reduces HTTP overhead and latency
- Cleaner architecture for real-time features

### 2. File Chunking for Large Uploads
For files > 100MB:
- Client-side chunking with resumable uploads
- Parallel chunk upload for faster transfer
- Progress tracking at byte-level, not just row-level
- Handles network interruptions gracefully

### 3. Job Retry with Exponential Backoff
Current: Failed jobs stay failed  
Improvement:
- Automatic retry with exponential backoff (1s, 2s, 4s, 8s...)
- Dead letter queue for permanently failed jobs
- Admin dashboard to replay failed jobs

### 4. Batch Upload & Processing
- Upload multiple files in single request
- Group jobs into batches with aggregate progress
- Bulk download all results as ZIP

### 5. CSV Preview & Column Mapping
- Preview first 10 rows before processing
- Allow users to map columns if headers don't match exactly
- Detect encoding issues (UTF-8 vs Latin-1)

### 6. Authentication & Multi-tenancy
- JWT-based auth with refresh tokens
- Users only see their own jobs
- Quota management (max files per hour)

### 7. Analytics Dashboard
- Processing time trends
- Error rate by category
- Most common validation failures
- Peak usage patterns

### 8. Webhook Notifications
- Callback to user's API when job completes
- Slack/Discord integration options
- Email notifications for long-running jobs

### 9. Streaming Processing
Current: Loads entire file into memory  
Improvement:
- Stream rows from disk
- Process files larger than available RAM
- Lower memory footprint for workers

### 10. Testing Improvements
- Integration tests with testcontainers
- Load testing with k6 (1000 concurrent uploads)
- Frontend E2E tests with Playwright
- Contract testing for API stability

## Development

### Backend Development

```bash
cd backend

# Download dependencies
go mod download

# Run tests
go test ./...

# Run API server
go run cmd/main.go

# Run worker (separate terminal)
go run cmd/worker/main.go
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Start dev server
npm run dev

# Build for production
npm run build
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | postgres | PostgreSQL hostname |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | postgres | Database user |
| `DB_PASSWORD` | postgres | Database password |
| `DB_NAME` | csv_jobs | Database name |
| `REDIS_ADDR` | redis:6379 | Redis connection string |
| `SERVER_PORT` | 8080 | API server port |
| `ENVIRONMENT` | development | Environment mode |

## License

MIT

---

**Built for the CSV Processor Take-Home Assignment**  
Time invested: ~6 hours for core functionality + additional time for UI polish
