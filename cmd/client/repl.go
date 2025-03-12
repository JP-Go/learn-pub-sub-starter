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

func StartGame(conn *amqp.Connection, username string) {
	gameState := gamelogic.NewGameState(username)
	_, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.QueueTypeTransient,
	)
	if err != nil {
		log.Fatalf("Could not start game due to %v", err)
	}

	pubChan, err := conn.Channel()
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		string(routing.ArmyMovesPrefix)+"."+gameState.GetUsername(),
		string(routing.ArmyMovesPrefix)+"."+"*",
		pubsub.QueueTypeDurable,
		handlerMove(gameState, pubChan),
	)

	if err != nil {
		log.Fatalf("Could not connect to moves queue due to %v", err)
	}
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		string(routing.WarRecognitionsPrefix),
		string(routing.WarRecognitionsPrefix)+".*",
		pubsub.QueueTypeDurable,
		handlerWar(gameState, pubChan),
	)

	if err != nil {
		log.Fatalf("Could not connect to publish channel due to %v", err)
	}

	runREPLForUser(gameState, pubChan)
}

func runREPLForUser(gameState *gamelogic.GameState, publishChan *amqp.Channel) {
	for {
		input := gamelogic.GetInput()
		if len(input) < 1 {
			break
		}
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
				break
			}
			pubsub.PublishJSON(
				publishChan,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+move.Player.Username,
				move,
			)
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
			os.Exit(0)
		default:
			fmt.Printf("ERROR: could not understand command %s\n", input[0])
		}
	}
}
