package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/justinas/alice"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/streadway/handy/report"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()

	http.Handle("/metrics", prometheus.Handler())
	http.HandleFunc("/api/", handleInstrumentedAPI)

	// Log every received HTTP request to stdout.
	go http.ListenAndServe(*addr, alice.New(
		report.JSONMiddleware(os.Stdout),
	).Then(http.DefaultServeMux))

	startClient(*addr)

	select {}
}
