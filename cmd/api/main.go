// cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nishithcs/bank-info/internal/api"
	"github.com/Nishithcs/bank-info/internal/config"
	"github.com/Nishithcs/bank-info/internal/queue"
	"github.com/Nishithcs/bank-info/internal/repository/mongo"
	"github.com/Nishithcs/bank-info/internal/repository/postgres"
	"github.com/Nishithcs/bank-info/internal/service"
	"github.com/gorilla/mux"
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

	// Initialize HTTP handler
	handler := api.NewHandler(accountService)

	// Setup router with middleware
	router := mux.NewRouter()
	router.Use(loggingMiddleware)
	
	// Register routes
	handler.RegisterRoutes(router)

	// Configure HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on port %d", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Server is shutting down...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

func loadConfig() *config.Config {
	// For simplicity, using environment variables
	// In a production app, consider using a config library like viper
	return &config.Config{
		ServerPort:   getEnvAsInt("SERVER_PORT", 8080),
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

	// Ensure tables exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			balance DECIMAL(15,2) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
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

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

// Helper functions for environment variables
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := parseInt(value); err == nil {
			return i
		}
	}
	return fallback
}

func parseInt(value string) (int, error) {
	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	return result, err
}