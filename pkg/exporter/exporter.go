package exporter

import (
	"errors"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-resty/resty/v2"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"strings"
)

type Exporter struct {
	address           string
	logger            log.Logger
	up                *prometheus.Desc
	ActiveConnections *prometheus.Desc
	Accepts           *prometheus.Desc
	Handled           *prometheus.Desc
	Requests          *prometheus.Desc
	Reading           *prometheus.Desc
	Writing           *prometheus.Desc
	Waiting           *prometheus.Desc
}

func New(address string, logger log.Logger) *Exporter {
	return &Exporter{
		address: address,
		logger:  logger,
		up: prometheus.NewDesc(
			prometheus.BuildFQName("nginx", "status", "up"),
			"up",
			nil,
			nil,
		),
		ActiveConnections: prometheus.NewDesc(
			prometheus.BuildFQName("nginx", "status", "active_connections"),
			"Active connections",
			nil,
			nil,
		),
		Accepts: prometheus.NewDesc(
			prometheus.BuildFQName("nginx", "status", "accepts"),
			"accepts",
			nil,
			nil,
		),
		Handled: prometheus.NewDesc(
			prometheus.BuildFQName("nginx", "status", "handled"),
			"handled",
			nil,
			nil,
		),
		Requests: prometheus.NewDesc(
			prometheus.BuildFQName("nginx", "status", "requests"),
			"requests",
			nil,
			nil,
		),
		Reading: prometheus.NewDesc(
			prometheus.BuildFQName("nginx", "status", "Reading"),
			"Reading",
			nil,
			nil,
		),
		Writing: prometheus.NewDesc(
			prometheus.BuildFQName("nginx", "status", "Writing"),
			"Writing",
			nil,
			nil,
		),
		Waiting: prometheus.NewDesc(
			prometheus.BuildFQName("nginx", "status", "Waiting"),
			"Waiting",
			nil,
			nil,
		),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.ActiveConnections
	ch <- e.Accepts
	ch <- e.Handled
	ch <- e.Requests
	ch <- e.Reading
	ch <- e.Writing
	ch <- e.Waiting
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	client := resty.New()
	resp, err := client.R().EnableTrace().Get(e.address)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		level.Error(e.logger).Log("msg", "Failed to connect nginx", "err", err)
		return
	}
	up := float64(1)
	if resp.StatusCode() != http.StatusOK {
		level.Error(e.logger).Log("msg", "Failed to collect stats from nginx", "err", errors.New(fmt.Sprintf("status_code: %d", resp.StatusCode())))
		up = 0
	}
	if err := e.parseStats(ch, resp.Body()); err != nil {
		up = 0
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, up)
}

func (e *Exporter) parseStats(ch chan<- prometheus.Metric, body []byte) error {
	// TODO 官方处理sub_status的方法 https://github.com/nginxinc/nginx-prometheus-exporter/blob/v0.11.0/client/nginx.go
	parts := strings.Fields(string(body))
	activeConnections, _ := strconv.Atoi(parts[2])
	accepts, _ := strconv.Atoi(parts[7])
	handled, _ := strconv.Atoi(parts[8])
	requests, _ := strconv.Atoi(parts[9])
	reading, _ := strconv.Atoi(parts[11])
	writing, _ := strconv.Atoi(parts[13])
	waiting, _ := strconv.Atoi(parts[15])
	ch <- prometheus.MustNewConstMetric(e.ActiveConnections, prometheus.CounterValue, float64(activeConnections))
	ch <- prometheus.MustNewConstMetric(e.Accepts, prometheus.CounterValue, float64(accepts))
	ch <- prometheus.MustNewConstMetric(e.Handled, prometheus.CounterValue, float64(handled))
	ch <- prometheus.MustNewConstMetric(e.Requests, prometheus.CounterValue, float64(requests))
	ch <- prometheus.MustNewConstMetric(e.Reading, prometheus.CounterValue, float64(reading))
	ch <- prometheus.MustNewConstMetric(e.Writing, prometheus.CounterValue, float64(writing))
	ch <- prometheus.MustNewConstMetric(e.Waiting, prometheus.CounterValue, float64(waiting))
	return nil
}
