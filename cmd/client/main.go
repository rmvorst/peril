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
		log.Fatalf("Error: %s\n", err)
	}
}

func main() {
	conn, err := amqp.Dial(connString)
	logErr(err)
	defer conn.Close()
	fmt.Println("Connection to Rabbitmq successful")

	ch, err := conn.Channel()
	logErr(err)

	username, err := gamelogic.ClientWelcome()
	logErr(err)

	queueName := routing.PauseKey + "." + username
	pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, 1)
	gamestate := gamelogic.NewGameState(username)

	for {
		words := gamelogic.GetInput()

		if words[0] == "spawn" {
			log.Println("spawning units")
			err = gamestate.CommandSpawn(words)
			if err != nil {
				log.Fatalf("Error: %s\n", err)
			}
		} else if words[0] == "move" {
			log.Println("moving units")
			_, err := gamestate.CommandMove(words)
			if err != nil {
				log.Fatalf("Error: %s\n", err)
			}
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
