package runamqp

import (
	"github.com/mergermarket/run-amqp/helpers"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
