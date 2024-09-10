package main

import (
	"fmt"
	"log"
	"time"

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

	pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, namePause, routing.PauseKey, pubsub.Transient, handlerPause(state))
	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+".#", pubsub.Durable, handlerWar(state, rCh))
	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, nameMove, keyMove, pubsub.Transient, handlerMove(state, rCh))

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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)

		return pubsub.Ack
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(wr gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		outcome, winner, loser := gs.HandleWar(wr)

		key := fmt.Sprintf("%s.%s", routing.GameLogSlug, gs.GetUsername())

		msg := ""

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			msg = fmt.Sprintf("%s won a war against %s", winner, loser)
		case gamelogic.WarOutcomeYouWon:
			msg = fmt.Sprintf("%s won a war against %s", winner, loser)
		case gamelogic.WarOutcomeDraw:
			msg = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
		default:
			log.Printf("Unknown outcome id: %d", outcome)
			return pubsub.NackDiscard
		}

		gameLog := routing.GameLog{
			CurrentTime: time.Now(),
			Message:     msg,
			Username:    gs.GetUsername(),
		}

		if err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, key, gameLog); err != nil {
			log.Print(err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(am gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")

		mo := gs.HandleMove(am)

		switch mo {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:

			key := fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.GetUsername())

			row := gamelogic.RecognitionOfWar{Attacker: am.Player, Defender: gs.Player}

			if err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, key, row); err != nil {
				log.Print(err)
				return pubsub.NackRequeue
			}

			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}
