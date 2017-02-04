package connection

import (
	"fmt"
	"github.com/streadway/amqp"
)

type channelConnection interface {
	OpenChannel(connection *amqp.Connection)
	NewChannel() chan *amqp.Channel
	getErrors() chan *amqp.Error
}

type cConnection struct {
	channels           chan *amqp.Channel
	connection         *amqp.Connection
	openChannel        *amqp.Channel
	errors             chan *amqp.Error
	logger             logger
	channelDescription string
}

func newChannelConnection(logger logger, channelDescription string) channelConnection {
	channel := cConnection{
		logger:             logger,
		channels:           make(chan *amqp.Channel),
		channelDescription: channelDescription,
		errors:             make(chan *amqp.Error),
	}

	return &channel
}

func (c *cConnection) OpenChannel(connection *amqp.Connection) {
	c.connection = connection
	go c.create()
}

func (c *cConnection) NewChannel() chan *amqp.Channel {
	return c.channels
}

func (c *cConnection) getErrors() chan *amqp.Error {
	return c.errors
}

func (c *cConnection) create() {
	c.logger.Debug(fmt.Sprintf(`openning a new channel for "%s"`, c.channelDescription))

	openChannel, err := c.connection.Channel()

	if err != nil {
		c.logger.Error(fmt.Sprintf(`failed to open a new channel for "%s"`, c.channelDescription), err)
		return
	}

	c.logger.Debug(fmt.Sprintf(`successfully opened a new channel for "%s"`, c.channelDescription))
	c.openChannel = openChannel
	c.listenForChannelError()
	func() {
		c.channels <- openChannel
	}()
}

func (c *cConnection) listenForChannelError() {

	go func() {

		c.errors = c.openChannel.NotifyClose(c.errors)

		for {
			err, ok := <-c.errors
			if err != nil && ok {
				c.logger.Error(fmt.Sprintf(`something bad happend with channel opened for "%s" with error code : "%d" reason: "%s" - will try to re-open channel`, c.channelDescription, err.Code, err.Reason))
				c.closeOpenChannel()
				c.create()
			}
		}
	}()
}
func (c *cConnection) closeOpenChannel() {
	err := c.openChannel.Close()

	if err != nil {
		c.logger.Error(fmt.Sprintf(`failed to close the channel opened for "%s"`, c.channelDescription), err)
	}
}
