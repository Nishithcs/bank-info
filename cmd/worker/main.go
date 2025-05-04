//  cmd/worker/main.go
package main

import (
	"fmt"
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

	//  Initialize Workers
	accountWorker := mq.NewAccountWorker(accountRepo)
	transactionWorker := mq.NewTransactionWorker(accountRepo, transactionRepo)

	//  Consume messages
	accountMessages, err := mqChan.Consume(
		cfg.RabbitMQ.AccountQueue, //  Queue name
		"",                       //  Consumer tag
		false,                    //  Auto-ack
		false,                    //  Exclusive
		false,                    //  No-local
		false,                    //  No-wait
		nil,                      //  Arguments
	)
	if err != nil {
		log.Fatalf("Failed to consume from account queue: %v", err)
	}

	transactionMessages, err := mqChan.Consume(
		cfg.RabbitMQ.TransactionQueue, //  Queue name
		"",                           //  Consumer tag
		false,                        //  Auto-ack
		false,                        //  Exclusive
		false,                        //  No-local
		false,                        //  No-wait
		nil,                          //  Arguments
	)
	if err != nil {
		log.Fatalf("Failed to consume from transaction queue: %v", err)
	}

	//  Start workers
	go accountWorker.ProcessMessages(accountMessages)
	go transactionWorker.ProcessMessages(transactionMessages)

	//  Keep the worker running
	forever := make(chan bool)
	fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}