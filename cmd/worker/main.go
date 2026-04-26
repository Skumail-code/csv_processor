package main

import (
	"log"

	"csv-processor/internal/config"
	database "csv-processor/internal/db"
	"csv-processor/internal/repository"
	"csv-processor/internal/worker"

	"github.com/hibiken/asynq"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("Starting CSV Processor Worker")
	log.Printf("Connecting to database at %s:%d", cfg.DBHost, cfg.DBPort)

	// Connect to database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize repository
	jobRepo := repository.NewJobRepository(db.DB)

	// Create Asynq server
	redisAddr := cfg.RedisAddr
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	// Create processor
	processor := worker.NewProcessor(jobRepo)

	// Register handler
	mux := asynq.NewServeMux()
	mux.HandleFunc("process:csv", processor.HandleProcessCSVTask)

	log.Println("Worker starting, listening on Redis:", redisAddr)
	if err := srv.Run(mux); err != nil {
		log.Fatal("Failed to run worker:", err)
	}
}
