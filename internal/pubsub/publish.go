package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	gobPublisher  = NewPublisher("application/gob", gobEncoder[any])
	jsonPublisher = NewPublisher("application/json", jsonEncoder[any])
)

func gobEncoder[T any](data T) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func jsonEncoder[T any](data T) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	return jsonPublisher(ch, val, exchange, key)
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	return gobPublisher(ch, val, exchange, key)
}

func NewPublisher[T any](contentType string, encoder func(T) ([]byte, error)) func(ch *amqp.Channel, val T, exchange, key string) error {
	return func(ch *amqp.Channel, val T, exchange, key string) error {
		data, err := encoder(val)
		if err != nil {
			return err
		}
		return publish(ch, exchange, key, contentType, data)
	}
}

func publish(ch *amqp.Channel, exchange, key, contentType string, data []byte) error {
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: contentType,
		Body:        data,
	})

}
