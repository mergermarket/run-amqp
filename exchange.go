package runamqp

import (
	"fmt"
	"github.com/streadway/amqp"
	"strings"
)

// ExchangeType corresponds to the type of exchanges available in rabbit mq
type ExchangeType string

const (
	// Fanout should have an explanation from Baktash
	Fanout ExchangeType = "fanout"

	// Topic should have an explanation from Baktash
	Topic ExchangeType = "topic"

	// Direct should have an explanation from Baktash
	Direct ExchangeType = "direct"

	// Unrecognised is a catch all for exchanges that arent supported
	Unrecognised ExchangeType = "unrecognised"
)

// NewExchangeType returns an Exchange type for a given string if it is a valid
func NewExchangeType(typ string) (ExchangeType, error) {
	lowercaseType := strings.ToLower(typ)

	switch lowercaseType {
	case "topic":
		return Topic, nil
	case "fanout":
		return Fanout, nil
	case "direct":
		return Direct, nil
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

//
//func makeExchangePassive(ch *amqp.Channel, exchangeName string, exchangeType ExchangeType) error {
//
//	if exchangeType == Unrecognised {
//		return fmt.Errorf("Unrecognised exchange type, check config")
//	}
//
//	return ch.ExchangeDeclarePassive(
//		exchangeName,
//		string(exchangeType),
//		durable,
//		autoDelete,
//		internal,
//		nowait,
//		nil,
//	)
//}
