package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

	connString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connString)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected to rabbitmq broker.")
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.QueueTypeTransient)
	if err != nil {
		log.Fatal(err)
	}
	gameState := gamelogic.NewGameState(username)
outer:
	for {
		input := gamelogic.GetInput()
		switch input[0] {
		case "spawn":
			err := gameState.CommandSpawn(input)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
			}
			break
		case "move":
			move, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
			} else {
				fmt.Printf("Player %s moved troops to %s\n", move.Player.Username, move.ToLocation)
			}
			break
		case "status":
			gameState.CommandStatus()
			break
		case "help":
			gamelogic.PrintClientHelp()
			break
		case "spam":
			fmt.Println("Spamming is not allowed yet!")
			break
		case "quit":
			gamelogic.PrintQuit()
			break outer
		default:
			fmt.Printf("ERROR: could not understand command %s\n", input[0])
		}
	}
}
