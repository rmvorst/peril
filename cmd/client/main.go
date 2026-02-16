package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const connString = "amqp://guest:guest@localhost:5672/"

func logErr(err error) {
	if err != nil {
		log.Printf("Error: %s\n", err)
	}
}

func fatalErr(err error) {
	if err != nil {
		log.Fatalf("Error: %s\n", err)
	}
}

func main() {
	conn, err := amqp.Dial(connString)
	fatalErr(err)
	defer conn.Close()
	fmt.Println("Connection to Rabbitmq successful")

	ch, err := conn.Channel()
	fatalErr(err)

	username, err := gamelogic.ClientWelcome()
	fatalErr(err)

	// Initialize new game state for the user
	gamestate := gamelogic.NewGameState(username)

	// Subscribe to paused game status
	pauseQueueName := routing.PauseKey + "." + username
	pauseKey := routing.PauseKey
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, pauseQueueName, pauseKey, 1, handlerPause(gamestate))
	logErr(err)

	// Subscribe to move game status
	moveQueueName := routing.ArmyMovesPrefix + "." + username
	moveKey := routing.ArmyMovesPrefix + ".*"
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, moveQueueName, moveKey, 1, handlerMove(gamestate))
	logErr(err)

	for {
		words := gamelogic.GetInput()

		if words[0] == "spawn" {
			err = gamestate.CommandSpawn(words)
			if err != nil {
				log.Printf("Error Spawning Units: %s\n", err)
				continue
			}
			log.Println("spawning units")

		} else if words[0] == "move" {
			armyMove, err := gamestate.CommandMove(words)
			if err != nil {
				log.Printf("Error Moving Units: %s\n", err)
				continue
			}
			pubsub.PublishJSON(ch, routing.ExchangePerilTopic, moveQueueName, armyMove)

		} else if words[0] == "status" {
			gamestate.CommandStatus()
		} else if words[0] == "help" {
			gamelogic.PrintClientHelp()
		} else if words[0] == "spam" {
			log.Println("Spamming not allowed yet!")
		} else if words[0] == "quit" {
			gamelogic.PrintQuit()
			break
		} else {
			log.Printf("Error: %s not a valid command.\n", words[0])
		}
	}
	pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("\nConnection closing")
}
