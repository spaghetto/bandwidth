package run

import (
	"context"
	"fmt"
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

var collectTimeout = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: Namespace,
	Name:      "collect_timeout_seconds",
	Help:      "Seconds after which the job is aborted",
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

type Job func(context.Context) error

const MaxTimeout = 5 * time.Minute

func Every(interval time.Duration, job Job) error {
	timeout := interval - (interval / 10)
	if timeout > MaxTimeout {
		timeout = MaxTimeout
	}

	collectInterval.Set(interval.Seconds())
	collectTimeout.Set(timeout.Seconds())

	log.Printf("Running first collect, timeout is %s", timeout)

	// action runs the actual job
	action := func() bool {
		collectCount.Inc()

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := job(ctx); err != nil {
			errCount.Inc()
			log.Println(err)
			return false
		}

		lastSuccess.Set(float64(time.Now().Unix()))
		return true
	}

	profile := func() (bool, time.Duration) {
		start := time.Now()
		ok := action()
		took := time.Since(start)

		if ok {
			collectDuration.Set(took.Seconds())
		}

		return ok, took
	}

	// try 5x initially to get it working
	success := false
	for i := 0; i <= 5; i++ {
		ok, took := profile()
		success = success || ok

		if success {
			log.Printf("First collect succeeded in %s. Refreshing every %s", took, interval)
			break
		}
	}

	// failed. backoff
	if !success {
		return fmt.Errorf("None of the first 5 runs succeeded. Aborting")
	}

	// did work. wait 1x interval before running on schedule
	time.Sleep(interval)

	// run on schedule
	for {
		ok, took := profile()

		delay := interval - took
		if !ok {
			log.Printf("Retrying in %s", delay)
		}
		time.Sleep(interval - took)
	}
}
