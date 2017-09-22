package connection

import (
	"fmt"
	"github.com/streadway/amqp"
	"math"
	"strings"
	"time"
)

type serverConnection interface {
	GetConnections() chan *amqp.Connection
	GetConnectionStatus() chan bool
	sendError(err *amqp.Error)
}

type sConnection struct {
	logger              logger
	URL                 string
	openConnection      *amqp.Connection
	connections         chan *amqp.Connection
	isConnectionBlocked chan bool
	errors              chan *amqp.Error
	blockings           chan amqp.Blocking
}

func newServerConnection(URL string, logger logger) serverConnection {
	newConnection := sConnection{
		URL:                 URL,
		logger:              logger,
		connections:         make(chan *amqp.Connection),
		isConnectionBlocked: make(chan bool),
		errors:              make(chan *amqp.Error),
		blockings:           make(chan amqp.Blocking),
	}

	go newConnection.connect()

	return &newConnection
}

func (c *sConnection) GetConnections() chan *amqp.Connection {
	return c.connections
}

func (c *sConnection) GetConnectionStatus() chan bool {
	return c.isConnectionBlocked
}

func (c *sConnection) sendError(err *amqp.Error) {
	c.errors <- err
}

const takeHeartbeatFromServer = 900 * time.Millisecond // less than 1s uses the server's interval

func (c *sConnection) connect() {
	attempts := 0
	for {
		c.logger.Info("Connecting to", c.URL)
		attempts++
		openConnection, err := amqp.DialConfig(c.URL, amqp.Config{
			Heartbeat: takeHeartbeatFromServer,
		})

		if err != nil {
			c.logger.Error(fmt.Errorf("problem connecting to %s, %v", c.URL, err))
			millis := math.Exp2(float64(attempts))
			sleepDuration := time.Duration(int(millis)) * time.Second
			c.logger.Info("Trying to reconnect to RabbitMQ at", c.URL, "after", sleepDuration)
			time.Sleep(sleepDuration)
			continue
		}

		c.logger.Info("Connected to", c.URL)

		c.listenForConnectionError()
		c.listenForConnectionBlocked()
		c.openConnection = openConnection
		go func() {
			c.connections <- openConnection
		}()

		return

	}
}

func (c *sConnection) listenForConnectionError() {

	go func() {

		c.errors = c.openConnection.NotifyClose(c.errors)

		for {
			err, ok := <-c.errors
			if err != nil && ok {
				c.logger.Error(fmt.Sprintf(`there was a connection error with Code: "%d" Reason: "%s" - will try to re-connect now.`, err.Code, err.Reason))
				c.closeOpenConnection()
				c.connect()
			}
		}
	}()
}

func (c *sConnection) listenForConnectionBlocked() {

	go func() {

		c.blockings = c.openConnection.NotifyBlocked(c.blockings)

		for {
			if blocking, ok := <-c.blockings; ok {
				c.logger.Info(fmt.Sprintf("connection blocking received with TCP %t ready, with reason: %s", blocking.Active, blocking.Reason))
				c.isConnectionBlocked <- blocking.Active
			}
		}
	}()
}

func (c *sConnection) closeOpenConnection() {

	err := c.openConnection.Close()
	if err != nil {
		if strings.Contains(err.Error(), "channel/connection is not open") {
			c.logger.Info("could not close connection because it's no longer open", err)
		} else {
			c.logger.Error("failed to close the connection", err)
		}
	}
}
