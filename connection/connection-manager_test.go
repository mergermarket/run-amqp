package connection

import (
	"github.com/mergermarket/run-amqp/helpers"
	"github.com/streadway/amqp"
	"testing"
	"time"
)

func TestNewConnectionManager_OpenChannel(t *testing.T) {
	logger := helpers.NewTestLogger(t)

	t.Run("should open a managed connection with channel", func(t *testing.T) {

		manager := NewConnectionManager(testRabbitURI, logger)

		channel := manager.OpenChannel("test channel")

		select {
		case <-channel:
		case <-time.After(2 * time.Second):
			t.Fatal("failed to get first channels in time")
		}
	})

	t.Run("should re-open connections and all the associated channels due to some connection error", func(t *testing.T) {

		manager := NewConnectionManager(testRabbitURI, logger)

		firstChannels := manager.OpenChannel("firstChannels")

		select {
		case <-firstChannels:
		case <-time.After(2 * time.Second):
			t.Fatal("failed to get first channels in time")
		}

		secondChannels := manager.OpenChannel("seconChannels")

		select {
		case <-secondChannels:
		case <-time.After(2 * time.Second):
			t.Fatal("failed to get second channels in time")
		}

		manager.sendConnectionError(amqp.ErrClosed)

		select {
		case <-firstChannels:
		case <-time.After(2 * time.Second):
			t.Fatal("did not get first channels in time after re-connecting")
		}

		select {
		case <-secondChannels:
		case <-time.After(2 * time.Second):
			t.Fatal("did not get first channels in time after re-connecting")
		}
	})

	t.Run("should re-open the channel that is closed due to some the channel error", func(t *testing.T) {

		manager := NewConnectionManager(testRabbitURI, logger)

		channelNotToBeReOpened := manager.OpenChannel("channelNotToBeReOpened")

		select {
		case <-channelNotToBeReOpened:
		case <-time.After(2 * time.Second):
			t.Fatal("failed to get 'channelNotToBeReOpened' channels in time")
		}

		channelToBeOpenedAfterChannelError := manager.OpenChannel("channelToBeOpenedAfterChannelError")

		select {
		case <-channelToBeOpenedAfterChannelError:
		case <-time.After(2 * time.Second):
			t.Fatal("failed to get 'channelToBeOpenedAfterChannelError' channels in time")
		}

		err := manager.sendChannelError(1, amqp.ErrClosed)

		if err != nil {
			t.Fatal("failed to send error to the specified channel", err)
		}

		select {
		case <-channelNotToBeReOpened:
			t.Fatal("'channelNotToBeReOpened' should not have been re-opened")
		case <-time.After(2 * time.Second):

		}

		select {
		case <-channelToBeOpenedAfterChannelError:
		case <-time.After(2 * time.Second):
			t.Fatal("did not get a new channel for 'channelToBeOpenedAfterChannelError' in time after re-connecting")
		}
	})
}
