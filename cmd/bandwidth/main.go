package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-clix/cli"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sh0rez/bandwidth_exporter/pkg/metrics"
	"github.com/sh0rez/bandwidth_exporter/pkg/run"
)

var (
	TypeMeasured = "measured"
	TypeExpected = "expected"
)

var transmitRate = metrics.GaugeOpts{
	Namespace:      "bandwidth",
	Name:           "transmit_rate_bits",
	Help:           "Transmit (upload) rate in bits per second",
	VariableLabels: []string{"type"},
}

var receiveRate = metrics.GaugeOpts{
	Namespace:      "bandwidth",
	Name:           "receive_rate_bits",
	Help:           "Receive (download) rate in bits per second",
	VariableLabels: []string{"type"},
}

var latency = metrics.GaugeOpts{
	Namespace: "bandwidth",
	Name:      "latency_seconds",
	Help:      "Measured latency (ping) in seconds",
}

var packetLoss = metrics.GaugeOpts{
	Namespace: "bandwidth",
	Name:      "packet_loss",
	Help:      "Packet loss in percent",
}

var info = metrics.GaugeOpts{
	Namespace: "bandwidth",
	Name:      "info",
	Help:      "Metadata gathered during the Speedtest",
}

func main() {
	cmd := &cli.Command{
		Use: "speed-exporter",
	}

	interval := cmd.Flags().Duration("interval", time.Minute*30, "Time between measurements. Be aware of network load!")
	serverID := cmd.Flags().String("server-id", "30593", "Speedtest.net server ID")

	expectDownload := cmd.Flags().Float64("expect-download", 0, "Expected download rate in bits/s")
	expectUpload := cmd.Flags().Float64("expect-upload", 0, "Expected upload rate in bits/s")

	c := metrics.NewRegister()

	cmd.Run = func(cmd *cli.Command, args []string) error {
		go run.Every(*interval, func() error {
			result, err := Test(*serverID)
			if err != nil {
				c.Clear()
				return err
			}

			var m = metrics.Metrics{
				metrics.Gauge(transmitRate, result.Upload.Bandwidth*8, TypeMeasured),
				metrics.Gauge(receiveRate, result.Download.Bandwidth*8, TypeMeasured),

				metrics.Gauge(latency, result.Ping.Latency),
				metrics.Gauge(packetLoss, result.PacketLoss),

				metrics.Info(info, prometheus.Labels{
					"isp":        result.ISP,
					"externalIP": result.Iface.ExternalIP,
				}),
			}

			if *expectDownload != 0 {
				m = append(m, metrics.Gauge(receiveRate, *expectDownload, TypeExpected))
			}
			if *expectUpload != 0 {
				m = append(m, metrics.Gauge(transmitRate, *expectUpload, TypeExpected))
			}

			c.Set(m...)
			return nil
		})

		log.Println("Listening on :2112")
		return http.ListenAndServe(":2112", promhttp.Handler())
	}

	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
