package connection

import (
	"github.com/streadway/amqp"
)

type ConnectionManager interface {
	OpenChannels() chan *amqp.Channel
}
type manager struct {
	openConnection *amqp.Connection
	connections chan *amqp.Connection
	logger         logger
}

func NewConnectionManager(URL string, logger logger) ConnectionManager {

	server := newServerConnection(URL, logger)

	newManager := manager{
		connections:server.GetConnections(),
		logger:logger,
	}

	return &newManager
}

func (m *manager) OpenChannel() chan *amqp.Channel {
	c := newChannelConnection(m.logger)

	go func() {
		for conn := range m.connections {
			m.openConnection = conn
			c.OpenChannelOn(m.openConnection)
		}
	}()

	return c.NewChannels()
}

