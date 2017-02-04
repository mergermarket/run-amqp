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
	OpenChannel(description string) chan *amqp.Channel
}
type manager struct {
	openConnection     *amqp.Connection
	connections        chan *amqp.Connection
	logger             logger
	channelConnections []channelConnection
}

func NewConnectionManager(URL string, logger logger) ConnectionManager {

	server := newServerConnection(URL, logger)

	newManager := manager{
		connections:        server.GetConnections(),
		logger:             logger,
		channelConnections: make([]channelConnection, 0),
	}

	go newManager.listenForNewOpenConnections()

	return &newManager
}

func (m *manager) listenForNewOpenConnections() {
	for conn := range m.connections {
		m.openConnection = conn
		for _, channelConnection := range m.channelConnections {
			channelConnection.OpenChannel(conn)
		}
	}
}

func (m *manager) OpenChannel(description string) chan *amqp.Channel {

	channelConnection := newChannelConnection(m.logger, description)
	m.channelConnections = append(m.channelConnections, channelConnection)

	go func() {
		if m.openConnection != nil {
			channelConnection.OpenChannel(m.openConnection)
		}
	}()

	return channelConnection.NewChannel()
}
