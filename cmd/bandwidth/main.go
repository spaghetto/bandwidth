package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-clix/cli"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sh0rez/bandwidth_exporter/pkg/run"
)

var transmitRate = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: "bandwidth",
	Name:      "transmit_rate_bits",
	Help:      "Measured transmit (upload) rate in bits per second",
})

var receiveRate = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: "bandwidth",
	Name:      "receive_rate_bits",
	Help:      "Measured receive (download) rate in bits per second",
})

var latency = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: "bandwidth",
	Name:      "latency_seconds",
	Help:      "Measured latency (ping) in seconds",
})

var info = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "bandwidth",
	Name:      "info",
	Help:      "Metadata gathered during the Speedtest",
}, []string{"isp"})

func main() {
	cmd := &cli.Command{
		Use: "speed-exporter",
	}

	interval := cmd.Flags().Duration("interval", time.Minute*30, "Time between measurements. Be aware of network load!")
	serverID := cmd.Flags().String("server-id", "30593", "Speedtest.net server ID")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		go run.Every(*interval, func() error {
			result, err := Test(*serverID)
			if err != nil {
				return err
			}

			transmitRate.Set(result.Upload.Bandwidth * 8)
			receiveRate.Set(result.Download.Bandwidth * 8)
			latency.Set(result.Ping.Latency)

			info.WithLabelValues(result.ISP).Set(1)

			return nil
		})

		log.Println("Listening on :2112")
		return http.ListenAndServe(":2112", promhttp.Handler())
	}

	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
