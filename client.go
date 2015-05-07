package main

import (
	"bytes"
	"flag"
	"math"
	"net/http"
	"time"
)

var (
	oscillationPeriod = flag.Duration("oscillation-period", 5*time.Minute, "The duration of the rate oscillation period.")
)

func startClient(servAddr string) {
	start := time.Now()

	oscillationFactor := func() float64 {
		return 2 + math.Sin(math.Sin(2*math.Pi*float64(time.Since(start))/float64(*oscillationPeriod)))
	}

	// GET /api/foo.
	go func() {
		for {
			http.Get("http://" + servAddr + "/api/foo")
			time.Sleep(time.Duration(10*oscillationFactor()) * time.Millisecond)
		}
	}()
	// POST /api/foo.
	go func() {
		for {
			http.Post("http://"+servAddr+"/api/foo", "text/plain", &bytes.Buffer{})
			time.Sleep(time.Duration(150*oscillationFactor()) * time.Millisecond)
		}
	}()
	// GET /api/bar.
	go func() {
		for {
			http.Get("http://" + servAddr + "/api/bar")
			time.Sleep(time.Duration(20*oscillationFactor()) * time.Millisecond)
		}
	}()
	// POST /api/bar.
	go func() {
		for {
			http.Post("http://"+servAddr+"/api/bar", "text/plain", &bytes.Buffer{})
			time.Sleep(time.Duration(100*oscillationFactor()) * time.Millisecond)
		}
	}()
	// GET /api/nonexistent.
	go func() {
		for {
			http.Get("http://" + servAddr + "/api/nonexistent")
			time.Sleep(time.Duration(500*oscillationFactor()) * time.Millisecond)
		}
	}()
}
