package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-clix/cli"
	"github.com/go-rod/rod"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sh0rez/bandwidth_exporter/pkg/run"
)

const Namespace = "pageload"

var loadDuration = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: Namespace,
	Name:      "duration_seconds",
	Help:      "Seconds the pageload took",
})

func main() {
	run.Namespace = Namespace

	cmd := &cli.Command{
		Use: "pageload-exporter",
	}

	interval := cmd.Flags().Duration("interval", time.Second*10, "Time between measurements. Be aware of network load!")
	url := cmd.Flags().String("url", "https://google.de", "URL to load")

	cmd.Run = func(cmd *cli.Command, args []string) error {
		chrome := rod.New().MustConnect()
		go run.Every(*interval, func() error {
			start := time.Now()
			page := chrome.MustPage(*url).MustWaitLoad()
			loadDuration.Set(time.Since(start).Seconds())

			return page.Close()
		})

		log.Println("Listening on :2112")
		return http.ListenAndServe(":2112", promhttp.Handler())
	}

	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
