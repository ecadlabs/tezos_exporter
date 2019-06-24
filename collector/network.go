package collector

import (
	"context"
	"net"
	"net/http"
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

	bootstrappedDesc = prometheus.NewDesc(
		"tezos_node_bootstrapped",
		"Returns 1 if the node has synchronized its chain with a few peers.",
		nil,
		nil)

	/*
		mempoolDesc = prometheus.NewDesc(
			"tezos_node_mempool_operations",
			"The current number of mempool operations.",
			[]string{"pool", "proto", "kind"},
			nil)
	*/

	rpcFailedDesc = prometheus.NewDesc(
		"tezos_rpc_failed",
		"A gauge that is set to 1 when a metrics collection RPC failed during the current scrape, 0 otherwise.",
		[]string{"rpc"},
		nil)
)

// NetworkCollector collects metrics about a Tezos node's network properties.
type NetworkCollector struct {
	service *tezos.Service
	timeout time.Duration
	chainID string
}

// NewNetworkCollector returns a new NetworkCollector.
func NewNetworkCollector(service *tezos.Service, timeout time.Duration, chainID string) *NetworkCollector {
	return &NetworkCollector{
		service: service,
		timeout: timeout,
		chainID: chainID,
	}
}

// Describe implements prometheus.Collector.
func (c *NetworkCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- sentBytesDesc
	ch <- recvBytesDesc
	ch <- connsDesc
	ch <- peersDesc
	ch <- pointsDesc
	ch <- bootstrappedDesc
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

func getBootstrapped(ctx context.Context, service *tezos.Service) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, bootstrappedTimeout)
	defer cancel()

	ch := make(chan *tezos.BootstrappedBlock, 10)
	var err error

	go func() {
		err = service.GetBootstrapped(ctx, ch)
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

/*
func getMempoolStats(ctx context.Context, service *tezos.Service, chainID string) (map[string]map[string]map[string]int, error) {
	buildStats := func(ops []*tezos.Operation) map[string]map[string]int {
		stats := map[string]map[string]int{}
		for _, op := range ops {
			if _, ok := stats[op.Protocol]; !ok {
				stats[op.Protocol] = map[string]int{}
			}
			for _, contents := range op.Contents {
				stats[op.Protocol][contents.OperationElemKind()]++
			}
		}
		return stats
	}

	opsFromOpsWithErrorAlt := func(in []*tezos.OperationWithErrorAlt) []*tezos.Operation {
		out := make([]*tezos.Operation, len(in))
		for i, op := range in {
			out[i] = &op.Operation
		}
		return out
	}

	opsFromOpsAlt := func(in []*tezos.OperationAlt) []*tezos.Operation {
		out := make([]*tezos.Operation, len(in))
		for i, op := range in {
			out[i] = (*tezos.Operation)(op)
		}
		return out
	}

	ops, err := service.GetMempoolPendingOperations(ctx, chainID)
	if err != nil {
		return nil, err
	}

	return map[string]map[string]map[string]int{
		"applied":        buildStats(ops.Applied),
		"branch_delayed": buildStats(opsFromOpsWithErrorAlt(ops.BranchDelayed)),
		"branch_refused": buildStats(opsFromOpsWithErrorAlt(ops.BranchRefused)),
		"refused":        buildStats(opsFromOpsWithErrorAlt(ops.Refused)),
		"unprocessed":    buildStats(opsFromOpsAlt(ops.Unprocessed)),
	}, nil
}
*/

// Collect implements prometheus.Collector and is called by the Prometheus registry when collecting metrics.
func (c *NetworkCollector) Collect(ch chan<- prometheus.Metric) {
	client := *c.service.Client
	client.RPCStatusCallback = func(req *http.Request, status int, duration time.Duration, err error) {
		var val float64
		if err != nil {
			val = 1
		}
		ch <- prometheus.MustNewConstMetric(rpcFailedDesc, prometheus.GaugeValue, val, req.URL.Path)
	}

	srv := *c.service
	srv.Client = &client

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	stats, err := c.service.GetNetworkStats(ctx)
	if err == nil {
		ch <- prometheus.MustNewConstMetric(sentBytesDesc, prometheus.CounterValue, float64(stats.TotalBytesSent))
		ch <- prometheus.MustNewConstMetric(recvBytesDesc, prometheus.CounterValue, float64(stats.TotalBytesRecv))
	}

	connStats, err := getConnStats(ctx, &srv)
	if err == nil {
		for direction, stats := range connStats {
			for private, count := range stats {
				ch <- prometheus.MustNewConstMetric(connsDesc, prometheus.GaugeValue, float64(count), direction, private)
			}
		}
	}

	peerStats, err := getPeerStats(ctx, &srv)
	if err == nil {
		for trusted, stats := range peerStats {
			for state, count := range stats {
				ch <- prometheus.MustNewConstMetric(peersDesc, prometheus.GaugeValue, float64(count), trusted, state)
			}
		}
	}

	pointStats, err := getPointStats(ctx, &srv)
	if err == nil {
		for trusted, stats := range pointStats {
			for eventKind, count := range stats {
				ch <- prometheus.MustNewConstMetric(pointsDesc, prometheus.GaugeValue, float64(count), trusted, eventKind)
			}
		}
	}

	bootstrapped, err := getBootstrapped(ctx, &srv)
	if err == nil {
		var v float64
		if bootstrapped {
			v = 1.0
		}
		ch <- prometheus.MustNewConstMetric(bootstrappedDesc, prometheus.GaugeValue, v)
	}

	/*
		mempoolStats, err := getMempoolStats(ctx, &srv, c.chainID)
		if err == nil {
			for pool, stats := range mempoolStats {
				for proto, protoStats := range stats {
					for kind, count := range protoStats {
						ch <- prometheus.MustNewConstMetric(mempoolDesc, prometheus.GaugeValue, float64(count), pool, proto, kind)
					}
				}
			}
		}
	*/
}
