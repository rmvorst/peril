package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gamestate *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(playState routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gamestate.HandlePause(playState)
		return pubsub.Ack
	}
}

func handlerMove(gamestate *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome := gamestate.HandleMove(move)

		switch outcome {
		case gamelogic.MoveOutcomeMakeWar:
			routingKey := routing.WarRecognitionsPrefix + "." + gamestate.GetUsername()
			payload := gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gamestate.GetPlayerSnap(),
			}

			err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routingKey, payload)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack

		default:
			return pubsub.Ack
		}
	}
}

func handlerWar(gamestate *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(recognitionOfWar gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome, _, _ := gamestate.HandleWar(recognitionOfWar)

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard

		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.Ack

		case gamelogic.WarOutcomeYouWon:
			return pubsub.Ack

		case gamelogic.WarOutcomeDraw:
			return pubsub.Ack

		default:
			log.Println("Error: Unexpected war outcome.")
			return pubsub.NackDiscard
		}
	}
}
