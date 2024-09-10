package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType string

const (
	Ack         AckType = "ack"
	NackRequeue AckType = "nack_requeue"
	NackDiscard AckType = "nack_discard"
)

type SimpleQueueType = int

const (
	Durable SimpleQueueType = iota
	Transient
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {

	f := func(b []byte) (T, error) {
		var v T

		err := json.Unmarshal(b, &v)

		return v, err
	}

	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, f)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	f := func(b []byte) (T, error) {
		var v T

		err := gob.NewDecoder(bytes.NewReader(b)).Decode(&v)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return v, err
			}
		}

		return v, nil
	}

	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, f)
}
