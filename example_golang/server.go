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
