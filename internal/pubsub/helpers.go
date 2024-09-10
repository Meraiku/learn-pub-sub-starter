package pubsub

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
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

			v, err := unmarshaller(msg.Body)
			if err != nil {
				log.Print(err)
				continue
			}
			t := handler(v)

			switch t {
			case Ack:
				if err := msg.Ack(false); err != nil {
					log.Print(err)
					continue
				}
			case NackRequeue:
				if err := msg.Nack(false, true); err != nil {
					log.Print(err)
					continue
				}
			case NackDiscard:
				if err := msg.Nack(false, false); err != nil {
					log.Print(err)
					continue
				}
			default:
				log.Printf("Unknown type! Want 'AckType', got %T", t)
			}

		}
	}()

	return nil
}
