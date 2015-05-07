package main

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "codelab"
	subsystem = "api"
)

var (
	requestHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "A summary over the API HTTP request durations in seconds.",
			Buckets:   prometheus.ExponentialBuckets(0.005, 1.2, 15),
		},
		[]string{"method", "path", "status"},
	)
	requestsInProgress = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "http_requests_inprogress",
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
}

var opts = map[string]map[string]responseOpts{
	"/api/foo": map[string]responseOpts{
		"GET": responseOpts{
			baseLatency: 10 * time.Millisecond,
			errorRatio:  0.005,
		},
		"POST": responseOpts{
			baseLatency: 20 * time.Millisecond,
			errorRatio:  0.02,
		},
	},
	"/api/bar": map[string]responseOpts{
		"GET": responseOpts{
			baseLatency: 15 * time.Millisecond,
			errorRatio:  0.0025,
		},
		"POST": responseOpts{
			baseLatency: 50 * time.Millisecond,
			errorRatio:  0.01,
		},
	},
}

type statusLoggingResponseWriter struct {
	status int
	http.ResponseWriter
}

func (w *statusLoggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func handleInstrumentedAPI(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestsInProgress.Inc()

	lw := statusLoggingResponseWriter{http.StatusOK, w}
	handleAPI(&lw, r)

	requestHistogram.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(lw.status)).Observe(float64(time.Since(start)) / float64(time.Second))
	requestsInProgress.Dec()
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	pathOpts, ok := opts[r.URL.Path]
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	methodOpts, ok := pathOpts[r.Method]
	if !ok {
		http.Error(w, "Method not Allowed", http.StatusMethodNotAllowed)
		return
	}

	time.Sleep(methodOpts.baseLatency + time.Duration(rand.NormFloat64()*float64(time.Millisecond))*10)
	if rand.Float64() <= methodOpts.errorRatio {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
