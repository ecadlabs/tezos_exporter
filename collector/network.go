package collector

import (
	"context"
	"net"
	"net/http"
	"time"

	tezos "github.com/ecadlabs/tezos_exporter/go-tezos"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const bootstrappedTimeout = 5 * time.Second
const bootstrappedPollInterval = 30 * time.Second

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

	peersDesc = prometheus.NewDesc(
		"tezos_node_peers",
		"Stats about all peers this node ever met.",
		[]string{"trusted", "state"},
		nil)

	pointsDesc = prometheus.NewDesc(
		"tezos_node_points",
		"Stats about known network points.",
		[]string{"trusted", "event_kind"},
		nil)

	rpcFailedDesc = prometheus.NewDesc(
		"tezos_rpc_failed",
		"A gauge that is set to 1 when a metrics collection RPC failed during the current scrape, 0 otherwise.",
		[]string{"rpc"},
		nil)
)

// NetworkCollector collects metrics about a Tezos node's network properties.
type NetworkCollector struct {
	service      *tezos.Service
	timeout      time.Duration
	chainID      string
	bootstrapped prometheus.Gauge
	sem          chan struct{}
	cancel       context.CancelFunc
}

// NewNetworkCollector returns a new NetworkCollector.
func NewNetworkCollector(service *tezos.Service, timeout time.Duration, chainID string) *NetworkCollector {
	c := &NetworkCollector{
		service: service,
		timeout: timeout,
		chainID: chainID,
		bootstrapped: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "tezos_node",
			Name:      "bootstrapped",
			Help:      "Returns 1 if the node has synchronized its chain with a few peers.",
		}),
		sem: make(chan struct{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	go c.bootstrappedPollLoop(ctx)

	return c
}

func (c *NetworkCollector) getBootstrapped(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, bootstrappedTimeout)
	defer cancel()

	ch := make(chan *tezos.BootstrappedBlock, 10)
	var err error

	go func() {
		err = c.service.MonitorBootstrapped(ctx, ch)
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

func (c *NetworkCollector) bootstrappedPollLoop(ctx context.Context) {
	defer close(c.sem)
	t := time.NewTicker(bootstrappedPollInterval)
	defer t.Stop()

	for {
		ok, err := c.getBootstrapped(ctx)
		if err == context.Canceled {
			return
		}
		var v float64
		if ok {
			v = 1
		}
		c.bootstrapped.Set(v)

		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// Describe implements prometheus.Collector.
func (c *NetworkCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func getConnStats(ctx context.Context, service *tezos.Service) (map[string]map[string]int, error) {
	conns, err := service.GetNetworkConnections(ctx)
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

func getPointStats(ctx context.Context, service *tezos.Service) (map[string]map[string]int, error) {
	points, err := service.GetNetworkPoints(ctx, "")
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
func getPeerStats(ctx context.Context, service *tezos.Service) (map[string]map[string]int, error) {
	peers, err := service.GetNetworkPeers(ctx, "")
	if err != nil {
		return nil, err
	}

	peerStats := map[string]map[string]int{
		"false": map[string]int{},
		"true":  map[string]int{},
	}

	for _, peer := range peers {
		trusted := "false"
		if peer.Trusted {
			trusted = "true"
		}

		peerStats[trusted][peer.State]++
	}

	return peerStats, nil
}

// Collect implements prometheus.Collector and is called by the Prometheus registry when collecting metrics.
func (c *NetworkCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client := *c.service.Client
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	var path string
	client.Transport = promhttp.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		path = r.URL.Path
		return transport.RoundTrip(r)
	})

	srv := *c.service
	srv.Client = &client

	stats, err := srv.GetNetworkStats(ctx)
	if err == nil {
		ch <- prometheus.MustNewConstMetric(sentBytesDesc, prometheus.CounterValue, float64(stats.TotalBytesSent))
		ch <- prometheus.MustNewConstMetric(recvBytesDesc, prometheus.CounterValue, float64(stats.TotalBytesRecv))
	}
	var val float64
	if err != nil {
		val = 1
	}
	ch <- prometheus.MustNewConstMetric(rpcFailedDesc, prometheus.GaugeValue, val, path)

	connStats, err := getConnStats(ctx, &srv)
	if err == nil {
		for direction, stats := range connStats {
			for private, count := range stats {
				ch <- prometheus.MustNewConstMetric(connsDesc, prometheus.GaugeValue, float64(count), direction, private)
			}
		}
	}
	if err != nil {
		val = 1
	} else {
		val = 0
	}
	ch <- prometheus.MustNewConstMetric(rpcFailedDesc, prometheus.GaugeValue, val, path)

	peerStats, err := getPeerStats(ctx, &srv)
	if err == nil {
		for trusted, stats := range peerStats {
			for state, count := range stats {
				ch <- prometheus.MustNewConstMetric(peersDesc, prometheus.GaugeValue, float64(count), trusted, state)
			}
		}
	}
	if err != nil {
		val = 1
	} else {
		val = 0
	}
	ch <- prometheus.MustNewConstMetric(rpcFailedDesc, prometheus.GaugeValue, val, path)

	pointStats, err := getPointStats(ctx, &srv)
	if err == nil {
		for trusted, stats := range pointStats {
			for eventKind, count := range stats {
				ch <- prometheus.MustNewConstMetric(pointsDesc, prometheus.GaugeValue, float64(count), trusted, eventKind)
			}
		}
	}
	if err != nil {
		val = 1
	} else {
		val = 0
	}
	ch <- prometheus.MustNewConstMetric(rpcFailedDesc, prometheus.GaugeValue, val, path)

	c.bootstrapped.Collect(ch)
}

// Shutdown stops all listeners
func (c *NetworkCollector) Shutdown(ctx context.Context) error {
	c.cancel()

	sem := make(chan struct{})
	go func() {
		<-c.sem
		close(sem)
	}()

	select {
	case <-sem:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
