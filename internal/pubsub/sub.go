package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T),
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}

	deliveryCh, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range deliveryCh {
			var v T

			if err := json.Unmarshal(msg.Body, &v); err != nil {
				log.Print(err)
				continue
			}
			handler(v)

			if err := msg.Ack(false); err != nil {
				log.Print(err)
				continue
			}
		}
	}()

	return nil
}
