package collector

import (
	"context"
	"net"
	"time"

	"github.com/ecadlabs/go-tezos"
	"github.com/prometheus/client_golang/prometheus"
)

const bootstrappedTimeout = 2 * time.Second

var (
	sentBytesDesc = prometheus.NewDesc(
		"tezos_node_sent_bytes_total",
		"Total number of bytes sent from this node.",
		nil,
		nil)

	recvBytesDesc = prometheus.NewDesc(
		"tezos_node_recv_bytes_total",
		"Total number of bytes received by this node.",
		nil,
		nil)

	connsDesc = prometheus.NewDesc(
		"tezos_node_connections",
		"Current number of connections to/from this node.",
		[]string{"direction", "private"},
		nil)

	pointsDesc = prometheus.NewDesc(
		"tezos_node_points",
		"Stats about known network points.",
		[]string{"trusted", "event_kind"},
		nil)

	bootstrappedDesc = prometheus.NewDesc(
		"tezos_node_bootstrapped",
		"Returns 1 if the node has synchronized its chain with a few peers.",
		nil,
		nil)
)

// NetworkCollector collects metrics about a Tezos node's network properties.
type NetworkCollector struct {
	reportRPCResult func(string, error, chan<- prometheus.Metric)
	service         *tezos.Service
	timeout         time.Duration
}

// NewNetworkCollector returns a new NetworkCollector.
func NewNetworkCollector(
	reportRPCResult func(string, error, chan<- prometheus.Metric),
	service *tezos.Service,
	timeout time.Duration) *NetworkCollector {
	return &NetworkCollector{
		reportRPCResult: reportRPCResult,
		service:         service,
		timeout:         timeout,
	}
}

// Describe implements prometheus.Collector.
func (c *NetworkCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- sentBytesDesc
	ch <- recvBytesDesc
}

func (c *NetworkCollector) getConnStats(ctx context.Context) (map[string]map[string]int, error) {
	conns, err := c.service.GetNetworkConnections(ctx)
	if err != nil {
		return nil, err
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

	return connStats, nil
}

func (c *NetworkCollector) getPointStats(ctx context.Context) (map[string]map[string]int, error) {
	points, err := c.service.GetNetworkPoints(ctx, "")
	if err != nil {
		return nil, err
	}

	pointStats := map[string]map[string]int{
		"false": map[string]int{},
		"true":  map[string]int{},
	}

	for _, point := range points {
		trusted := "false"
		if point.Trusted {
			trusted = "true"
		}

		pointStats[trusted][point.State.EventKind]++
	}

	return pointStats, nil
}

func (c *NetworkCollector) getBootstrapped(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, bootstrappedTimeout)
	defer cancel()

	ch := make(chan *tezos.BootstrappedBlock, 10)
	var err error

	go func() {
		err = c.service.GetBootstrapped(ctx, ch)
		close(ch)
	}()

	var cnt int
	for range ch {
		if cnt > 0 {
			// More than one record returned
			return false, nil
		}
		cnt++
	}

	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Collect implements prometheus.Collector and is called by the Prometheus registry when collecting metrics.
func (c *NetworkCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	stats, err := c.service.GetNetworkStats(ctx)
	c.reportRPCResult("/network/stat", err, ch)
	if err == nil {
		ch <- prometheus.MustNewConstMetric(sentBytesDesc, prometheus.CounterValue, float64(stats.TotalBytesSent))
		ch <- prometheus.MustNewConstMetric(recvBytesDesc, prometheus.CounterValue, float64(stats.TotalBytesRecv))
	}

	connStats, err := c.getConnStats(ctx)
	c.reportRPCResult("/network/connections", err, ch)
	if err == nil {
		for direction, stats := range connStats {
			for private, count := range stats {
				ch <- prometheus.MustNewConstMetric(connsDesc, prometheus.GaugeValue, float64(count), direction, private)
			}
		}
	}

	pointStats, err := c.getPointStats(ctx)
	c.reportRPCResult("/network/points", err, ch)
	if err == nil {
		for trusted, stats := range pointStats {
			for eventKind, count := range stats {
				ch <- prometheus.MustNewConstMetric(pointsDesc, prometheus.GaugeValue, float64(count), trusted, eventKind)
			}
		}
	}

	bootstrapped, err := c.getBootstrapped(ctx)
	c.reportRPCResult("/monitor/bootstrapped", err, ch)
	if err == nil {
		var v float64
		if bootstrapped {
			v = 1.0
		}
		ch <- prometheus.MustNewConstMetric(bootstrappedDesc, prometheus.GaugeValue, v)
	}
}
