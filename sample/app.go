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

	log.Println("Listening on 8080, POST /entry {some body} to publish to the exchange")

	router := http.NewServeMux()
	router.Handle("/internal/rabbit/", http.StripPrefix("/internal/rabbit", publisher))
	router.Handle("/brett", brett{})

	err := http.ListenAndServe(":8080", router)

	if err != nil {
		log.Fatal(err)
	}
}

type brett struct{}

func (brett) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hi brett")
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
