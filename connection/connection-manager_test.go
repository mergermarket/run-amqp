package connection

import (
	"github.com/mergermarket/run-amqp/helpers"
	"testing"
)

func xTestNewConnectionManager_OpenChannel(t *testing.T) {
	logger := helpers.NewTestLogger(t)
	t.Run("should re-open connections and all the associated channels", func(t *testing.T) {

		server := newServerConnection(testRabbitURI, logger)

		newManager := manager{
			connections:        server.GetConnections(),
			logger:             logger,
			channelConnections: make([]channelConnection, 0),
		}

		<-newManager.channelConnections[0].NewChannel()
	})
}
