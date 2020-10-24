package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"qrator/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
)

func healthz(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(response, "ok")
}

func main() {
	var (
		clientID      = kingpin.Flag("qrator.client-id", "Your personal dashboard ID which obtained in dashboard.").Short('c').Default("1").String()
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Short('l').Default(":9805").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path to expose metrics.").Short('p').Default("/metrics").String()
	)

	tokenAuth, isSet := os.LookupEnv("QRATOR_TOKEN_AUTH")
	if !isSet {
		log.Fatalf("environment variable QRATOR_TOKEN_AUTH isn't set.")
	}

	kingpin.Parse()

	collector, err := collector.NewCollector(*clientID, tokenAuth)
	if err != nil {
		log.Fatalln(err)
	}

	prometheus.MustRegister(collector)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Qrator Exporter</head>
			<body>
			<h2>Qrator Exporter</h2>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			<html>`))
	})

	fmt.Printf("Starting Qrator exporter server on address: %s\n", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
