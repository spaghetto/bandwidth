package main

import (
	"context"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/go-clix/cli"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "bandwidth"

const (
	TypeMeasured = "measured"
	TypeExpected = "expected"
)

var (
	transmitRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "transmit_rate_bits",
		Help:      "Transmit (upload) rate in bits per second",
	}, []string{"type"})

	receiveRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "receive_rate_bits",
		Help:      "Receive (upload) rate in bits per second",
	}, []string{"type"})

	latency = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "latency_seconds",
		Help:      "Measured latency (ping) in seconds",
	})

	packetLoss = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "packet_loss",
		Help:      "Packet loss in percent",
	})

	info = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "info",
		Help:      "Metadata gathered during the Speedtest",
	}, []string{"isp", "externalIP"})
)

func main() {
	cmd := &cli.Command{
		Use: "bandwidth_exporter [flags]",
	}

	interval := cmd.Flags().Duration("interval", 15*time.Minute, "Time between measurements. Be aware of network load!")
	listen := cmd.Flags().String("listen", ":9516", "Network address to bind http server to")

	expectDownload := cmd.Flags().Float64("expect-download", 0, "Expected download rate in bits/s")
	expectUpload := cmd.Flags().Float64("expect-upload", 0, "Expected upload rate in bits/s")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		if *expectDownload != 0 {
			receiveRate.WithLabelValues(TypeExpected).Set(*expectDownload)
		}
		if *expectUpload != 0 {
			transmitRate.WithLabelValues(TypeExpected).Set(*expectUpload)
		}

		collectInterval.Set(interval.Seconds())
		go run(*interval)

		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(*listen, nil); err != nil {
			return err
		}

		return nil
	}

	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

const timeoutMax = 5 * time.Minute

func run(interval time.Duration) {
	timeout := interval - (interval / 10)
	if timeout > timeoutMax {
		timeout = timeoutMax
	}
	collectTimeout.Set(timeout.Seconds())

	first := true
	log.Printf("Running first collect, timeout is %s", timeout)

	for ; true; <-time.Tick(interval) {
		start := time.Now()
		collectCount.Inc()

		// try at most 3 times
		var err error
		for i := 0; i < 3; i++ {
			err = measure(timeout)
			if err == nil {
				break
			}
			log.Printf("Speedtest failed, trying %d more times", 3-i)
		}

		if err != nil {
			if first {
				log.Fatalln(err)
			}

			log.Println(err)
			errCount.Inc()
			continue
		}

		took := time.Since(start)
		collectDuration.Set(took.Seconds())

		if first {
			first = false
			log.Printf("First collect succeeded in %s. Refreshing every %s", took, interval)
		}

		lastSuccess.Set(float64(time.Now().Unix()))
	}
}

func measure(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var (
		transmitRate = transmitRate.WithLabelValues(TypeMeasured)
		receiveRate  = receiveRate.WithLabelValues(TypeMeasured)
	)

	r, err := Test(ctx)
	if err != nil {
		clear(transmitRate, receiveRate, latency, packetLoss)
		return err
	}

	transmitRate.Set(r.Upload.Bandwidth * 8)
	receiveRate.Set(r.Download.Bandwidth * 8)
	latency.Set(r.Ping.Latency)
	packetLoss.Set(r.PacketLoss)

	info.Reset()
	info.With(prometheus.Labels{
		"isp":        r.ISP,
		"externalIP": r.Iface.ExternalIP,
	}).Set(1)

	return nil
}

func clear(gauges ...prometheus.Gauge) {
	for _, g := range gauges {
		g.Set(math.NaN())
	}
}
