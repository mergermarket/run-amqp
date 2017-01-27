package runamqp

import (
	"github.com/streadway/amqp"
	"time"
	"fmt"
	"math"
)

type connection struct {
	connectionConfig    *connectionConfig
	connections         chan *amqp.Connection
	openConnection      *amqp.Connection
	isConnectionBlocked chan bool
	connectionErrors    chan *amqp.Error
	connectionBlocking  chan amqp.Blocking
}

func NewConnection(connectionConfig *connectionConfig) (connections chan *amqp.Connection, isConnectionBlocked chan bool) {
	connection := connection{
		connectionConfig:    connectionConfig,
		connections:         make(chan *amqp.Connection),
		isConnectionBlocked: make(chan bool),
	}

	go connection.connect()

	return connection.connections, connection.isConnectionBlocked
}

func (c *connection) connect() {
	attempts := 0
	for {
		c.connectionConfig.Logger.Info("Connecting to", c.connectionConfig.URL)
		attempts++
		openConnection, err := amqp.DialConfig(c.connectionConfig.URL, amqp.Config{
			Heartbeat: 30 * time.Second,
		})

		if err == nil {
			c.connectionConfig.Logger.Info("Connected to", c.connectionConfig.URL)

			c.listenForConnectionError()
			c.listenForConnectionBlocked()
			c.openConnection = openConnection
			c.connections <- openConnection

		}

		c.connectionConfig.Logger.Error(fmt.Errorf("problem connecting to %s, %v", c.connectionConfig.URL, err))
		millis := math.Exp2(float64(attempts))
		sleepDuration := time.Duration(int(millis)) * time.Second
		c.connectionConfig.Logger.Info("Trying to reconnect to RabbitMQ at", c.connectionConfig.URL, "after", sleepDuration)
		time.Sleep(sleepDuration)
	}
}

func (c *connection) listenForConnectionError() {
	close(c.connectionErrors)
	c.connectionErrors = make(chan *amqp.Error)
	c.openConnection.NotifyClose(c.connectionErrors)

	go func() {
		for {
			if err, ok := <-c.connectionErrors; err != nil && ok {
				c.connectionConfig.Logger.Error(fmt.Sprintf("there was connection error with Code: %d Reason: $s", err.Code, err.Reason))
				c.closeOpenConnection()
				c.connect()
			}
		}
	}()
}

func (c *connection) listenForConnectionBlocked() {

	close(c.connectionBlocking)
	c.connectionBlocking = make(chan amqp.Blocking)
	c.openConnection.NotifyBlocked(c.connectionBlocking)

	go func() {
		for {
			if blocking, ok := <-c.connectionBlocking; ok {
				message := fmt.Sprintf("connection blocking received with TCP %t ready, with reason: %s", blocking.Active, blocking.Reason)
				c.connectionConfig.Logger.Info(message)
				c.isConnectionBlocked <- blocking.Active
			}
		}
	}()
}

func (c *connection) closeOpenConnection()  {
	err:= c.openConnection.Close()

	if err != nil {
		c.connectionConfig.Logger.Error("failed to close the connection", err)
	}
}