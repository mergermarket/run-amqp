package runamqp

import (
	"testing"

	"github.com/mergermarket/run-amqp/helpers"
)

func TestNakedPublisher(t *testing.T) {
	t.Parallel()

	expectedExchangeName := "chris-rulz" + randomString(5)
	c := NewPublisherConfig{
		URL:          testRabbitURI,
		ExchangeName: expectedExchangeName,
		ExchangeType: Fanout,
		Confirmable:  true,
		Logger:       helpers.NewTestLogger(t),
	}
	config := c.Config()

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
	c := NewPublisherConfig{
		URL:          testRabbitURI,
		ExchangeName: expectedExchangeName,
		ExchangeType: Fanout,
		Confirmable:  true,
		Logger:       helpers.NewTestLogger(t),
	}
	config := c.Config()

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
