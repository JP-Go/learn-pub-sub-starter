package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connString)
	if err != nil {
		panic(err)
	}
	rabbitChan, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	pubsub.PublishJSON(rabbitChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
		IsPaused: true,
	})
	defer conn.Close()
	fmt.Println("Successfully connected to rabbitmq broker.")
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	<-signalChannel
	fmt.Println("Shutting down. Goodbye")
}
