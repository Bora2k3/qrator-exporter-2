package collector

import "github.com/prometheus/client_golang/prometheus"

func (c *Collector) metrics() {
	c.up = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "up",
		Help:      "Was the last scrape successfull.",
	})

	c.failedScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_failed_scrpes",
		Help:      "Number of failed qrator scrapes.",
	})

	c.totalScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_total_scrapes",
		Help:      "Number of total qrator scrapes.",
	})

	c.failedDomainScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_failed_domain_scrpes",
		Help:      "Number of failed domain scrapes from Qrator API.",
	})

	c.failedStatisticsScrapes = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "failed_statistics_scrapes",
			Help:      "Number of failed statistics scrapes.",
		},
		[]string{
			"domain",
			"api_method",
		},
	)

	c.failedJSONDecode = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "failed_json_decode",
			Help:      "Number of failed json response decode from API.",
		},
		[]string{
			"domain",
			"api_method",
		},
	)

	c.BandwidthTraffic = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "bandwidth_traffic",
			Help:      "Bandwidth traffic (bps).",
		},
		[]string{
			"domain",
			"state",
			"api_method",
		},
	)

	c.Packets = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "packets",
			Help:      "Packets (pps).",
		},
		[]string{
			"domain",
			"state",
			"api_method",
		},
	)

	c.Blacklist = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "blacklist",
			Help:      "Number of IPs banned by services.",
		},
		[]string{
			"domain",
			"service",
			"api_method",
		},
	)

	c.HTTPRequests = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_requests",
			Help:      "HTTP request rate (rps).",
		},
		[]string{
			"domain",
			"api_method",
		},
	)

	c.HTTPResponses = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_responses",
			Help:      "HTTP responses rate (rsp).",
		},
		[]string{
			"domain",
			"duration",
			"api_method",
		},
	)

	c.HTTPErrors = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_errors",
			Help:      "HTTP errors rate (rsp).",
		},
		[]string{
			"domain",
			"http_code",
			"api_method",
		},
	)
}
