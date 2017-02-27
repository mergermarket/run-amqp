package runamqp

import (
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

// Message represents a delivery over an amqp channel
type Message interface {
	Ack() error
	Body() []byte
	Nack(reason string) error
	Requeue(reason string) error
}

type amqpMessage struct {
	delivery          amqp.Delivery
	dleChannel        *amqp.Channel
	retryChannel      *amqp.Channel
	retryLimit        int
	retryExchangeName string
	dleExchangeName   string
}

// Body returns the body of the AMQP message
func (m *amqpMessage) Body() []byte {
	return m.delivery.Body
}

// Ack will acknowledge the message.
func (m *amqpMessage) Ack() error {
	return m.delivery.Ack(false)
}

// nackCalls is used when you cant process a message. The "reason" will appear in the rabbit console under the message headers which is useful for debugging
func (m *amqpMessage) Nack(reason string) error {

	err := m.Ack()

	if err != nil {
		return err
	}

	headers := make(map[string]interface{})
	headers["x-dle-reason"] = reason

	payload := amqp.Publishing{
		Body:      m.Body(),
		Headers:   headers,
		Timestamp: time.Now(),
	}

	err = m.dleChannel.Publish(m.dleExchangeName, m.delivery.RoutingKey, false, false, payload)

	return err
}

// requeueCalls requeues a message, which is useful for when you have transient problems
func (m *amqpMessage) Requeue(reason string) error {

	if m.retryLimit > 0 {
		retryCount := 1
		if headerRetryCount, found := m.delivery.Headers["x-retry-count"]; found {

			if temp, ok := headerRetryCount.(int64); ok {
				retryCount = int(temp) + 1
			} else {
				return fmt.Errorf("The message %+v retry count could not be parsed correctly, this is probably a bug in run-amqp", m)
			}

		}

		if retryCount > m.retryLimit {
			return m.Nack(fmt.Sprintf("%s - Reached the max %d number of retries.", reason, m.retryLimit))
		}

		headers := make(map[string]interface{})
		headers["x-retry-count"] = int64(retryCount)

		payload := amqp.Publishing{
			Body:      m.Body(),
			Headers:   headers,
			Timestamp: time.Now(),
		}

		err := m.Ack()

		if err != nil {
			return err
		}

		return m.dleChannel.Publish(m.retryExchangeName, m.delivery.RoutingKey, false, false, payload)
	}

	return m.delivery.Reject(true)

}
