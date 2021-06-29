package collector

import (
	"context"
	"net/http"
	"sync"

	tezos "github.com/ecadlabs/tezos_exporter/go-tezos"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MempoolOperationsCollector collects mempool operations count
type MempoolOperationsCollector struct {
	counter        *prometheus.CounterVec
	rpcTotalHist   prometheus.ObserverVec
	rpcConnectHist prometheus.Histogram

	service *tezos.Service
	chainID string
	wg      sync.WaitGroup
}

func (m *MempoolOperationsCollector) listener(pool string) {
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
		err := m.service.MonitorMempoolOperations(context.Background(), m.chainID, pool, ch)
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
				Namespace: "tezos_node",
				Subsystem: "mempool",
				Name:      "operations_total",
				Help:      "The total number of mempool operations.",
			},
			[]string{"pool", "proto", "kind"},
		),
		rpcTotalHist: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "tezos_rpc",
				Subsystem: "mempool",
				Name:      "monitor_connection_total_duration_seconds",
				Help:      "The total life time of the mempool monitor RPC connection.",
				Buckets:   prometheus.ExponentialBuckets(0.25, 2, 12),
			},
			[]string{},
		),
		rpcConnectHist: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: "tezos_rpc",
				Subsystem: "mempool",
				Name:      "monitor_connection_connect_duration_seconds",
				Help:      "Mempool monitor (re)connection duration (time until HTTP header arrives).",
				Buckets:   prometheus.ExponentialBuckets(0.25, 2, 12),
			},
		),
		chainID: chainID,
	}

	it := promhttp.InstrumentTrace{
		GotConn: func(t float64) {
			c.rpcConnectHist.Observe(t)
		},
	}

	client := *service.Client
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}

	client.Transport = promhttp.InstrumentRoundTripperDuration(c.rpcTotalHist, client.Transport)
	client.Transport = promhttp.InstrumentRoundTripperTrace(&it, client.Transport)

	srv := *service
	srv.Client = &client
	c.service = &srv

	for _, p := range pools {
		c.wg.Add(1)
		go c.listener(p)
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
