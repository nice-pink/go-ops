package metric

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Listen(port int) {
	http.Handle("/metrics", promhttp.Handler())
	portString := ":" + strconv.Itoa(port)
	http.ListenAndServe(portString, nil)
}
