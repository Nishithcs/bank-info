package queue

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
	"os"
)

var conn *amqp.Connection
var ch *amqp.Channel

func InitRabbit() {
	uri := os.Getenv("RABBITMQ_URI")
	var err error
	conn, err = amqp.Dial(uri)
	if err != nil {
		log.Fatal(err)
	}
	ch, err = conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	ch.QueueDeclare("transaction", true, false, false, false, nil)
}

func Publish(msg interface{}) {
	body, _ := json.Marshal(msg)
	err := ch.Publish("", "transaction", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		log.Println("Publish error:", err)
	}
}