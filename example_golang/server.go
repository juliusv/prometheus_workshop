package main

import (
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	namespace = "codelab"
	subsystem = "api"

	requestHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "A histogram of the API HTTP request durations in seconds.",
			Buckets:   prometheus.ExponentialBuckets(0.0001, 1.5, 25),
		},
		[]string{"method", "path", "status"},
	)
	requestsInProgress = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "http_requests_in_progress",
			Help:      "The current number of API HTTP requests in progress.",
		})
)

func init() {
	prometheus.MustRegister(requestHistogram)
	prometheus.MustRegister(requestsInProgress)
}

type responseOpts struct {
	baseLatency time.Duration
	errorRatio  float64

	// Whenever 10*outageDuration has passed, an outage will be simulated
	// that lasts for outageDuration. During the outage, errorRatio is
	// increased by a factor of 10, and baseLatency by a factor of 3.  At
	// start-up time, an outage is simulated, too (so that you can see the
	// effects right ahead and don't have to wait for 10*outageDuration).
	outageDuration time.Duration
}

var opts = map[string]map[string]responseOpts{
	"/api/foo": map[string]responseOpts{
		"GET": responseOpts{
			baseLatency:    10 * time.Millisecond,
			errorRatio:     0.005,
			outageDuration: 23 * time.Second,
		},
		"POST": responseOpts{
			baseLatency:    20 * time.Millisecond,
			errorRatio:     0.02,
			outageDuration: time.Minute,
		},
	},
	"/api/bar": map[string]responseOpts{
		"GET": responseOpts{
			baseLatency:    15 * time.Millisecond,
			errorRatio:     0.0025,
			outageDuration: 13 * time.Second,
		},
		"POST": responseOpts{
			baseLatency:    50 * time.Millisecond,
			errorRatio:     0.01,
			outageDuration: 47 * time.Second,
		},
	},
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	begun := time.Now()
	requestsInProgress.Inc()
	status := http.StatusOK

	defer func() {
		requestsInProgress.Dec()
		requestHistogram.With(prometheus.Labels{
			"method": r.Method,
			"path":   r.URL.Path,
			"status": fmt.Sprint(status),
		}).Observe(time.Since(begun).Seconds())
	}()

	pathOpts, ok := opts[r.URL.Path]
	if !ok {
		status = http.StatusNotFound
		http.Error(w, fmt.Sprintf("Path %q not found.", r.URL.Path), status)
		return
	}
	methodOpts, ok := pathOpts[r.Method]
	if !ok {
		status = http.StatusMethodNotAllowed
		http.Error(w, fmt.Sprintf("Method %q not allowed.", r.Method), status)
		return
	}

	latencyFactor := time.Duration(1)
	errorFactor := 1.
	if time.Since(start)%(10*methodOpts.outageDuration) < methodOpts.outageDuration {
		latencyFactor *= 3
		errorFactor *= 10
	}
	time.Sleep(
		(methodOpts.baseLatency + time.Duration(rand.NormFloat64()*float64(methodOpts.baseLatency)/10)) * latencyFactor,
	)
	if rand.Float64() <= methodOpts.errorRatio*errorFactor {
		status = http.StatusInternalServerError
		http.Error(w, "Fake error to test monitoring.", status)
	}
}
