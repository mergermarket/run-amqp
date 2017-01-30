package connection

import (
	"github.com/streadway/amqp"
)

type logger interface {
	Info(...interface{})
	Error(...interface{})
	Debug(...interface{})
}

type ConnectionManager interface {
	OpenChannels() chan *amqp.Channel
}
type manager struct {
	openConnection *amqp.Connection
	connections    chan *amqp.Connection
	logger         logger
}

func NewConnectionManager(URL string, logger logger) ConnectionManager {

	server := newServerConnection(URL, logger)

	newManager := manager{
		connections: server.GetConnections(),
		logger:      logger,
	}

	return &newManager
}

func (m *manager) OpenChannels() chan *amqp.Channel {

	channelConnection := newChannelConnection(m.logger)

	go func() {
		for conn := range m.connections {
			m.openConnection = conn
			channelConnection.OpenChannel(conn)

		}
	}()

	return channelConnection.NewChannels()
}
