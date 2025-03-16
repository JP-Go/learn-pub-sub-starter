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
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.QueueTypeDurable,
		func(gl routing.GameLog) pubsub.Acktype {
			defer fmt.Print("> ")
			err := gamelogic.WriteLog(gl)
			if err != nil {
				return pubsub.AcktypeNackRequeue
			}
			return pubsub.AcktypeAck
		},
	)

	if err != nil {
		log.Fatalf("Could not connect to game logs queue. Exiting. Error: %s", err)
	}
	pubChan, err := conn.Channel()
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
			pubsub.PublishJSON(pubChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				})
			break
		case "resume":
			fmt.Println("Resuming game.")
			pubsub.PublishJSON(pubChan,
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
