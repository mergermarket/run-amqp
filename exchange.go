package runamqp

import (
	"fmt"
	"github.com/streadway/amqp"
	"strings"
)

type ExchangeType string

const (
	Fanout       ExchangeType = "fanout"
	Topic        ExchangeType = "topic"
	Direct       ExchangeType = "direct"
	Headers      ExchangeType = "headers"
	Unrecognised ExchangeType = "unrecognised"
)

func NewExchangeType(typ string) (ExchangeType, error) {
	lowercaseType := strings.ToLower(typ)

	switch lowercaseType {
	case "topic":
		return Topic, nil
	case "fanout":
		return Fanout, nil
	case "direct":
		return Direct, nil
	case "headers":
		return Headers, nil
	default:
		return Unrecognised, fmt.Errorf("Unrecognised exchange type %s", typ)
	}
}

const (
	durable    = true
	autoDelete = false
	internal   = false
	nowait     = false
)

func makeExchange(ch *amqp.Channel, exchangeName string, exchangeType ExchangeType) error {

	if exchangeType == Unrecognised {
		return fmt.Errorf("Unrecognised exchange type, check config")
	}

	return ch.ExchangeDeclare(
		exchangeName,
		string(exchangeType),
		durable,
		autoDelete,
		internal,
		nowait,
		nil,
	)
}
