package runamqp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/mergermarket/run-amqp/helpers"
)

type stubPublisher struct {
	ready                    bool
	publishCalled            bool
	publishCalledWithMessage string
	publishCalledWithOptions PublishOptions
	err                      error
}

func (s *stubPublisher) IsReady() bool {
	return s.ready
}

func (s *stubPublisher) Publish(message []byte, options PublishOptions) error {
	s.publishCalled = true
	s.publishCalledWithMessage = string(message)
	s.publishCalledWithOptions = options
	return s.err
}

const testExchangeName = "experts exchange"

func TestPublisherServer_ServeHTTP(t *testing.T) {
	logger := helpers.NewTestLogger(t)

	t.Run("/up should return 503 when NOT ready", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = false

		publisherServer := newPublisherServer(publisher, testExchangeName, logger)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/up", nil)
		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusServiceUnavailable {
			t.Error("expected", http.StatusServiceUnavailable, "but got", w.Code)
		}

	})

	t.Run("/up should return 200 when ready", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = true

		publisherServer := newPublisherServer(publisher, testExchangeName, logger)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/up", nil)
		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Error("expected", http.StatusOK, "but got", w.Code)
		}

	})

}

func TestPublisherServerEntry_ServeHTTP(t *testing.T) {
	logger := helpers.NewTestLogger(t)

	t.Run("/entry should return 503 when NOT ready", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = false

		publisherServer := newPublisherServer(publisher, testExchangeName, logger)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodPost, "/entry", nil)
		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusServiceUnavailable {
			t.Error("expected", http.StatusServiceUnavailable, "but got", w.Code)
		}

	})

	t.Run("/entry should return 200 on GET when ready", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = true

		publisherServer := newPublisherServer(publisher, testExchangeName, logger)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/entry", nil)
		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Error("expected", http.StatusOK, "but got", w.Code)
		}

	})

	t.Run("/entry should return 405 on DELETE and publisher ready", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = true

		publisherServer := newPublisherServer(publisher, testExchangeName, logger)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodDelete, "/entry", strings.NewReader("some string"))
		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusMethodNotAllowed {
			t.Error("expected", http.StatusMethodNotAllowed, "but got", w.Code)
		}

	})

	t.Run("/entry should return 500 on POST when publishing fails", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = true
		publisher.err = fmt.Errorf("This should be returned on publisher.Publish")

		publisherServer := newPublisherServer(publisher, testExchangeName, logger)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodPost, "/entry", strings.NewReader("some string"))
		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Error("expected", http.StatusInternalServerError, "but got", w.Code)
		}

	})

	t.Run("/entry should return 200 on POST when submitting form with PublishOptions", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = true

		publisherServer := newPublisherServer(publisher, testExchangeName, logger)

		w := httptest.NewRecorder()

		message := "some string"
		pattern := "pattern"
		var priority uint8 = 2
		publishToQueue := ""

		form := url.Values{}
		form.Add("pattern", pattern)
		form.Add("message", message)
		form.Add("priority", strconv.Itoa(int(priority)))
		form.Add("publishToQueue", publishToQueue)

		r, _ := http.NewRequest(http.MethodPost, "/entry", strings.NewReader(form.Encode()))

		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Error("expected", http.StatusOK, "but got", w.Code)
		}

		if !publisher.publishCalled {
			t.Error("publisher.PublishWithOptions should have been called but it was not")
		}

		if publisher.publishCalledWithMessage != message {
			t.Error("publisher.PublishWithOptions should have been called with", message, "but it was called with", publisher.publishCalledWithMessage)
		}

		expectedOptions := PublishOptions{Priority: priority, Pattern: pattern, PublishToQueue: publishToQueue}
		if publisher.publishCalledWithOptions != expectedOptions {
			t.Error("publisher.PublishWithOptions should have been called with", expectedOptions, "but it was called with", publisher.publishCalledWithOptions)
		}

	})

	t.Run("/entry should return 200 on POST when publishing with options succeeds", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = true

		publisherServer := newPublisherServer(publisher, testExchangeName, logger)

		w := httptest.NewRecorder()

		message := "some string"
		pattern := "pattern"
		var priority uint8 = 2
		publishToQueue := ""

		r, _ := http.NewRequest(http.MethodPost, "/entry", strings.NewReader(message))

		q := r.URL.Query()
		q.Add("pattern", pattern)
		q.Add("priority", strconv.Itoa(int(priority)))
		q.Add("publishToQueue", publishToQueue)
		r.URL.RawQuery = q.Encode()

		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Error("expected", http.StatusOK, "but got", w.Code)
		}

		if !publisher.publishCalled {
			t.Error("publisher.PublishWithOptions should have been called but it was not")
		}

		if publisher.publishCalledWithMessage != message {
			t.Error("publisher.PublishWithOptions should have been called with", message, "but it was called with", publisher.publishCalledWithMessage)
		}

		expectedOptions := PublishOptions{Priority: priority, Pattern: pattern, PublishToQueue: publishToQueue}
		if publisher.publishCalledWithOptions != expectedOptions {
			t.Error("publisher.PublishWithOptions should have been called with", expectedOptions, "but it was called with", publisher.publishCalledWithOptions)
		}

	})

}
