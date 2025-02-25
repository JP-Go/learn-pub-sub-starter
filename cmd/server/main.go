package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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
	defer conn.Close()
	fmt.Println("Successfully connected to rabbitmq broker.")
	gamelogic.PrintServerHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "pause":
			fmt.Println("Game is paused.")
			pubsub.PublishJSON(rabbitChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				})
			break
		case "resume":
			fmt.Println("Resuming game.")
			pubsub.PublishJSON(rabbitChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				})
		case "quit":
			fmt.Println("Quitting game. Goodbye")
			break
		default:
			fmt.Println("Could not understand command '" + input[0] + "'")
		}
	}
}
