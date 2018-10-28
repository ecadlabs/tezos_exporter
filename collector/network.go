package collector

import (
	"context"
	"time"

	"github.com/ecadlabs/go-tezos"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	sentBytesDesc = prometheus.NewDesc(
		"tezos_node_sent_bytes_total",
		"Total number of bytes sent from this node.",
		nil,
		nil,
	)
	recvBytesDesc = prometheus.NewDesc(
		"tezos_node_recv_bytes_total",
		"Total number of bytes received by this node.",
		nil,
		nil,
	)

	connsDesc = prometheus.NewDesc(
		"tezos_node_connections",
		"Current number of connections to/from this node.",
		[]string{"direction", "private"},
		nil,
	)
)

// NetworkCollector collects metrics about a Tezos node's network properties.
type NetworkCollector struct {
	logger  log.Logger
	errors  *prometheus.CounterVec
	service tezos.NetworkService
	timeout time.Duration
}

// NewNetworkCollector returns a new NetworkCollector.
func NewNetworkCollector(logger log.Logger, errors *prometheus.CounterVec, service tezos.NetworkService, timeout time.Duration) *NetworkCollector {
	errors.WithLabelValues("network").Add(0)

	return &NetworkCollector{
		logger:  logger,
		errors:  errors,
		service: service,
		timeout: timeout,
	}
}

// Describe implements prometheus.Collector.
func (c *NetworkCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- sentBytesDesc
	ch <- recvBytesDesc
}

// Collect implements prometheus.Collector and is called by the Prometheus registry when collecting metrics.
func (c *NetworkCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	stats, err := c.service.GetStats(ctx)
	if err != nil {
		c.errors.WithLabelValues("network_stat").Add(1)
		level.Warn(c.logger).Log("msg", "error querying /network/stat", "err", err)
	}
	ch <- prometheus.MustNewConstMetric(sentBytesDesc, prometheus.CounterValue, float64(stats.TotalBytesSent))
	ch <- prometheus.MustNewConstMetric(recvBytesDesc, prometheus.CounterValue, float64(stats.TotalBytesRecv))

	conns, err := c.service.GetConnections(ctx)
	if err != nil {
		c.errors.WithLabelValues("network_stat").Add(1)
		level.Warn(c.logger).Log("msg", "error querying /network/stat", "err", err)
	}

	connStats := map[string]map[string]int{
		"incoming": map[string]int{
			"false": 0,
			"true":  0,
		},
		"outgoing": map[string]int{
			"false": 0,
			"true":  0,
		},
	}
	for _, conn := range conns {
		direction := "outgoing"
		if conn.Incoming {
			direction = "incoming"
		}
		private := "false"
		if conn.Private {
			private = "true"
		}

		connStats[direction][private]++
	}

	for direction, stats := range connStats {
		for private, count := range stats {
			ch <- prometheus.MustNewConstMetric(connsDesc, prometheus.GaugeValue, float64(count), direction, private)
		}
	}
}
