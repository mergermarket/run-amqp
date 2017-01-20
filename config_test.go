package runamqp

import (
	"testing"
)

func TestItDerivesConsumerExchanges(t *testing.T) {

	logger := &testLogger{t: t}

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

	expectededDLEName := "exchange-for-service-dle"

	if consumerConfig.exchange.DLE != expectededDLEName {
		t.Error("Expected", expectededDLEName, "but got", consumerConfig.exchange.DLE)
	}

	expectedRetryNowExchangeName := "exchange-for-service-retry-now"

	if consumerConfig.exchange.RetryNow != expectedRetryNowExchangeName {
		t.Error("Expected", expectedRetryNowExchangeName, "but got", consumerConfig.exchange.RetryNow)
	}

	expectedRetryLaterExchangeName := "exchange-for-service-retry-200ms-later"

	if consumerConfig.exchange.RetryLater != expectedRetryLaterExchangeName {
		t.Error("Expected", expectedRetryLaterExchangeName, "but got", expectedRetryLaterExchangeName)
	}

	expectedDLQName := "exchange-for-service-queue-dlq"

	if consumerConfig.queue.DLQ != expectedDLQName {
		t.Error("Expected", expectedDLQName, "but got", consumerConfig.queue.DLQ)
	}

	expectedRetryLaterQueueName := "exchange-for-service-queue-retry-200ms-later"
	if consumerConfig.queue.RetryLater != expectedRetryLaterQueueName {
		t.Error("Expected", expectedRetryLaterQueueName, "but got", consumerConfig.queue.RetryLater)
	}
}

func TestItSetsPatternToHashWhenNoneSupplied(t *testing.T) {
	consumerConfig := NewConsumerConfig(
		testRabbitURI,
		"exchange",
		Fanout,
		noPatterns,
		&testLogger{t: t},
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
	pattern := "pretty.pattern"
	consumerConfig := NewConsumerConfig(
		testRabbitURI,
		"exchange",
		Fanout,
		[]string{pattern},
		&testLogger{t: t},
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
