package runamqp

import (
	"github.com/golang/mock/gomock"
	"github.com/mergermarket/run-amqp/connection"
	"github.com/mergermarket/run-amqp/helpers"
	"github.com/streadway/amqp"
	"testing"
)

type stubbedConnecionManager struct {
	channel *connection.MockAMQPChannel
}

func (s *stubbedConnecionManager) OpenChannel(description string) chan connection.AMQPChannel {
	ch := make(chan connection.AMQPChannel, 1)
	ch <- s.channel
	return ch
}

func TestMainChannelConfiguration(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("main exchange with queues with no routing key pattern", func(t *testing.T) {
		stubChannel := connection.NewMockAMQPChannel(mockCtrl)

		testLogger := helpers.NewTestLogger(t)
		consumerConfig := NewConsumerConfig("url", testExchangeName, Fanout, noPatterns, testLogger, testRequeueTTL, testRequeueLimit, serviceName)
		consumerChannels := newConsumerChannels(consumerConfig)

		stubChannel.EXPECT().ExchangeDeclare(consumerConfig.exchange.Name,
			string(consumerConfig.exchange.Type),
			true,
			false,
			false,
			false,
			nil).Return(nil)

		stubChannel.EXPECT().QueueDeclare(consumerConfig.queue.Name, // name
			true,  // durable
			false, // delete when usused
			false, // exclusive
			false, // no-wait
			nil).Return(amqp.Queue{}, nil)
		//
		stubChannel.EXPECT().QueueBind(consumerConfig.queue.Name, "#", consumerConfig.exchange.Name, false, nil).Return(nil)

		err := consumerChannels.setUpMainExchangeWithQueue(stubChannel)

		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("main exchange and queues with multiple routing key patterns", func(t *testing.T) {
		stubChannel := connection.NewMockAMQPChannel(mockCtrl)

		firstPattern := "test1.test.test"
		secondPattern := "test2.test.test"
		patterns := []string{firstPattern, secondPattern}

		testLogger := helpers.NewTestLogger(t)
		consumerConfig := NewConsumerConfig("url", testExchangeName, Topic, patterns, testLogger, testRequeueTTL, testRequeueLimit, serviceName)
		consumerChannels := newConsumerChannels(consumerConfig)

		stubChannel.EXPECT().ExchangeDeclare(consumerConfig.exchange.Name,
			string(consumerConfig.exchange.Type),
			true,
			false,
			false,
			false,
			nil).Return(nil)

		stubChannel.EXPECT().QueueDeclare(consumerConfig.queue.Name, // name
			true,  // durable
			false, // delete when usused
			false, // exclusive
			false, // no-wait
			nil).Return(amqp.Queue{}, nil)
		//
		stubChannel.EXPECT().QueueBind(consumerConfig.queue.Name, firstPattern, consumerConfig.exchange.Name, false, nil).Return(nil)
		stubChannel.EXPECT().QueueBind(consumerConfig.queue.Name, secondPattern, consumerConfig.exchange.Name, false, nil).Return(nil)

		err := consumerChannels.setUpMainExchangeWithQueue(stubChannel)

		if err != nil {
			t.Fatal(err)
		}
	})

}

func TestDLEChannelConfiguration(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	testLogger := helpers.NewTestLogger(t)

	firstPattern := "test1.test.test"
	secondPattern := "test2.test.test"
	patterns := []string{firstPattern, secondPattern}

	consumerConfig := NewConsumerConfig("url", testExchangeName, Fanout, patterns, testLogger, testRequeueTTL, testRequeueLimit, serviceName)

	t.Run("dle exchange with queues with no routing key pattern", func(t *testing.T) {
		stubChannel := connection.NewMockAMQPChannel(mockCtrl)

		consumerChannels := newConsumerChannels(consumerConfig)

		stubChannel.EXPECT().ExchangeDeclare(consumerConfig.exchange.DLE,
			string(consumerConfig.exchange.Type),
			true,
			false,
			false,
			false,
			nil).Return(nil)

		stubChannel.EXPECT().QueueDeclare(consumerConfig.queue.DLQ, // name
			true,  // durable
			false, // delete when usused
			false, // exclusive
			false, // no-wait
			nil).Return(amqp.Queue{}, nil)
		//
		stubChannel.EXPECT().QueueBind(consumerConfig.queue.DLQ, "#", consumerConfig.exchange.DLE, false, nil).Return(nil)

		err := consumerChannels.setUpDeadLetterExchangeWithQueue(stubChannel)

		if err != nil {
			t.Fatal(err)
		}
	})

}

func TestRetryChannelConfiguration(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	testLogger := helpers.NewTestLogger(t)

	firstPattern := "test1.test.test"
	secondPattern := "test2.test.test"
	patterns := []string{firstPattern, secondPattern}

	consumerConfig := NewConsumerConfig("url", testExchangeName, Fanout, patterns, testLogger, testRequeueTTL, testRequeueLimit, serviceName)

	t.Run("dle exchange with queues with no routing key pattern", func(t *testing.T) {
		stubChannel := connection.NewMockAMQPChannel(mockCtrl)

		consumerChannels := newConsumerChannels(consumerConfig)

		stubChannel.EXPECT().ExchangeDeclare(consumerConfig.exchange.RetryLater,
			string(consumerConfig.exchange.Type),
			true,
			false,
			false,
			false,
			nil).Return(nil)

		stubChannel.EXPECT().ExchangeDeclare(consumerConfig.exchange.RetryNow,
			string(consumerConfig.exchange.Type),
			true,
			false,
			false,
			false,
			nil).Return(nil)

		args := make(map[string]interface{})
		args["x-dead-letter-exchange"] = consumerConfig.exchange.RetryNow
		args["x-message-ttl"] = consumerConfig.queue.RequeueTTL
		args["x-dead-letter-routing-key"] = "#"

		stubChannel.EXPECT().QueueDeclare(consumerConfig.queue.RetryLater, // name
			true,  // durable
			false, // delete when usused
			false, // exclusive
			false, // no-wait
			args).Return(amqp.Queue{}, nil)
		//
		stubChannel.EXPECT().QueueBind(consumerConfig.queue.RetryLater, "#", consumerConfig.exchange.RetryLater, false, nil).Return(nil)
		stubChannel.EXPECT().QueueBind(consumerConfig.queue.Name, "#", consumerConfig.exchange.RetryNow, false, nil).Return(nil)

		err := consumerChannels.setUpRetryExchangeWithQueue(stubChannel)

		if err != nil {
			t.Fatal(err)
		}
	})

}
