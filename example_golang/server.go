package main

import (
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"
)

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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
