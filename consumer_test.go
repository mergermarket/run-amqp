package runamqp

import (
	"fmt"
	"github.com/mergermarket/gotools"
	"math/rand"
	"testing"
	"time"
)

var (
	payload = []byte("test")
)

const testRabbitURI = "amqp://guest:guest@rabbitmq:5672/"
const testRequeueTTL = 200
const testRequeueLimit = 5
const serviceName = "testservice"

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var noPatterns = []string{""}

func TestConsumerConsumesMessages(t *testing.T) {
	t.Parallel()
	config := NewConsumerConfig(
		testRabbitURI,
		"test-exchange",
		Fanout,
		"test-queue",
		noPatterns,
		tools.TestLogger{T: t},
		testRequeueTTL,
		testRequeueLimit,
		serviceName,
	)

	consumer, queuesReady := NewConsumer(config)

	if ok := <-queuesReady; !ok {
		t.Fatal("Didn't bind original queues")
	}

	publisherConfig := NewPublisherConfig(config.URL, config.exchange.Name, config.exchange.Type, config.Logger)

	publish, publishReady := NewPublisher(publisherConfig)
	if ok := <-publishReady; !ok {
		t.Fatal("Is not ready to publish")
	}

	err := publish(payload, "")

	if err != nil {
		t.Fatal("Error when Publishing the message")
	}

	message := <-consumer

	if string(message.Body()) != string(payload) {
		t.Fatal("failed to publish")
	}

	err = message.Ack()

	if err != nil {
		t.Fatal("Error when Acking the message")
	}
}

func TestDLQ(t *testing.T) {
	t.Parallel()

	config := NewConsumerConfig(
		"amqp://guest:guest@rabbitmq:5672/",
		"test-exchange-"+randomString(5),
		Fanout,
		"test-queue-"+randomString(5),
		noPatterns,
		tools.TestLogger{T: t},
		testRequeueTTL,
		testRequeueLimit,
		serviceName,
	)

	consumer, queuesReady := NewConsumer(config)

	if ok := <-queuesReady; !ok {
		t.Fatal("Didn't bind original queues")
	}

	publisherConfig := NewPublisherConfig(config.URL, config.exchange.Name, config.exchange.Type, config.Logger)

	publish, publishReady := NewPublisher(publisherConfig)

	if ok := <-publishReady; !ok {
		t.Fatal("Is not ready to publish")
	}

	dlqConfig := NewConsumerConfig(
		"amqp://guest:guest@rabbitmq:5672/",
		config.exchange.DLE,
		Fanout,
		config.queue.DLQ,
		noPatterns,
		tools.TestLogger{T: t},
		testRequeueTTL,
		testRequeueLimit,
		serviceName,
	)

	dlqConsumer, dlqQueuesReady := newDirectConsumer(dlqConfig)

	if ok := <-dlqQueuesReady; !ok {
		t.Fatal("Didnt bind DLQ")
	}

	if err := publish(payload, ""); err != nil {
		t.Fatal("Error when Publishing the message")
	}

	message := <-consumer

	rejectReason := "chris is poo"

	if err := message.Nack(rejectReason); err != nil {
		t.Fatal("Error when Nacking the message")
	}

	dlqMessage := <-dlqConsumer

	if string(dlqMessage.Body()) != string(payload) {
		t.Fatal("failed to get dlq'd message")
	}

	if dlqMessage.delivery.Headers["x-dle-reason"] != rejectReason {
		t.Fatal("x-dle-reason was not set correctly")
	}

}

func TestRequeue(t *testing.T) {
	t.Parallel()

	const twoRetries = 2
	patterns := []string{"*.notifications.bounced", "*.notifications.dropped"}
	config := NewConsumerConfig(
		"amqp://guest:guest@rabbitmq:5672/",
		"test-exchange-"+randomString(5),
		Topic,
		"test-queue-"+randomString(5),
		patterns,
		tools.TestLogger{T: t},
		testRequeueTTL,
		twoRetries,
		serviceName,
	)

	consumer, queuesReady := NewConsumer(config)

	if ok := <-queuesReady; !ok {
		t.Fatal("Didn't bind original queues")
	}

	publisherConfig := NewPublisherConfig(config.URL, config.exchange.Name, config.exchange.Type, config.Logger)

	publish, publishReady := NewPublisher(publisherConfig)

	if ok := <-publishReady; !ok {
		t.Fatal("Is not ready to publish")
	}

	if err := publish(payload, "all.notifications.bounced"); err != nil {
		t.Fatal("Error when Publishing the message")
	}

	var publishedMessage *Message
	select {
	case msg := <-consumer:
		publishedMessage = msg
	case <-time.After(1 * time.Second):
		t.Fatal("Timedout waiting for the consumer")
	}

	if publishedMessage == nil {
		t.Fatal("Did not get the published message")
	}

	if err := publishedMessage.Requeue("Requeuing the first message"); err != nil {
		t.Fatal("Could not REQUEUE the message")
	}

	var requeuedMessage *Message
	select {
	case reqMsg := <-consumer:
		requeuedMessage = reqMsg
	case <-time.After(1 * time.Second):
		t.Fatal("Timedout waiting for the consumer")
	}

	if requeuedMessage == nil {
		t.Fatal("Did not get the requeued message")
	}

	actualMessage := string(requeuedMessage.Body())
	expectedMessage := string(payload)

	if actualMessage != expectedMessage {
		t.Fatalf("Failed to get the requeued message: %s but got %s", expectedMessage, actualMessage)
	}

	if _, found := requeuedMessage.delivery.Headers["x-retry-count"]; !found {
		t.Fatal("x-retry-count was not set correctly")
	}

}

