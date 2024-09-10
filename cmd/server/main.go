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

	rCh, err := conn.Channel()
	if err != nil {
		log.Print(err)
	}

	gamelogic.PrintServerHelp()

	key := fmt.Sprintf("%s.*", routing.GameLogSlug)

	if err := pubsub.SubscribeGob(conn, routing.ExchangePerilTopic, routing.GameLogSlug, key, pubsub.Durable, hadleGameLogs()); err != nil {
		log.Fatal(err)
	}

	var quit bool

	for !quit {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			log.Println("Sending pause message")
			if err := pubsub.PublishJSON(rCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true}); err != nil {
				log.Print(err)
			}
		case "resume":
			log.Println("Sending resume message")
			if err := pubsub.PublishJSON(rCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false}); err != nil {
				log.Print(err)
			}
		case "quit":
			log.Println("Exiting...")
			quit = true
		default:
			log.Println("Unknown command")
		}
	}

	fmt.Println("Server is shutting down...")
}

func hadleGameLogs() func(l routing.GameLog) pubsub.AckType {
	return func(l routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")

		gamelogic.WriteLog(l)

		return pubsub.Ack
	}
}
