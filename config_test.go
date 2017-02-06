package runamqp

import (
	"github.com/mergermarket/run-amqp/helpers"
	"testing"
)

func TestItDerivesConsumerExchanges(t *testing.T) {

	logger := helpers.NewTestLogger(t)

	consumerConfig := NewConsumerConfig(
		testRabbitURI,
		"producer-stuff",
		Fanout,
		noPatterns,
		logger,
		200,
		testRequeueLimit,
		"service",
	)

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
}

func TestItSetsPatternToHashWhenNoneSupplied(t *testing.T) {
	logger := helpers.NewTestLogger(t)

	consumerConfig := NewConsumerConfig(
		testRabbitURI,
		"exchange",
		Fanout,
		noPatterns,
		logger,
		200,
		testRequeueLimit,
		"service",
	)

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
	consumerConfig := NewConsumerConfig(
		testRabbitURI,
		"exchange",
		Fanout,
		[]string{pattern},
		logger,
		200,
		testRequeueLimit,
		"service",
	)

	if len(consumerConfig.queue.Patterns) != 1 {
		t.Fatal("There should be one pattern set")
	}

	if consumerConfig.queue.Patterns[0] != pattern {
		t.Error("Unexpected pattern, expected", pattern, "but got", consumerConfig.queue.Patterns[0])
	}
}
