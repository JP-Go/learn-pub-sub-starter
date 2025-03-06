package pubsub

import (
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType = int

const (
	QueueTypeTransient SimpleQueueType = iota
	QueueTypeDurable
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange, queueName, key string,
	simpleQueueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	table := amqp.Table{
		"x-dead-letter-exchange": routing.ExchangePerilDlx,
	}
	var queue amqp.Queue
	if simpleQueueType == QueueTypeDurable {
		queue, err = channel.QueueDeclare(queueName, true, false, false, false, table)
		if err != nil {
			return nil, amqp.Queue{}, err
		}
	} else {
		queue, err = channel.QueueDeclare(queueName, false, true, true, false, table)
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
