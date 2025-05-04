// cmd/worker/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nishithcs/bank-info/internal/config"
	"github.com/Nishithcs/bank-info/internal/queue"
	"github.com/Nishithcs/bank-info/internal/repository/mongo"
	"github.com/Nishithcs/bank-info/internal/repository/postgres"
	"github.com/Nishithcs/bank-info/internal/service"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	// Load configuration
	cfg := loadConfig()

	// Connect to PostgreSQL
	db := connectPostgres(cfg.PostgresURL)
	defer db.Close()

	// Connect to MongoDB
	mongoClient := connectMongo(cfg.MongoURI)
	mongodb := mongoClient.Database(cfg.MongoDB)

	// Connect to RabbitMQ
	queueService, err := queue.NewRabbitMQService(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer queueService.Close()

	// Initialize repositories
	accountRepo := postgres.NewAccountRepository(db)
	transactionRepo := mongo.NewTransactionRepository(mongodb)

	// Initialize services
	accountService := service.NewAccountService(accountRepo, transactionRepo, queueService)

	// Setup task handlers
	setupConsumers(queueService, accountService)

	// Wait for interrupt signal to gracefully shut down the worker
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Worker is shutting down...")
}

func setupConsumers(queueService queue.QueueService, accountService service.AccountService) {
	// Handler for account creation tasks
	err := queueService.Consume(queue.AccountCreationQueue, func(payload []byte) error {
		ctx := context.Background()
		log.Printf("Processing account creation task")
		
		err := accountService.ProcessAccountCreation(ctx, payload)
		if err != nil {
			log.Printf("Failed to process account creation: %v", err)
			return err
		}
		
		return nil
	})
	
	if err != nil {
		log.Fatalf("Failed to register account creation consumer: %v", err)
	}

	// Handler for transaction tasks
	err = queueService.Consume(queue.TransactionQueue, func(payload []byte) error {
		ctx := context.Background()
		
		var task queue.Task
		if err := json.Unmarshal(payload, &task); err != nil {
			return err
		}
		
		log.Printf("Processing transaction task: %s", task.ID)
		
		err := accountService.ProcessTransactionTask(ctx, payload)
		if err != nil {
			log.Printf("Failed to process transaction: %v", err)
			return err
		}
		
		return nil
	})
	
	if err != nil {
		log.Fatalf("Failed to register transaction consumer: %v", err)
	}

	log.Println("Workers registered and consuming tasks")
}

func loadConfig() *config.Config {
	// For simplicity, using environment variables
	// In a production app, consider using a config library like viper
	return &config.Config{
		PostgresURL:  getEnv("POSTGRES_URL", "postgres://postgres:postgres@postgres:5432/bankdb?sslmode=disable"),
		MongoURI:     getEnv("MONGO_URI", "mongodb://mongo:27017"),
		MongoDB:      getEnv("MONGO_DB", "bankdb"),
		RabbitMQURL:  getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
	}
}

func connectPostgres(url string) *sqlx.DB {
	var db *sqlx.DB
	var err error
	
	// Retry connection with backoff
	for retries := 5; retries > 0; retries-- {
		db, err = sqlx.Connect("postgres", url)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to PostgreSQL, retrying in 5 seconds: %v", err)
		time.Sleep(5 * time.Second)
	}
	
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	return db
}

func connectMongo(uri string) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Retry connection with backoff
	var client *mongo.Client
	var err error
	
	for retries := 5; retries > 0; retries-- {
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err == nil {
			break
		}
		log.Printf("Failed to connect to MongoDB, retrying in 5 seconds: %v", err)
		time.Sleep(5 * time.Second)
	}
	
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ping to verify connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	return client
}

// Helper functions for environment variables
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}