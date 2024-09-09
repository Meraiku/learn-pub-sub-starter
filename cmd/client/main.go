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

	namePause := fmt.Sprintf("%s.%s", routing.PauseKey, userName)
	nameMove := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, userName)

	keyMove := fmt.Sprintf("%s.*", routing.ArmyMovesPrefix)

	rCh, err := conn.Channel()
	if err != nil {
		log.Print(err)
	}

	state := gamelogic.NewGameState(userName)

	pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, namePause, routing.PauseKey, 1, handlerPause(state))
	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, nameMove, keyMove, 1, handlerMove(state))

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
			glMove, err := state.CommandMove(words)
			if err != nil {
				log.Print(err)
			}
			if err := pubsub.PublishJSON(rCh, routing.ExchangePerilTopic, keyMove, glMove); err != nil {
				log.Print(err)
			}
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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(am gamelogic.ArmyMove) {
		defer fmt.Print("> ")

		gs.HandleMove(am)
	}
}
