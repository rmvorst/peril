package routing

import (
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ArmyMovesPrefix = "army_moves"

	WarRecognitionsPrefix = "war"

	PauseKey = "pause"

	GameLogSlug = "game_logs"
)

const (
	ExchangePerilDirect = "peril_direct"
	ExchangePerilTopic  = "peril_topic"
	ExchangePerilFanout = "peril_dlx"
)

func Server_ConfigureRabbitMQ(conn *amqp.Connection, ch *amqp.Channel) error {
	// Declare relevant exchagnes for Peril on RabbitMQ
	err := ch.ExchangeDeclare(ExchangePerilDirect, "direct", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(ExchangePerilTopic, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(ExchangePerilFanout, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	queueName := GameLogSlug + ".*"
	pubsub.DeclareAndBind(conn, ExchangePerilTopic, GameLogSlug, queueName, 0)

	return nil
}
