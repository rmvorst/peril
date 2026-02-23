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

func Server_ConfigureRabbitMQ(conn *amqp.Connection, ch *amqp.Channel) error {
	// Declare relevant exchagnes for Peril on RabbitMQ
	err := ch.ExchangeDeclare(routing.ExchangePerilDirect, "direct", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(routing.ExchangePerilTopic, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(routing.ExchangePerilFanout, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	queueName := routing.GameLogSlug + ".*"
	pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, queueName, 0)

	return nil
}

func main() {
	conn, err := amqp.Dial(connString)
	logErr(err)
	defer conn.Close()
	fmt.Println("Connection to Rabbitmq successful")

	ch, err := conn.Channel()
	logErr(err)

	err = Server_ConfigureRabbitMQ(conn, ch)
	logErr(err)

	gamelogic.PrintServerHelp()
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		if words[0] == "pause" {
			log.Println("Sending a pause message")
			pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
		} else if words[0] == "resume" {
			log.Println("Sending a resume message")
			pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
		} else if words[0] == "quit" {
			log.Println("Exiting")
			break
		} else {
			log.Println("Command not understood:", words[0])
		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("\nConnection closing")
}
