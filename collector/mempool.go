package collector

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/ecadlabs/go-tezos"
	"github.com/prometheus/client_golang/prometheus"
)

// MempoolOperationsCollector collects mempool operations count
type MempoolOperationsCollector struct {
	counter        *prometheus.CounterVec
	rpcTotalHist   prometheus.Histogram
	rpcConnectHist prometheus.Histogram

	service *tezos.Service
	chainID string
	wg      sync.WaitGroup
	cancel  context.CancelFunc
}

func (m *MempoolOperationsCollector) listener(ctx context.Context, pool string) {
	ch := make(chan []*tezos.Operation, 100)
	defer close(ch)

	go func() {
		for ops := range ch {
			for _, op := range ops {
				for _, elem := range op.Contents {
					m.counter.WithLabelValues(pool, op.Protocol, elem.OperationElemKind()).Inc()
				}
			}
		}
	}()

	for {
		err := m.service.MonitorMempoolOperations(ctx, m.chainID, pool, ch)
		if err == context.Canceled {
			return
		}
	}
}

// NewMempoolOperationsCollectorCollector returns new mempool collector for given pools like "applied", "refused" etc.
func NewMempoolOperationsCollectorCollector(service *tezos.Service, chainID string, pools []string) *MempoolOperationsCollector {
	c := &MempoolOperationsCollector{
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tezos_node_mempool_operations_total",
				Help: "The total number of mempool operations.",
			},
			[]string{"pool", "proto", "kind"},
		),
		rpcTotalHist: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "tezos_rpc_mempool_monitor_connection_total_duration_seconds",
			Help:    "The total life time of the mempool monitir RPC connection.",
			Buckets: prometheus.ExponentialBuckets(0.25, 2, 12),
		}),
		rpcConnectHist: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "tezos_rpc_mempool_monitor_connection_connect_duration_seconds",
			Help:    "Mempool monitor (re)connection duration (time until HTTP header arrives).",
			Buckets: prometheus.ExponentialBuckets(0.25, 2, 12),
		}),
		chainID: chainID,
	}

	client := *service.Client
	client.RPCStatusCallback = func(req *http.Request, status int, duration time.Duration, err error) {
		c.rpcTotalHist.Observe(float64(duration / time.Second))
	}
	client.RPCHeaderCallback = func(req *http.Request, resp *http.Response, duration time.Duration) {
		c.rpcConnectHist.Observe(float64(duration / time.Second))
	}

	srv := *service
	srv.Client = &client
	c.service = &srv

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	for _, p := range pools {
		c.wg.Add(1)
		go c.listener(ctx, p)
	}

	return c
}

// Describe implements prometheus.Collector
func (m *MempoolOperationsCollector) Describe(ch chan<- *prometheus.Desc) {
	m.counter.Describe(ch)
	m.rpcTotalHist.Describe(ch)
	m.rpcConnectHist.Describe(ch)
}

// Collect implements prometheus.Collector
func (m *MempoolOperationsCollector) Collect(ch chan<- prometheus.Metric) {
	m.counter.Collect(ch)
	m.rpcTotalHist.Collect(ch)
	m.rpcConnectHist.Collect(ch)
}

// Shutdown stops all listeners
func (m *MempoolOperationsCollector) Shutdown(ctx context.Context) error {
	m.cancel()

	sem := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(sem)
	}()

	select {
	case <-sem:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
