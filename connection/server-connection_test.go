package connection

import (
	"github.com/mergermarket/run-amqp/helpers"
	amqp "github.com/rabbitmq/amqp091-go"
	"testing"
	"time"
)

const testRabbitURI = "amqp://guest:guest@rabbitmq:5672/"

func TestSConnection_GetConnections(t *testing.T) {
	logger := helpers.NewTestLogger(t)
	t.Run("should reconnect after an error has occured", func(t *testing.T) {
		server := newServerConnection(testRabbitURI, logger)

		connections := server.GetConnections()

		select {
		case <-connections:
			t.Log("i got a new connection")
		case <-time.After(2 * time.Second):
			t.Fatalf("failed to connect to RabbitMQ at URL: %s in %d ms", testRabbitURI, 2)

		}

		server.sendError(amqp.ErrClosed)

		select {
		case <-connections:
			t.Log("i got reconnected after i was closed")
		case <-time.After(2 * time.Second):
			t.Fatalf("failed to after re-connect to RabbitMQ at URL: %s in %d ms", testRabbitURI, 2)

		}
	})
}
