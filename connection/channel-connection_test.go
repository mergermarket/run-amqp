package connection

import (
	"github.com/mergermarket/run-amqp/helpers"
	amqp "github.com/rabbitmq/amqp091-go"
	"testing"
	"time"
)

func TestChannelConnection_OpenChannel(t *testing.T) {
	logger := helpers.NewTestLogger(t)
	t.Run("should re-open a channel after an error has occured", func(t *testing.T) {
		server := newServerConnection(testRabbitURI, logger)

		connections := server.GetConnections()

		channelConnection := newChannelConnection(logger, "testing re-openning of channel due to an error")

		select {
		case conn := <-connections:
			channelConnection.OpenChannel(conn)
		case <-time.After(2 * time.Second):
			t.Fatalf("failed to connect to RabbitMQ at URL: %s in %d ms", testRabbitURI, 2)

		}

		select {
		case <-channelConnection.NewChannel():
			t.Log("i got a new channel.")
		case <-time.After(2 * time.Second):
			t.Fatalf("failed to connect to open a channel in time %d ms", 2)

		}

		channelConnection.sendError(amqp.ErrClosed)

		select {
		case <-channelConnection.NewChannel():
			t.Log("i got a new channel.")
		case <-time.After(2 * time.Second):
			t.Fatalf("failed to connect to open a channel in time %d ms", 2)

		}
	})
}
