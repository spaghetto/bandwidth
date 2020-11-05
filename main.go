package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-clix/cli"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var errCount = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "bandwidth",
	Name:      "collect_errors_total",
	Help:      "Total number of failed collections (speedtests)",
})

var testCount = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "bandwidth",
	Name:      "collects_total",
	Help:      "Count of collects (speedtest)",
})

var testDuration = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: "bandwidth",
	Name:      "collect_duration_seconds",
	Help:      "Duration of last collection (speedtest)",
})

var transmitRate = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: "bandwidth",
	Name:      "transmit_rate_bytes",
	Help:      "Measured transmit (upload) rate in bytes per second",
})

var receiveRate = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: "bandwidth",
	Name:      "receive_rate_bytes",
	Help:      "Measured receive (download) rate in bytes per second",
})

var latency = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: "bandwidth",
	Name:      "latency_seconds",
	Help:      "Measured latency (ping) in seconds",
})

func main() {
	cmd := &cli.Command{
		Use: "speed-exporter",
	}

	interval := cmd.Flags().Duration("interval", time.Minute*30, "Time between measurements. Be aware of network load!")
	serverID := cmd.Flags().String("server-id", "30593", "Speedtest.net server ID")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		go func() {
			first := true

			for {
				testCount.Inc()
				start := time.Now()

				result, err := Test(*serverID)
				if err != nil {
					errCount.Inc()
					log.Println(err)
				} else {
					transmitRate.Set(result.Upload)
					receiveRate.Set(result.Download)
					latency.Set(result.Ping)
				}

				took := time.Since(start)

				if first {
					log.Printf("Initial measurement suceeded in %s", took)
					first = false
				}

				testDuration.Set(took.Seconds())
				time.Sleep(*interval - took)
			}
		}()

		log.Println("Listening on :2112")
		return http.ListenAndServe(":2112", promhttp.Handler())
	}

	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
