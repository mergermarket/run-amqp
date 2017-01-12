package main

import (
	"fmt"
	"time"
	"log"
	"github.com/mergermarket/run-amqp"
)

const numberOfWorkers = 3
const exchangeName = "test-example-exchange"
var noPatterns = []string{""}

func main() {
	forever := make(chan bool)

	fmt.Println("Run amqp test bench")

	config := runamqp.NewConsumerConfig(
		"amqp://guest:guest@rabbitmq:5672/",
		exchangeName,
		runamqp.Fanout,
		"test-example-queue",
		noPatterns,
		&logger{},
		500,
		5,
		"sample-app",
	)

	consumer := runamqp.NewConsumer(config)

	select {
	case <-consumer.QueuesBound:
		log.Println("Connected!")
	case <-time.After(10 * time.Second):
		log.Fatal("Timed out waiting to set up rabbit")
	}

	handler := &SampleHandler{}

	consumer.Process(handler, numberOfWorkers)

	publisherConfig := runamqp.NewPublisherConfig(config.URL, exchangeName, runamqp.Fanout, &logger{})
	publish, ready := runamqp.NewPublisher(publisherConfig)

	select {
	case <-ready:
	case <-time.After(10 * time.Second):
		log.Fatal("Timed out waiting to set up rabbit")
	}

	publish([]byte("1"), "")
	publish([]byte("2"), "")
	publish([]byte("3"), "")

	<-forever
}

type logger struct {}

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
	log.Println("Got message", string(msg.Body()))
}
