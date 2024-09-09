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
	url := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("Connection to RabbitMQ was successful!")

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Print(err)
	}

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, userName)

	ch, queue, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, 1)
	if err != nil {
		log.Fatal(err)
	}

	_, _ = ch, queue

	state := gamelogic.NewGameState(userName)

	var quit bool

	for !quit {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			if err := state.CommandSpawn(words); err != nil {
				log.Print(err)
			}
		case "move":
			_, err := state.CommandMove(words)
			if err != nil {
				log.Print(err)
			}
			fmt.Println("Move succeeded!")
		case "status":
			state.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			quit = true
		default:
			fmt.Println("Unknown command!")
		}
	}

}
