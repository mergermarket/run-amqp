package runamqp

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestPublisherServerEntryWithOptions_ServeHTTP(t *testing.T) {
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

	t.Run("/entry should return 200 on POST when publishing with options succeeds", func(t *testing.T) {

		publisher := new(stubPublisher)
		publisher.ready = true

		publisherServer := newPublisherServer(publisher, testExchangeName, &testLogger{t})

		w := httptest.NewRecorder()

		message := "some string"
		pattern := "pattern"
		var priority uint8 = 2

		r, _ := http.NewRequest(http.MethodPost, "/entrywithoptions", strings.NewReader(message))

		q := r.URL.Query()
		q.Add("pattern", pattern)
		q.Add("priority", strconv.Itoa(int(priority)))
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

		expectedOptions := PublishOptions{Priority: priority}
		if publisher.publishCalledWithOptions != expectedOptions {
			t.Error("publisher.Publish should have been called with", priority, "but it was called with", publisher.publishCalledWithOptions)
		}

	})
}
