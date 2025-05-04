package queue

import (
	"encoding/json"
	"log"
	"github.com/streadway/amqp"
	"github.com/Nishithcs/bank-info/internal/service"
)

func StartConsumer() {
	InitRabbit()
	msgs, _ := ch.Consume("transaction", "", true, false, false, false, nil)
	go func() {
		for d := range msgs {
			var task service.TransactionTask
			if err := json.Unmarshal(d.Body, &task); err == nil {
				service.ProcessTransaction(task)
			} else {
				log.Println("Unmarshal error:", err)
			}
		}
	}()
}