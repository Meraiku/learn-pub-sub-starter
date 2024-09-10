package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	b, err := json.Marshal(val)
	if err != nil {
		return err
	}

	ctx := context.Background()

	if err := ch.PublishWithContext(ctx,
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        b,
		},
	); err != nil {
		return err
	}

	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {

	var b bytes.Buffer

	err := gob.NewEncoder(&b).Encode(val)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}

	ctx := context.Background()

	if err := ch.PublishWithContext(ctx,
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        b.Bytes(),
		},
	); err != nil {
		return err
	}

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	var durable, autoDelete, exclusive, noWait bool

	switch simpleQueueType {
	case Durable:
		durable = true
	case Transient:
		autoDelete = true
		exclusive = true
	}

	rCh, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	queue, err := rCh.QueueDeclare(queueName, durable, autoDelete, exclusive, noWait, amqp.Table{"x-dead-letter-exchange": "peril_dlx"})
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	if err := rCh.QueueBind(queueName, key, exchange, noWait, nil); err != nil {
		return nil, amqp.Queue{}, err
	}

	return rCh, queue, nil
}
