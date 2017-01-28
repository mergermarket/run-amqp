package connection

import (
	"github.com/streadway/amqp"
	"fmt"
	"time"
)

type channelConnection interface {
	OpenChannelOn(connection *amqp.Connection)
	NewChannels() chan *amqp.Channel
}

type cConnection struct {
	channels      chan *amqp.Channel
	connection    *amqp.Connection
	openChannel   *amqp.Channel
	channelErrors chan *amqp.Error
	logger        logger
}

func newChannelConnection(logger logger) channelConnection {
	channel := cConnection{
		logger:    logger,
		channels:  make(chan *amqp.Channel),
	}

	return &channel
}

func (c *cConnection) OpenChannelOn(connection *amqp.Connection) {
	c.connection = connection
	c.create()
}

func (c *cConnection) NewChannels() chan *amqp.Channel {
	return c.channels
}

func (c *cConnection) create() {
	c.logger.Debug("openning a new channel")

	openChannel, err := c.connection.Channel()

	if err != nil {
		c.logger.Error("failed to open a new channel", err)
		return
	}

	c.logger.Debug("openned a new channel")
	c.openChannel = openChannel
	c.listenForChannelError()
	c.channels <- openChannel
}

func (c *cConnection) listenForChannelError() {
	close(c.channelErrors)
	c.channelErrors = make(chan *amqp.Error)
	c.openChannel.NotifyClose(c.channelErrors)

	go func() {
		for {
			select {
			case err, ok := <-c.channelErrors:
				if err != nil && ok {
					c.logger.Error(fmt.Sprintf("there was channel/sConnection error with Code: %d Reason: $s", err.Code, err.Reason))
					c.closeOpenChannel()
					c.create()
				}
			default:
				c.logger.Debug("will resume listening for channel errors in 3 seconds")
				time.Sleep(3 * time.Second)

			}

		}
	}()
}
func (c *cConnection) closeOpenChannel() {
	err := c.openChannel.Close()

	if err != nil {
		c.logger.Error("failed to close the channel", err)
	}
}
