package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	var durable, autoDelete, exclusive, noWait bool

	switch simpleQueueType {
	case 0:
		durable = true
	case 1:
		autoDelete = true
		exclusive = true
	}

	rCh, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	queue, err := rCh.QueueDeclare(queueName, durable, autoDelete, exclusive, noWait, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	if err := rCh.QueueBind(queueName, key, exchange, noWait, nil); err != nil {
		return nil, amqp.Queue{}, err
	}

	return rCh, queue, nil
}
