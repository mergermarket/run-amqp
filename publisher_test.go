package runamqp

import (
	"github.com/mergermarket/run-amqp/helpers"
	"testing"
)

func TestNakedPublisher(t *testing.T) {
	t.Parallel()

	expectedExchangeName := "chris-rulz" + randomString(5)
	config := NewPublisherConfig(testRabbitURI, expectedExchangeName, Fanout, true, helpers.NewTestLogger(t))

	publisher, err := NewPublisher(config)

	if err != nil {
		t.Fatal("problem creating publisher", err)
	}

	err = publisher.Publish([]byte("whatever"), nil)

	if err != nil {
		t.Error("Should not get an error", err)
	}
}

func TestNotReadyPublisherErrorsOnPublish(t *testing.T) {
	t.Parallel()

	expectedExchangeName := "chris-rulz" + randomString(5)
	config := NewPublisherConfig(testRabbitURI, expectedExchangeName, Fanout, true, helpers.NewTestLogger(t))

	publisher, err := NewPublisher(config)

	if err != nil {
		t.Fatal("problem creating publisher", err)
	}

	publisher.publishReady = false

	err = publisher.Publish([]byte("whatever"), nil)

	if err == nil {
		t.Error("Should get an error")
	}
}
