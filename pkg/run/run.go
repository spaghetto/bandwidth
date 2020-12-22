package run

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var Namespace = "bandwidth"

var errCount = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: Namespace,
	Name:      "collect_errors_total",
	Help:      "Total number of failed collections",
})

var collectCount = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: Namespace,
	Name:      "collects_total",
	Help:      "Count of collects",
})

var collectDuration = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: Namespace,
	Name:      "collect_duration_seconds",
	Help:      "Duration of last collection",
})

var lastSuccess = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: Namespace,
	Name:      "collect_last_success_timestamp_seconds",
	Help:      "UNIX timestamp of the last successful collection",
})

var collectInterval = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: Namespace,
	Name:      "collect_interval",
	Help:      "Interval at with the collection results are refreshed",
})

func Every(interval time.Duration, action func() error) {
	initialSuccess := false
	collectInterval.Set(interval.Seconds())

	log.Printf("Running first collect ...")

	for {
		collectCount.Inc()
		start := time.Now()

		if err := action(); err != nil {
			errCount.Inc()
			log.Println(err)
			continue
		}

		lastSuccess.Set(float64(time.Now().Unix()))

		if !initialSuccess {
			initialSuccess = true
			log.Printf("First collect succeeded. Refreshing every %s", interval)
		}

		took := time.Since(start)
		collectDuration.Set(took.Seconds())

		time.Sleep(interval - took)
	}

}
