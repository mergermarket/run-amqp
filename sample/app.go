package main

import (
	"fmt"
	"github.com/mergermarket/run-amqp"
	"log"
	"net/http"
	"time"
)

const numberOfWorkers = 3
const exchangeName = "producer-messages"
const requeueTTL = 500
const requeueLimit = 5
const serviceName = "sample-app"
const amqpURL = "amqp://guest:guest@rabbitmq:5672/"

var noPatterns = []string{""}

func main() {

	consumerConfig := runamqp.NewConsumerConfig(
		amqpURL,
		exchangeName,
		runamqp.Fanout,
		noPatterns,
		&logger{},
		requeueTTL,
		requeueLimit,
		serviceName,
	)

	consumer := runamqp.NewConsumer(consumerConfig)

	select {
	case <-consumer.QueuesBound:
		log.Println("Waiting for messages")
	case <-time.After(10 * time.Second):
		log.Fatal("Timed out waiting to set up rabbit")
	}

	handler := &SampleHandler{}

	consumer.Process(handler, numberOfWorkers)

	publisher := runamqp.NewPublisher(consumerConfig.NewPublisherConfig())

	select {
	case <-publisher.PublishReady:
		log.Println("Publisher ready")
	case <-time.After(10 * time.Second):
		log.Fatal("Timed out waiting to set up rabbit")
	}

	publisher.Publish([]byte("1"), "")
	publisher.Publish([]byte("2"), "")
	publisher.Publish([]byte("3"), "")

	log.Println("Listening on 8080, POST /entry {some body} to publish to the exchange or GET /up to see if rabbit is ready")

	err := http.ListenAndServe(":8080", publisher)

	if err != nil {
		log.Fatal(err)
	}
}

type logger struct{}

func (*logger) Info(x ...interface{}) {
	log.Println(x)
}

func (*logger) Error(x ...interface{}) {
	log.Println(x)
}

func (*logger) Debug(x ...interface{}) {
	log.Println(x)
}

type SampleHandler struct {
}

func (*SampleHandler) Name() string {
	return "Sample consumer"
}

func (*SampleHandler) Handle(msg runamqp.Message) {
	log.Println("Sample handler got message", string(msg.Body()))
	msg.Ack()
}
