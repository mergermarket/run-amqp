package runamqp

import (
	"github.com/streadway/amqp"
	"testing"
	"time"
)

func TestRabbitState_reconnect(t *testing.T) {
	config := NewPublisherConfig(testRabbitURI, "test-exchange-"+randomString(5), Fanout, &testLogger{t})
	state := makeNewConnectedRabbit(config.connection, config.exchange)

	select {
	case <-state.newlyOpenedChannels:
		t.Log("Opened AMQP channel")
	case <-time.After(5 * time.Second):
		t.Fatal("Did not connect to AMQP channel in time")
	}

	state.channelErrors <- amqp.ErrClosed

	select {
	case <-state.newlyOpenedChannels:
		t.Log("Opened  another AMQP channel")
	case <-time.After(5 * time.Second):
		t.Fatal("Did not recconnect to AMQP channel in time")
	}

}