func TestRequeue_DLQ_Message_After_Retries(t *testing.T) {
	t.Parallel()

	oneRetry := 1
	config := NewConsumerConfig(
		"amqp://guest:guest@rabbitmq:5672/",
		"test-exchange-"+randomString(5),
		Fanout,
		"test-queue-"+randomString(5),
		noPatterns,
		tools.TestLogger{T: t},
		testRequeueTTL,
		oneRetry,
		serviceName,
	)

	consumer, queuesReady := NewConsumer(config)

	if ok := <-queuesReady; !ok {
		t.Fatal("Didn't bind original queues")
	}

	dlqConfig := NewConsumerConfig(
		config.URL,
		config.exchange.DLE,
		config.exchange.Type,
		config.queue.DLQ,
		noPatterns,
		config.Logger,
		testRequeueTTL,
		oneRetry,
		serviceName,
	)
	dlqConsumer, dlqQueuesReady := newDirectConsumer(dlqConfig)

	if ok := <-dlqQueuesReady; !ok {
		t.Fatal("Didnt bind DLQ")
	}

	publisherConfig := NewPublisherConfig(config.URL, config.exchange.Name, config.exchange.Type, config.Logger)

	publish, publishReady := NewPublisher(publisherConfig)

	if ok := <-publishReady; !ok {
		t.Fatal("Is not ready to publish")
	}

	if err := publish(payload, ""); err != nil {
		t.Fatal("Error when Publishing the message")
	}

	publishedMessage := <-consumer

	if publishedMessage == nil {
		t.Fatal("Did not get the published message")
	}

	if err := publishedMessage.Requeue("Requeuing the first message"); err != nil {
		t.Fatal("Could not REQUEUE the message")
	}

	firstRequeuedMessage := <-consumer

	if firstRequeuedMessage == nil {
		t.Fatal("Did not get the requeued message")
	}

	actualMessage := string(firstRequeuedMessage.Body())
	expectedMessage := string(payload)

	if actualMessage != expectedMessage {
		t.Fatalf("Failed to get the requeued message: %s but got %s", expectedMessage, actualMessage)
	}

	if _, found := firstRequeuedMessage.delivery.Headers["x-retry-count"]; !found {
		t.Fatal("x-retry-count was not set correctly")
	}

	if err := firstRequeuedMessage.Requeue("This should end up in the DLQ"); err != nil {
		t.Fatal("Could not REQUEUE the message for the second time")
	}

	dlqMessage := <-dlqConsumer

	if string(dlqMessage.Body()) != string(payload) {
		t.Fatal("failed to get dlq'd message")
	}

	if dlqMessage.delivery.Headers["x-dle-reason"] != "This should end up in the DLQ - Reached the max 1 number of retries." {
		t.Fatal("x-dle-reason was not set correctly")
	}
}

func TestRequeue_With_No_Requeue_Limit(t *testing.T) {
	t.Parallel()

	noTTL := int16(0)
	noRetries := 0

	config := NewConsumerConfig(
		"amqp://guest:guest@rabbitmq:5672/",
		"test-exchange-"+randomString(5),
		Fanout,
		"test-queue-"+randomString(5),
		noPatterns,
		tools.TestLogger{T: t},
		noTTL,
		noRetries,
		serviceName,
	)

	consumer, queuesReady := NewConsumer(config)

	if ok := <-queuesReady; !ok {
		t.Fatal("Didn't bind original queues")
	}

	publisherConfig := NewPublisherConfig(config.URL, config.exchange.Name, config.exchange.Type, config.Logger)

	publish, publishReady := NewPublisher(publisherConfig)

	if ok := <-publishReady; !ok {
		t.Fatal("Is not ready to publish")
	}

	if err := publish(payload, ""); err != nil {
		t.Fatal("Error when Publishing the message")
	}

	counter := 1
	for ; counter < 10; counter++ {
		msg := <-consumer
		actualMessage := string(msg.Body())
		expectedMessage := string(payload)

		if actualMessage != expectedMessage {
			t.Fatalf("Failed to get the requeued message: %s but got %s", expectedMessage, actualMessage)
		}
		msg.Requeue(fmt.Sprintf("Requing it for the %d time.", counter))
	}

	if counter != 10 {
		t.Fatalf("Counter: %d, expecting 10", counter)
	}
}
func TestPatterns(t *testing.T) {
	t.Parallel()

	patterns := []string{"A", "B"}

	config := NewConsumerConfig(
		testRabbitURI,
		"test-exchange-"+randomString(5),
		Topic,
		"test-queue-"+randomString(5),
		patterns,
		tools.TestLogger{T: t},
		testRequeueTTL,
		testRequeueLimit,
		serviceName,
	)

	consumer, queuesReady := NewConsumer(config)

	if ok := <-queuesReady; !ok {
		t.Fatal("Didn't bind original queues")
	}

	publisherConfig := NewPublisherConfig(config.URL, config.exchange.Name, config.exchange.Type, config.Logger)

	publish, publishReady := NewPublisher(publisherConfig)
	if ok := <-publishReady; !ok {
		t.Fatal("Is not ready to publish")
	}

	gotMessageForPattern := func(msg, pattern string) bool {

		if err := publish([]byte(msg), pattern); err != nil {
			t.Fatal("Error when Publishing the message")
		}

		select {
		case message := <-consumer:
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

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
