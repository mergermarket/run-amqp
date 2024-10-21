package runamqp

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/mergermarket/run-amqp/helpers"
)

var (
	payload = []byte("test")
)

const testRabbitURI = "amqp://guest:guest@rabbitmq:5672/"
const testRequeueTTL = 200
const testRequeueLimit = 5
const serviceName = "testservice"

func TestConsumerConsumesMessages(t *testing.T) {
	t.Parallel()

	consumer1Config := newTestConsumerConfig(t, consumerConfigOptions{})
	consumer2Config := newTestConsumerConfig(t, consumerConfigOptions{
		ExchangeName: consumer1Config.exchange.Name,
		ServiceName:  consumer1Config.queue.Name + "-second",
	})

	publisher, _ := NewPublisher(consumer1Config.NewPublisherConfig())

	consumer1 := NewConsumer(consumer1Config)
	assertReady(t, consumer1.QueuesBound)

	consumer2 := NewConsumer(consumer2Config)
	assertReady(t, consumer2.QueuesBound)

	err := publisher.Publish(payload, nil)
	if err != nil {
		t.Fatal("Error when Publishing the message")
	}

	message := getMessage(t, consumer1.Messages)
	if string(message.Body()) != string(payload) {
		t.Fatal("failed to publish - consumer1")
	}
	err = message.Ack()
	if err != nil {
		t.Fatal("Error when Acking the message - consumer1")
	}
	t.Log("First consumer - done")

	message = getMessage(t, consumer2.Messages)
	if string(message.Body()) != string(payload) {
		t.Fatal("failed to publish - consumer2")
	}
	err = message.Ack()
	if err != nil {
		t.Fatal("Error when Acking the message - consumer2")
	}
}

func TestPublishAndConsumeFromConfirmChannelMessages(t *testing.T) {
	t.Parallel()

	consumer1Config := newTestConsumerConfig(t, consumerConfigOptions{})
	consumer2Config := newTestConsumerConfig(t, consumerConfigOptions{
		ExchangeName: consumer1Config.exchange.Name,
		ServiceName:  consumer1Config.queue.Name + "-second",
	})

	config := consumer1Config.NewPublisherConfig()
	config.confirmable = true

	publisher, err := NewPublisher(config)
	assertNoError(t, err)

	consumer1 := NewConsumer(consumer1Config)
	assertReady(t, consumer1.QueuesBound)

	consumer2 := NewConsumer(consumer2Config)
	assertReady(t, consumer2.QueuesBound)

	err = publisher.Publish(payload, nil)
	if err != nil {
		t.Fatal("Error when Publishing the message")
	}

	message := getMessage(t, consumer1.Messages)
	if string(message.Body()) != string(payload) {
		t.Fatal("failed to publish - consumer1")
	}
	err = message.Ack()
	if err != nil {
		t.Fatal("Error when Acking the message - consumer1")
	}
	t.Log("First consumer - done")

	message = getMessage(t, consumer2.Messages)
	if string(message.Body()) != string(payload) {
		t.Fatal("failed to publish - consumer2")
	}
	err = message.Ack()
	if err != nil {
		t.Fatal("Error when Acking the message - consumer2")
	}
}

func TestDLQ(t *testing.T) {
	t.Parallel()

	consumerConfig := newTestConsumerConfig(t, consumerConfigOptions{})
	consumer := NewConsumer(consumerConfig)

	assertReady(t, consumer.QueuesBound)

	publisher, err := NewPublisher(consumerConfig.NewPublisherConfig())
	assertNoError(t, err)

	dlqConfig := newTestConsumerConfig(t, consumerConfigOptions{
		ExchangeName: consumerConfig.exchange.DLE,
	})

	dlqConsumer := NewConsumer(dlqConfig)

	assertReady(t, dlqConsumer.QueuesBound)

	if err = publisher.Publish(payload, nil); err != nil {
		t.Fatal("Error when Publishing the message")
	}

	message := getMessage(t, consumer.Messages)

	rejectReason := "chris is poo"

	if err = message.Nack(rejectReason); err != nil {
		t.Fatal("Error when Nacking the message")
	}

	dlqMessage := getMessage(t, dlqConsumer.Messages)

	if string(dlqMessage.Body()) != string(payload) {
		t.Fatal("failed to get dlq'd message")
	}

	dlqMessageAMQP, _ := dlqMessage.(*amqpMessage)

	if dlqMessageAMQP.delivery.Headers["x-dle-reason"] != rejectReason {
		t.Fatal("x-dle-reason was not set correctly")
	}

	timestamp, err := time.Parse(time.RFC3339, dlqMessageAMQP.delivery.Headers["x-dle-timestamp"].(string))
	if err != nil {
		t.Fatalf("x-dle-timestamp cannot be parsed: %v", err)
	}
	difference := time.Since(timestamp)
	if difference > time.Duration(2*time.Second) {
		t.Fatalf("x-dle-timestamp was not set correctly - difference: %s, timestamp: %s, now: %s", difference, timestamp, time.Now())
	}

}

