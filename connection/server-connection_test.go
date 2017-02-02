package connection

import (
	"github.com/mergermarket/run-amqp/helpers"
	"github.com/streadway/amqp"
	"testing"
	"time"
)

const testRabbitURI = "amqp://guest:guest@rabbitmq:5672/"

func xTestSConnection_GetConnections(t *testing.T) {
	logger := helpers.NewTestLogger(t)
	server := newServerConnection(testRabbitURI, logger)
	errors := server.getErrors()

	connections := server.GetConnections()

	select {
	case <-connections:
		t.Log("i got a new connection")
	case <-time.After(2 * time.Second):
		t.Fatalf("failed to connect to RabbitMQ at URL: %s in %d", testRabbitURI, 5)

	}

	errors <- amqp.ErrClosed

	select {
	case <-connections:
		t.Log("i got reconnected after i was closed")
	case <-time.After(2 * time.Second):
		t.Fatalf("failed to after re-connect to RabbitMQ at URL: %s in %d", testRabbitURI, 5)

	}

	t.Error("hi")
}
