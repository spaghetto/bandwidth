package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	errCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "collect_errors_total",
		Help:      "Total number of failed collections",
	})

	collectCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "collects_total",
		Help:      "Count of collects",
	})

	collectDuration = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "collect_duration_seconds",
		Help:      "Duration of last collection",
	})

	collectTimeout = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "collect_timeout_seconds",
		Help:      "Seconds after which the job is aborted",
	})

	lastSuccess = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "collect_last_success_timestamp_seconds",
		Help:      "UNIX timestamp of the last successful collection",
	})

	collectInterval = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "collect_interval",
		Help:      "Interval at with the collection results are refreshed",
	})
)