func TestRequeue(t *testing.T) {
	t.Parallel()

	const totalRetries = 3
	patterns := []string{"*.notifications.bounced", "*.notifications.dropped"}
	consumerConfig := newTestConsumerConfig(t, consumerConfigOptions{
		Patterns: patterns,
		Retries:  totalRetries,
	})

	consumer := NewConsumer(consumerConfig)

	assertReady(t, consumer.QueuesBound)

	publisher, err := NewPublisher(consumerConfig.NewPublisherConfig())
	assertNoError(t, err)

	if err := publisher.Publish(payload, &PublishOptions{Pattern: "all.notifications.bounced"}); err != nil {
		t.Fatal("Error when Publishing the message")
	}

	publishedMessage := getMessage(t, consumer.Messages)

	if publishedMessage == nil {
		t.Fatal("Did not get the published message")
	}

	for retryCount := 1; retryCount <= totalRetries; retryCount++ {

		if err := publishedMessage.Requeue("Requeuing the message"); err != nil {
			t.Fatal("Could not REQUEUE the message")
		}

		publishedMessage = getMessage(t, consumer.Messages)

		if publishedMessage == nil {
			t.Fatal("Did not get the requeued message")
		}

		actualMessage := string(publishedMessage.Body())
		expectedMessage := string(payload)

		if actualMessage != expectedMessage {
			t.Fatalf("Failed to get the requeued message: %s but got %s", expectedMessage, actualMessage)
		}

		amqpMsg, _ := publishedMessage.(*amqpMessage)

		count, found := amqpMsg.delivery.Headers["x-retry-count"]
		if !found {
			t.Fatal("x-retry-count was not set correctly")
		}

		if count.(int64) != int64(retryCount) {
			t.Error("First retry count should be", retryCount, "but its", count)
		}
	}

}

func TestRequeue_DLQ_Message_After_Retries(t *testing.T) {
	t.Parallel()

	oneRetry := 1
	consumerConfig := newTestConsumerConfig(t, consumerConfigOptions{
		Retries: oneRetry,
	})

	consumer := NewConsumer(consumerConfig)

	assertReady(t, consumer.QueuesBound)

	dlqConfig := newTestConsumerConfig(t, consumerConfigOptions{
		ExchangeName: consumerConfig.exchange.DLE,
		Retries:      oneRetry,
	})

	dlqConsumer := NewConsumer(dlqConfig)

	assertReady(t, dlqConsumer.QueuesBound)

	publisher, err := NewPublisher(consumerConfig.NewPublisherConfig())
	assertNoError(t, err)

	if err := publisher.Publish(payload, nil); err != nil {
		t.Fatal("Error when Publishing the message")
	}

	publishedMessage := getMessage(t, consumer.Messages)

	if publishedMessage == nil {
		t.Fatal("Did not get the published message")
	}

	if err := publishedMessage.Requeue("Requeuing the first message"); err != nil {
		t.Fatal("Could not REQUEUE the message")
	}

	firstRequeuedMessage := getMessage(t, consumer.Messages)

	if firstRequeuedMessage == nil {
		t.Fatal("Did not get the requeued message")
	}

	actualMessage := string(firstRequeuedMessage.Body())
	expectedMessage := string(payload)

	if actualMessage != expectedMessage {
		t.Fatalf("Failed to get the requeued message: %s but got %s", expectedMessage, actualMessage)
	}

	firstMsgAmqp, _ := firstRequeuedMessage.(*amqpMessage)

	if _, found := firstMsgAmqp.delivery.Headers["x-retry-count"]; !found {
		t.Fatal("x-retry-count was not set correctly")
	}

	if err := firstRequeuedMessage.Requeue("This should end up in the DLQ"); err != nil {
		t.Fatal("Could not REQUEUE the message for the second time")
	}

	dlqMessage := getMessage(t, dlqConsumer.Messages)

	if string(dlqMessage.Body()) != string(payload) {
		t.Fatal("failed to get dlq'd message")
	}

	dlqMsgAMQP, _ := dlqMessage.(*amqpMessage)

	if dlqMsgAMQP.delivery.Headers["x-dle-reason"] != "This should end up in the DLQ - Reached the max 1 number of retries." {
		t.Fatal("x-dle-reason was not set correctly")
	}
}

