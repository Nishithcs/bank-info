//  internal/mq/publisher.go
package mq

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
)

type Publisher struct {
	channel *amqp.Channel
}

func NewPublisher(ch *amqp.Channel) *Publisher {
	return &Publisher{channel: ch}
}

func (p *Publisher) Publish(queueName string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = p.channel.Publish(
		"",        //  Exchange
		queueName, //  Routing key
		false,     //  Mandatory
		false,     //  Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	fmt.Printf(" [x] Sent %s\n", body)
	return nil
}