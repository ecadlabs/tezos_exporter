package collector

import (
	"context"
	"sync"

	"github.com/ecadlabs/go-tezos"
	"github.com/prometheus/client_golang/prometheus"
)

type MempoolOperationsCollector struct {
	prometheus.Collector // refers to counter

	counter *prometheus.CounterVec
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

func NewMempoolOperationsCollectorCollector(service *tezos.Service, chainID string, pools []string) *MempoolOperationsCollector {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tezos_node_mempool_operations_total",
			Help: "The total number of mempool operations.",
		},
		[]string{"pool", "proto", "kind"},
	)

	c := MempoolOperationsCollector{
		Collector: counter,
		counter:   counter,
		service:   service,
		chainID:   chainID,
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	for _, p := range pools {
		c.wg.Add(1)
		go c.listener(ctx, p)
	}

	return &c
}

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
