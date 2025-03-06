package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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
	defer conn.Close()
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}
	StartGame(conn, username)
}
