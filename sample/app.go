package main

import (
	"fmt"
	"github.com/mergermarket/run-amqp"
	"log"
	"net/http"
	"time"
)

const numberOfWorkers = 3
const exchangeName = "test-example-exchange"

var noPatterns = []string{""}

func main() {

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
	publisher := runamqp.NewPublisher(publisherConfig)

	select {
	case <-publisher.PublishReady:
	case <-time.After(10 * time.Second):
		log.Fatal("Timed out waiting to set up rabbit")
	}

	publisher.Publish([]byte("1"), "")
	publisher.Publish([]byte("2"), "")
	publisher.Publish([]byte("3"), "")

	svr := httpPublisher{publisher}

	log.Println("Listening on 8080, hit it on / to publish a message and see what happens")

	err := http.ListenAndServe(":8080", &svr)

	if err != nil {
		log.Fatal(err)
	}
}

type httpPublisher struct {
	p *runamqp.Publisher
}

func (h *httpPublisher) ServeHTTP(http.ResponseWriter, *http.Request) {
	log.Println("Going to try and publish....")
	h.p.Publish([]byte("Pokey"), "")
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
	log.Println("Got message", string(msg.Body()))
}
