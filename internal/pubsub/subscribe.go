package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype = int

const (
	AcktypeAck Acktype = iota
	AcktypeNackRequeue
	AcktypeNackDiscard
)

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
	unmarshaler func([]byte) (T, error),
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}
	deliveryChan, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for delivery := range deliveryChan {
			message, err := unmarshaler(delivery.Body)
			if err != nil {
				log.Printf("Error reading message from queue: %v", err)
				continue
			}
			ack := handler(message)
			switch ack {
			case AcktypeAck:
				delivery.Ack(false)
			case AcktypeNackRequeue:
				delivery.Nack(false, true)
			case AcktypeNackDiscard:
				delivery.Nack(false, false)
			default:
				delivery.Nack(false, false)
				log.Printf("Invalid acknowledge type %+v. Discarding", ack)
			}

		}
	}()

	return nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	unmarshaler := func(data []byte) (T, error) {
		var message T
		err := json.Unmarshal(data, &message)
		return message, err
	}
	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, unmarshaler)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	unmarshaler := func(data []byte) (T, error) {
		dec := gob.NewDecoder(bytes.NewBuffer(data))
		var message T
		err := dec.Decode(&message)
		return message, err
	}
	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, unmarshaler)
}
