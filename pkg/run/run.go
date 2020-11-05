package run

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var errCount = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "bandwidth",
	Name:      "collect_errors_total",
	Help:      "Total number of failed collections",
})

var collectCount = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "bandwidth",
	Name:      "collects_total",
	Help:      "Count of collects",
})

var collectDuration = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: "bandwidth",
	Name:      "collect_duration_seconds",
	Help:      "Duration of last collection",
})

func Every(interval time.Duration, action func() error) {
	initialSuccess := false

	for {
		collectCount.Inc()
		start := time.Now()

		if err := action(); err != nil {
			errCount.Inc()
			log.Println(err)
		}

		if !initialSuccess {
			initialSuccess = true
			log.Println("First collect succeeded. This exporter is ready to go")
		}

		took := time.Since(start)
		collectDuration.Set(took.Seconds())

		time.Sleep(interval - took)
	}

}
