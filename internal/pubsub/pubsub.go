package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueTypeTransient = iota
	QueueTypeDurable
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	body, err := json.Marshal(val)
	if err != nil {
		return err
	}
	ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	return nil
}

func DeclareAndBind(conn *amqp.Connection,
	exchange, queueName, key string, simpleQueueType int) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	var queue amqp.Queue
	if simpleQueueType == QueueTypeDurable {
		queue, err = channel.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			return nil, amqp.Queue{}, err
		}
	} else {
		queue, err = channel.QueueDeclare(queueName, false, true, true, false, nil)
		if err != nil {
			return nil, amqp.Queue{}, err
		}
	}
	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return channel, queue, nil

}
