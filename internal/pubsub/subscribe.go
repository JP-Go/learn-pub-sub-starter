package pubsub

import (
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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}
	deliveryChan, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	unmarshaller := func(data []byte) (T, error) {
		var message T
		err := json.Unmarshal(data, &message)
		return message, err
	}
	go func() {
		for delivery := range deliveryChan {
			message, err := unmarshaller(delivery.Body)
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