func TestRequeue_With_No_Requeue_Limit(t *testing.T) {
	t.Parallel()

	consumerConfig := newTestConsumerConfig(t, consumerConfigOptions{
		SetNoRetries: true,
	})

	consumer := NewConsumer(consumerConfig)

	assertReady(t, consumer.QueuesBound)

	publisher, err := NewPublisher(consumerConfig.NewPublisherConfig())
	assertNoError(t, err)

	if err := publisher.Publish(payload, nil); err != nil {
		t.Fatal("Error when Publishing the message")
	}

	counter := 1
	for ; counter < 10; counter++ {
		msg := getMessage(t, consumer.Messages)
		actualMessage := string(msg.Body())
		expectedMessage := string(payload)

		if actualMessage != expectedMessage {
			t.Fatalf("Failed to get the requeued message: %s but got %s", expectedMessage, actualMessage)
		}
		err := msg.Requeue(fmt.Sprintf("Requing it for the %d time.", counter))
		if err != nil {
			t.Fatal("Failed to requeue message")
		}
	}

	if counter != 10 {
		t.Fatalf("Counter: %d, expecting 10", counter)
	}
}
func TestPatterns(t *testing.T) {
	t.Parallel()

	consumerConfig := newTestConsumerConfig(t, consumerConfigOptions{
		Patterns:     []string{"A", "B"},
		ExchangeType: Topic,
	})

	consumer := NewConsumer(consumerConfig)
	assertReady(t, consumer.QueuesBound)

	publisher, err := NewPublisher(consumerConfig.NewPublisherConfig())
	assertNoError(t, err)

	gotMessageForPattern := func(msg, pattern string) bool {

		if err := publisher.Publish([]byte(msg), &PublishOptions{Pattern: pattern}); err != nil {
			t.Fatal("Error when Publishing the message")
		}

		select {
		case message := <-consumer.Messages:
			if err := message.Ack(); err != nil {
				t.Fatal("Error when Acking the message")
			}
			return string(message.Body()) == msg

		case <-time.After(1 * time.Second):
			t.Log("Timed out waiting for message from consumer on pattern", pattern)
			return false
		}
	}

	if !gotMessageForPattern("a message", "A") {
		t.Error("Did not get message on 'A' routing key")
	}

	if !gotMessageForPattern("b message", "B") {
		t.Error("Did not get message on 'B' routing key")
	}

	if gotMessageForPattern("c message", "C") {
		t.Error("We should not have got a message on C because we dont care about that pattern")
	}
}

func TestPublishToASpecificQueue(t *testing.T) {
	t.Parallel()

	consumerConfig := newTestConsumerConfig(t, consumerConfigOptions{})

	publisher, err := NewPublisher(consumerConfig.NewPublisherConfig())
	assertNoError(t, err)

	consumer := NewConsumer(consumerConfig)
	assertReady(t, consumer.QueuesBound)

	consumerConfigForTargettedQueue := newTestConsumerConfig(t, consumerConfigOptions{
		ExchangeName: consumerConfig.exchange.Name,
		ServiceName:  "test-other-service",
	})
	consumerForTarget := NewConsumer(consumerConfigForTargettedQueue)
	assertReady(t, consumerForTarget.QueuesBound)

	err = publisher.Publish(payload, &PublishOptions{PublishToQueue: consumerConfigForTargettedQueue.queue.Name})

	if err != nil {
		t.Fatal("Error when Publishing the message")
	}

	shouldNotGetMessage(t, consumer.Messages)

	message := getMessage(t, consumerForTarget.Messages)

	if string(message.Body()) != string(payload) {
		t.Fatal("failed to publish")
	}

	err = message.Ack()

	if err != nil {
		t.Fatal("Error when Acking the message")
	}
}

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type consumerConfigOptions struct {
	ExchangeType ExchangeType
	Patterns     []string
	ExchangeName string
	Retries      int
	SetNoRetries bool
	RequeueTTL   int16
	ServiceName  string
}

func newTestConsumerConfig(t *testing.T, config consumerConfigOptions) ConsumerConfig {

	logger := helpers.NewTestLogger(t)

	if config.ExchangeType == "" {
		config.ExchangeType = Fanout
	}

	if len(config.Patterns) == 0 {
		config.Patterns = noPatterns
	}

	if config.ExchangeName == "" {
		config.ExchangeName = "test-exchange-" + randomString(5)
	}

	if config.Retries == 0 {
		config.Retries = testRequeueLimit
	}

	if config.RequeueTTL == 0 {
		config.RequeueTTL = testRequeueTTL
	}

	if config.SetNoRetries {
		config.RequeueTTL = 0
		config.Retries = 0
	}

	if config.ServiceName == "" {
		config.ServiceName = serviceName
	}

	c :=
		NewConsumerConfig{
			URL:          "amqp://guest:guest@rabbitmq:5672/",
			ExchangeName: config.ExchangeName,
			ExchangeType: config.ExchangeType,
			Patterns:     config.Patterns,
			Logger:       logger,
			RequeueTTL:   config.RequeueTTL,
			RequeueLimit: config.Retries,
			ServiceName:  config.ServiceName,
			Prefetch:     defaultPrefetch,
		}
	return c.Config()
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var noPatterns = []string{}

func assertReady(t *testing.T, ch <-chan bool) {
	select {
	case <-ch:
		t.Log("carry on...")
	case <-time.After(5 * time.Second):
		t.Fatal("Didnt get ready in time")
	}
}

func getMessage(t *testing.T, ch <-chan Message) (message Message) {
	select {
	case msg := <-ch:
		message = msg
	case <-time.After(1 * time.Second):
		t.Fatal("Timedout waiting for message")
	}
	return
}

func shouldNotGetMessage(t *testing.T, ch <-chan Message) {
	select {
	case <-ch:
		t.Fatal("Should not have received the message")
	case <-time.After(1 * time.Second):
		t.Log("Mesage was not received")
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
