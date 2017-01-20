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
		"queue",
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

	expectedDLQName := "queue-dlq"

	if consumerConfig.queue.DLQ != expectedDLQName {
		t.Error("Expected", expectedDLQName, "but got", consumerConfig.queue.DLQ)
	}

	expectedRetryLaterQueueName := "queue-retry-200ms-later"
	if consumerConfig.queue.RetryLater != expectedRetryLaterQueueName {
		t.Error("Expected", expectedRetryLaterQueueName, "but got", consumerConfig.queue.RetryLater)
	}
}
