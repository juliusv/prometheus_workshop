package main

import (
	"math/rand"
	"net/http"
	"time"
)

type responseOpts struct {
	baseLatency time.Duration
	errorRatio  float64
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	opts := map[string]map[string]responseOpts{
		"/api/foo": map[string]responseOpts{
			"GET": responseOpts{
				baseLatency: time.Millisecond,
				errorRatio:  0.5,
			},
			"POST": responseOpts{
				baseLatency: 100 * time.Millisecond,
				errorRatio:  2,
			},
		},
		"/api/bar": map[string]responseOpts{
			"GET": responseOpts{
				baseLatency: 5 * time.Millisecond,
				errorRatio:  .25,
			},
			"POST": responseOpts{
				baseLatency: 50 * time.Millisecond,
				errorRatio:  1,
			},
		},
	}

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

	time.Sleep(methodOpts.baseLatency + time.Duration(rand.NormFloat64()*200)*time.Millisecond)
	if rand.Float64()*100 <= methodOpts.errorRatio {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
