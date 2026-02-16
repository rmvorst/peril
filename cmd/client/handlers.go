package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gamestate *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(playState routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gamestate.HandlePause(playState)
		return pubsub.Ack
	}
}

func handlerMove(gamestate *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome := gamestate.HandleMove(move)

		if outcome == gamelogic.MoveOutComeSafe || outcome == gamelogic.MoveOutcomeMakeWar {
			return pubsub.Ack
		}
		return pubsub.NackDiscard
	}
}
