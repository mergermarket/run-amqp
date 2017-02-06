package runamqp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

type stubPublisher struct {
	ready                    bool
	publishCalled            bool
	publishCalledWithMessage string
	publishCalleWithPattern  string
	publishCalledWithOptions PublishOptions
	err                      error
}

func (s *stubPublisher) IsReady() bool {
	return s.ready
}

func (s *stubPublisher) Publish(message []byte, pattern string) error {
	s.publishCalled = true
	s.publishCalledWithMessage = string(message)
	s.publishCalleWithPattern = pattern
	return s.err
}

func (s *stubPublisher) PublishWithOptions(message []byte, pattern string, options PublishOptions) error {
	s.publishCalled = true
	s.publishCalledWithMessage = string(message)
	s.publishCalleWithPattern = pattern
	s.publishCalledWithOptions = options
	return s.err
}

const testExchangeName = "experts exchange"

func TestPublisherServer_ServeHTTP(t *testing.T) {
	t.Run("/up should return 503 when NOT ready", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = false

		publisherServer := newPublisherServer(publisher, testExchangeName, &testLogger{t})

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

		publisherServer := newPublisherServer(publisher, testExchangeName, &testLogger{t})

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/up", nil)
		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Error("expected", http.StatusOK, "but got", w.Code)
		}

	})

	t.Run("/entry should return 503 when NOT ready", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = false

		publisherServer := newPublisherServer(publisher, testExchangeName, &testLogger{t})

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

		publisherServer := newPublisherServer(publisher, testExchangeName, &testLogger{t})

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/entry", nil)
		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Error("expected", http.StatusOK, "but got", w.Code)
		}

		//if w.Body.String() != entryBody {
		//	t.Error("expected", entryBody, "but got", w.Body.String())
		//}

	})

	t.Run("/entry should return 405 on DELETE and publisher ready", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = true

		publisherServer := newPublisherServer(publisher, testExchangeName, &testLogger{t})

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

		publisherServer := newPublisherServer(publisher, testExchangeName, &testLogger{t})

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodPost, "/entry", strings.NewReader("some string"))
		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Error("expected", http.StatusInternalServerError, "but got", w.Code)
		}

	})

	t.Run("/entry should return 200 on POST when publishing succeeds", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = true

		publisherServer := newPublisherServer(publisher, testExchangeName, &testLogger{t})

		w := httptest.NewRecorder()

		message := "some string"
		pattern := "pattern"

		r, _ := http.NewRequest(http.MethodPost, "/entry", strings.NewReader(message))

		q := r.URL.Query()
		q.Add("pattern", pattern)
		r.URL.RawQuery = q.Encode()

		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Error("expected", http.StatusOK, "but got", w.Code)
		}

		if !publisher.publishCalled {
			t.Error("publisher.Publish should have been called but it was not")
		}

		if publisher.publishCalledWithMessage != message {
			t.Error("publisher.Publish should have been called with", message, "but it was called with", publisher.publishCalledWithMessage)
		}

		if publisher.publishCalleWithPattern != pattern {
			t.Error("publisher.Publish should have been called with", pattern, "but it was called with", publisher.publishCalleWithPattern)
		}

	})

	t.Run("/entry should return 200 on POST when submitting form", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = true

		publisherServer := newPublisherServer(publisher, testExchangeName, &testLogger{t})

		w := httptest.NewRecorder()

		message := "some string"
		pattern := "pattern"

		form := url.Values{}
		form.Add("pattern", pattern)
		form.Add("message", message)

		r, _ := http.NewRequest(http.MethodPost, "/entry", strings.NewReader(form.Encode()))

		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Error("expected", http.StatusOK, "but got", w.Code)
		}

		if !publisher.publishCalled {
			t.Error("publisher.Publish should have been called but it was not")
		}

		if publisher.publishCalledWithMessage != message {
			t.Error("publisher.Publish should have been called with", message, "but it was called with", publisher.publishCalledWithMessage)
		}

		if publisher.publishCalleWithPattern != pattern {
			t.Error("publisher.Publish should have been called with", pattern, "but it was called with", publisher.publishCalleWithPattern)
		}

	})

	t.Run("/entry should return 200 on POST when submitting form with PublishOptions", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = true

		publisherServer := newPublisherServer(publisher, testExchangeName, &testLogger{t})

		w := httptest.NewRecorder()

		message := "some string"
		pattern := "pattern"
		var priority uint8 = 2

		form := url.Values{}
		form.Add("pattern", pattern)
		form.Add("message", message)
		form.Add("priority", strconv.Itoa(int(priority)))

		r, _ := http.NewRequest(http.MethodPost, "/entrywithoptions", strings.NewReader(form.Encode()))

		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		publisherServer.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Error("expected", http.StatusOK, "but got", w.Code)
		}

		if !publisher.publishCalled {
			t.Error("publisher.Publish should have been called but it was not")
		}

		if publisher.publishCalledWithMessage != message {
			t.Error("publisher.Publish should have been called with", message, "but it was called with", publisher.publishCalledWithMessage)
		}

		if publisher.publishCalleWithPattern != pattern {
			t.Error("publisher.Publish should have been called with", pattern, "but it was called with", publisher.publishCalleWithPattern)
		}

		expectedOptions := PublishOptions{Priority: priority}
		if publisher.publishCalledWithOptions != expectedOptions {
			t.Error("publisher.Publish should have been called with", priority, "but it was called with", publisher.publishCalledWithOptions)
		}
	})

	// t.Run("/entry should return 200 on POST when publishing with options succeeds", func(t *testing.T) {

	// 	publisher := new(stubPublisher)

	// 	publisherServer := newPublisherServer(publisher, testExchangeName, &testLogger{t})

	// 	w := httptest.NewRecorder()

	// 	message := "some string"
	// 	pattern := "pattern"
	// 	options := &PublishOptions{Priority: 3}

	// 	body := []byte()

	// 	r, _ := http.NewRequest(http.MethodPost, "/entrywithoptions", bytes.NewBuffer(body))

	// 	fmt.Println("hereee", r)
	// 	q := r.URL.Query()
	// 	q.Add("pattern", pattern)
	// 	r.URL.RawQuery = q.Encode()

	// 	publisherServer.ServeHTTP(w, r)

	// 	if w.Code != http.StatusOK {
	// 		t.Error("expected", http.StatusOK, "but got", w.Code)
	// 	}

	// 	if !publisher.publishCalled {
	// 		t.Error("publisher.Publish should have been called but it was not")
	// 	}

	// 	if publisher.publishCalledWithMessage != message {
	// 		t.Error("publisher.Publish should have been called with", message, "but it was called with", publisher.publishCalledWithMessage)
	// 	}

	// 	if publisher.publishCalleWithPattern != pattern {
	// 		t.Error("publisher.Publish should have been called with", pattern, "but it was called with", publisher.publishCalleWithPattern)
	// 	}

	// })
}
