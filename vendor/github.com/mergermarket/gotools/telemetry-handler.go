package tools

import (
	"net/http"
	"time"
)

type logger interface {
	Debug(...interface{})
	Info(...interface{})
	Error(...interface{})
}

// WrapWithTelemetry takes your http.Handler and adds debug logs with request details and marks metrics
func WrapWithTelemetry(routeName string, router http.Handler, logger logger, statsd StatsD) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer timeTrack(routeName, time.Now(), logger, statsd)
		logger.Debug(r.Method, "at", r.URL.String())
		router.ServeHTTP(w, r)

		//todo: look at response code, log and metrics
	})
}

func timeTrack(routeName string, start time.Time, logger logger, statsd StatsD) {
	elapsed := time.Since(start)
	statsd.Histogram("web.response_time", float64(elapsed.Nanoseconds())/1000000, "route:"+routeName)
}
