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
	OpenChannels(count uint8) []chan *amqp.Channel
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

func (m *manager) OpenChannels(count uint8) []chan *amqp.Channel {
	newChannels := make([]channelConnection, 0)

	channels := make([]chan *amqp.Channel, 0)

	for i := count; i > 0; i-- {
		c := newChannelConnection(m.logger)
		newChannels = append(newChannels, c)
		channels = append(channels, c.NewChannels())
	}

	go func() {
		for conn := range m.connections {
			m.openConnection = conn
			for _, c := range newChannels {
				go func(cc channelConnection) {
					cc.OpenChannel(m.openConnection)
				}(c)

			}

		}
	}()

	return channels
}
