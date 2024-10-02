package connection

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type logger interface {
	Info(...interface{})
	Error(...interface{})
	Debug(...interface{})
}

// Disable the linter because it complains about the name ConnectionManager, which I don't want to change right now
//
//revive:disable
type ConnectionManager interface {
	OpenChannel(description string) chan *amqp.Channel
	sendConnectionError(err *amqp.Error)
	sendChannelError(index uint8, err *amqp.Error) error
}

//revive:enable

type manager struct {
	openConnection     *amqp.Connection
	connections        chan *amqp.Connection
	logger             logger
	channelConnections []channelConnection
	server             serverConnection
}

func NewConnectionManager(URL string, logger logger) ConnectionManager {

	server := newServerConnection(URL, logger)

	newManager := manager{
		connections:        server.GetConnections(),
		logger:             logger,
		channelConnections: make([]channelConnection, 0),
		server:             server,
	}

	go newManager.listenForNewOpenConnections()

	return &newManager
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

func (m *manager) listenForNewOpenConnections() {
	for conn := range m.connections {
		m.openConnection = conn
		for _, channelConnection := range m.channelConnections {
			channelConnection.OpenChannel(conn)
		}
	}
}

func (m *manager) sendConnectionError(err *amqp.Error) {
	m.server.sendError(err)
}

func (m *manager) sendChannelError(index uint8, err *amqp.Error) error {

	if int(index) >= len(m.channelConnections) {
		return fmt.Errorf("index %d is out of range of length %d", index, len(m.channelConnections))
	}
	m.channelConnections[index].sendError(err)

	return nil
}
