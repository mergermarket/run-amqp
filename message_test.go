package runamqp

import (
	"github.com/golang/mock/gomock"
	"github.com/mergermarket/run-amqp/connection"
	"github.com/streadway/amqp"
	"testing"
	"time"
)

type stubbedAcknowledger struct {
	ackCalled bool
}

func (a *stubbedAcknowledger) Ack(tag uint64, multiple bool) error {

	a.ackCalled = true
	return nil
}

func (a *stubbedAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	return nil
}

func (a *stubbedAcknowledger) Reject(tag uint64, requeue bool) error {
	return nil
}

func TestAmqpMessage_Body(t *testing.T) {
	body := []byte("test")
	message := amqpMessage{
		delivery: amqp.Delivery{Body: body},
	}

	if string(message.Body()) != string(body) {
		t.Error("got the wrong body.")
	}
}

func TestAmqpMessage_Ack(t *testing.T) {
	acknowldger := &stubbedAcknowledger{}
	message := amqpMessage{
		delivery: amqp.Delivery{Acknowledger: acknowldger},
	}

	err := message.Ack()

	if err != nil {
		t.Fatal("Failed to Ack message", err)
	}

	if !acknowldger.ackCalled {
		t.Fatal("failed to call the underline acknowldger Ack")
	}
}

func TestAmqpMessage_Nack(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	stubChannel := connection.NewMockAMQPChannel(mockCtrl)

	body := []byte("test")

	acknowldger := &stubbedAcknowledger{}
	routingKey := "routingKey"
	nackTime := time.Now()

	dleExchangeName := "dleExchange"
	message := amqpMessage{
		delivery:        amqp.Delivery{Body: body, Acknowledger: acknowldger, RoutingKey: routingKey},
		dleExchangeName: dleExchangeName,
		dleChannel:      stubChannel,
		now:             func() time.Time { return nackTime },
	}

	reason := "reason"

	headers := make(map[string]interface{})
	headers["x-dle-reason"] = reason

	payload := amqp.Publishing{
		Body:      body,
		Headers:   headers,
		Timestamp: nackTime,
	}

	stubChannel.EXPECT().Publish(dleExchangeName, routingKey, false, false, payload)

	err := message.Nack(reason)

	if err != nil {
		t.Fatal("failed to nack the message", err)
	}

	if !acknowldger.ackCalled {
		t.Fatal("failed to call the underline acknowldger Ack")
	}
}
