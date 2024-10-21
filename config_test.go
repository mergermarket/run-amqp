package runamqp

import (
	"testing"

	"github.com/mergermarket/run-amqp/helpers"
)

const defaultPrefetch = 10

func TestItDerivesConsumerExchanges(t *testing.T) {

	logger := helpers.NewTestLogger(t)

	c := NewConsumerConfig{
		URL:          testRabbitURI,
		exchangeName: "producer-stuff",
		exchangeType: Fanout,
		patterns:     noPatterns,
		logger:       logger,
		requeueTTL:   200,
		requeueLimit: testRequeueLimit,
		serviceName:  "service",
		prefetch:     defaultPrefetch,
	}
	consumerConfig := c.Config()

	expectedQueueName := "producer-stuff-for-service"

	if consumerConfig.queue.Name != expectedQueueName {
		t.Error("Expected", expectedQueueName, consumerConfig.queue.Name)
	}

	expectededDLEName := "producer-stuff-for-service-dle"

	if consumerConfig.exchange.DLE != expectededDLEName {
		t.Error("Expected", expectededDLEName, "but got", consumerConfig.exchange.DLE)
	}

	expectedRetryNowExchangeName := "producer-stuff-for-service-retry-now"

	if consumerConfig.exchange.RetryNow != expectedRetryNowExchangeName {
		t.Error("Expected", expectedRetryNowExchangeName, "but got", consumerConfig.exchange.RetryNow)
	}

	expectedRetryLaterExchangeName := "producer-stuff-for-service-retry-200ms-later"

	if consumerConfig.exchange.RetryLater != expectedRetryLaterExchangeName {
		t.Error("Expected", expectedRetryLaterExchangeName, "but got", expectedRetryLaterExchangeName)
	}

	expectedDLQName := "producer-stuff-for-service-dlq"

	if consumerConfig.queue.DLQ != expectedDLQName {
		t.Error("Expected", expectedDLQName, "but got", consumerConfig.queue.DLQ)
	}

	expectedRetryLaterQueueName := "producer-stuff-for-service-retry-200ms-later"
	if consumerConfig.queue.RetryLater != expectedRetryLaterQueueName {
		t.Error("Expected", expectedRetryLaterQueueName, "but got", consumerConfig.queue.RetryLater)
	}

	if consumerConfig.queue.PrefetchCount != defaultPrefetch {
		t.Error("Expected prefetch of", defaultPrefetch, "but got", consumerConfig.queue.PrefetchCount)
	}
}

func TestItSetsPatternToHashWhenNoneSupplied(t *testing.T) {
	logger := helpers.NewTestLogger(t)

	c := NewConsumerConfig{
		URL:          testRabbitURI,
		exchangeName: "exchange",
		exchangeType: Fanout,
		patterns:     noPatterns,
		logger:       logger,
		requeueTTL:   200,
		requeueLimit: testRequeueLimit,
		serviceName:  "service",
		prefetch:     defaultPrefetch,
	}
	consumerConfig := c.Config()

	if len(consumerConfig.queue.Patterns) != 1 {
		t.Fatal("When there are no patterns supplied it should've put one in")
	}

	if consumerConfig.queue.Patterns[0] != "#" {
		t.Error("Config should set the pattern to # (for catch all) if none are supplied but got", consumerConfig.queue.Patterns)
	}
}

func TestItSetsPatternsOnQueue(t *testing.T) {
	logger := helpers.NewTestLogger(t)
	pattern := "pretty.pattern"
	c := NewConsumerConfig{
		URL:          testRabbitURI,
		exchangeName: "exchange",
		exchangeType: Fanout,
		patterns:     []string{pattern},
		logger:       logger,
		requeueTTL:   200,
		requeueLimit: testRequeueLimit,
		serviceName:  "service",
		prefetch:     defaultPrefetch,
	}
	consumerConfig := c.Config()

	if consumerConfig.queue.MaxPriority != 0 {
		t.Error("Expected max priority to be set to 0 got", consumerConfig.queue.MaxPriority)
	}
	if len(consumerConfig.queue.Patterns) != 1 {
		t.Fatal("There should be one pattern set")
	}

	if consumerConfig.queue.Patterns[0] != pattern {
		t.Error("Unexpected pattern, expected", pattern, "but got", consumerConfig.queue.Patterns[0])
	}
}

func TestItSetsMaxPriority(t *testing.T) {
	logger := helpers.NewTestLogger(t)
	pattern := "pretty.pattern"
	c := NewConsumerConfig{
		URL:          testRabbitURI,
		exchangeName: "exchange",
		exchangeType: Fanout,
		patterns:     []string{pattern},
		logger:       logger,
		requeueTTL:   200,
		requeueLimit: testRequeueLimit,
		serviceName:  "service",
		prefetch:     defaultPrefetch,
		maxPriority:  7,
	}
	consumerConfig := c.Config()

	if consumerConfig.queue.MaxPriority != 7 {
		t.Error("Expected max priority to be set to 7 got", consumerConfig.queue.MaxPriority)
	}
	if len(consumerConfig.queue.Patterns) != 1 {
		t.Fatal("There should be one pattern set")
	}

	if consumerConfig.queue.Patterns[0] != pattern {
		t.Error("Unexpected pattern, expected", pattern, "but got", consumerConfig.queue.Patterns[0])
	}
}
