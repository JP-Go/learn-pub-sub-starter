package main

import (
	"fmt"
	"log"
	"os"

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
	rabbitChan, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		fmt.Sprintf("%s.*", routing.GameLogSlug),
		pubsub.QueueTypeDurable,
	)
	if err != nil {
		log.Fatalf("Could not connect to game logs queue. Exiting. Error: %s", err)
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
		case "help":
			gamelogic.PrintServerHelp()
			break
		case "quit":
			fmt.Println("Quitting game. Goodbye")
			os.Exit(0)
			break
		default:
			fmt.Println("Could not understand command '" + input[0] + "'")
		}
	}
}
