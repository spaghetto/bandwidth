package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

func New() *Collector {
	return &Collector{}
}

func NewRegister() *Collector {
	c := &Collector{}
	prometheus.MustRegister(c)
	return c
}

type Collector struct {
	mut     sync.Mutex
	metrics Metrics
}

type Metrics []prometheus.Metric

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mut.Lock()

	for _, m := range c.metrics {
		ch <- m
	}

	defer c.mut.Unlock()
}

func (c *Collector) Set(metrics ...prometheus.Metric) {
	c.mut.Lock()
	c.metrics = metrics
	c.mut.Unlock()
}

func (c *Collector) Clear() {
	c.Set()
}

type GaugeOpts struct {
	Name      string
	Namespace string
	Help      string

	// Constant labels
	ConstLabels prometheus.Labels
	// Keys for scrape-time labels
	VariableLabels []string
}

func Gauge(opts GaugeOpts, val float64, labelValues ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(opts.Namespace+"_"+opts.Name, opts.Help, opts.VariableLabels, opts.ConstLabels),
		prometheus.GaugeValue,
		val,
		labelValues...,
	)
}

func Info(opts GaugeOpts, labels prometheus.Labels) prometheus.Metric {
	opts.ConstLabels = labels
	return Gauge(opts, 1)
}
