//  cmd/api/main.go
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/Nishithcs/bank-info/internal/account"
	"github.com/Nishithcs/bank-info/internal/config"
	"github.com/Nishithcs/bank-info/internal/database"
	"github.com/Nishithcs/bank-info/internal/mq"
	"github.com/Nishithcs/bank-info/internal/transaction"
	"log"
)

func main() {
	//  Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	//  Connect to PostgreSQL
	pgDB, err := database.ConnectPostgres(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer func() {
		db, _ := pgDB.DB()
		db.Close()
	}()

	//  Connect to MongoDB
	mongoDB, err := database.ConnectMongo(cfg.Mongo)
	if err != nil {
		log.Fatalf("Failed to connect to Mongo: %v", err)
	}
	defer mongoDB.Client().Disconnect(nil)

	//  Connect to RabbitMQ
	mqConn, err := mq.Connect(cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer mqConn.Close()

	//  Create RabbitMQ channel
	mqChan, err := mqConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}
	defer mqChan.Close()

	//  Initialize Repositories
	accountRepo := account.NewAccountRepository(pgDB)
	transactionRepo := transaction.NewTransactionRepository(mongoDB.Database(cfg.Mongo.Database))

	//  Initialize Services
	accountService := account.NewAccountService(accountRepo)
	transactionService := transaction.NewTransactionService(transactionRepo)

	//  Initialize MQ Publisher
	mqPublisher := mq.NewPublisher(mqChan)

	//  Initialize Handlers
	accountHandler := account.NewAccountHandler(accountService, mqPublisher)
	transactionHandler := transaction.NewTransactionHandler(transactionService, mqPublisher)

	//  Set up Gin router
	router := gin.Default()

	//  Define API routes
	api := router.Group("/api/v1")
	{
		accounts := api.Group("/accounts")
		{
			accounts.POST("", accountHandler.CreateAccount)
			accounts.GET("/:id", accountHandler.GetAccount)
		}
		transactions := api.Group("/transactions")
		{
			transactions.POST("", transactionHandler.HandleTransaction) //  Deposit/Withdraw
			transactions.GET("/:accountID", transactionHandler.GetTransactionHistory)
		}
	}

	//  Run the API server
	if err := router.Run(fmt.Sprintf(":%s", cfg.API.Port)); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}