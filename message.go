package runamqp

import (
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

//todo: make a message interface and turn "Message" into an amqpMessage

// Message is wrapper for the RabbitMq delivery
type Message struct {
	delivery               amqp.Delivery
	amqpChannel            *amqp.Channel
	retryLimit             int
	retryLaterExchangeName string
	dleExchangeName        string
}

// Body returns the body of the AMQP message
func (m *Message) Body() []byte {
	return m.delivery.Body
}

// Ack will acknowledge the message.
func (m *Message) Ack() error {
	return m.delivery.Ack(false)
}

// Nack is used when you cant process a message. The "reason" will appear in the rabbit console under the message headers which is useful for debugging
func (m *Message) Nack(reason string) error {

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

	err = m.amqpChannel.Publish(m.dleExchangeName, m.delivery.RoutingKey, false, false, payload)

	if err != nil {
		return err
	}

	return nil
}

// Requeue requeues a message, which is useful for when you have transient problems
func (m Message) Requeue(reason string) error {

	if m.retryLimit > 0 {
		retryCount := 1
		if headerRetryCount, found := m.delivery.Headers["x-retry-count"]; found {

			if temp, ok := headerRetryCount.(int); ok {
				retryCount = temp + 1
			} else {
				retryCount++
			}

		}

		if retryCount > m.retryLimit {
			err := m.Nack(fmt.Sprintf("%s - Reached the max %d number of retries.", reason, m.retryLimit))

			if err != nil {
				return err
			}

			return nil
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

		err = m.amqpChannel.Publish(m.retryLaterExchangeName, m.delivery.RoutingKey, false, false, payload)

		if err != nil {
			return err
		}

		return nil

	}

	return m.delivery.Reject(true)

}
