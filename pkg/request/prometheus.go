package request

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	successCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "request_success",
		Help: "Count successful requests.",
	})
	error400Counter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "request_error4XX",
		Help: "Count errors 400.",
	})
	error500Counter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "request_error5XX",
		Help: "Count errors 500.",
	})
)

func ResponseMetrics(statusCode int) {
	if statusCode < 400 {
		successCounter.Inc()
	} else if statusCode < 500 {
		error400Counter.Inc()
	} else {
		error500Counter.Inc()
	}
}
