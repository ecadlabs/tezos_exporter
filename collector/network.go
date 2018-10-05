package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/ecadlabs/go-tezos"
	"github.com/prometheus/client_golang/prometheus"
)

// NetworkCollector collects metrics about a Tezos nodes network properties
type NetworkCollector struct {
	//	logger  log.Logger
	errors  *prometheus.CounterVec
	client  *tezos.Client
	timeout time.Duration

	nodeSentBytesTotal *prometheus.CounterVec
	nodeRecvBytesTotal *prometheus.CounterVec
}

// NewNetworkCollector returns a new NewtworkCollector
func NewNetworkCollector(errors *prometheus.CounterVec, client *tezos.Client, timeout time.Duration) *NetworkCollector {
	errors.WithLabelValues("network").Add(0)

	return &NetworkCollector{
		//		logger:  logger,
		errors:  errors,
		client:  client,
		timeout: timeout,

		nodeSentBytesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tezos_node_sent_bytes_total",
				Help: "Total number of bytes sent from this node",
			},
			[]string{"identity"},
		),
		nodeRecvBytesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tezos_node_recv_bytes_total",
				Help: "Total number of bytes recieved by this node",
			},
			[]string{"identity"},
		),
	}
}
func (c *NetworkCollector) Describe(ch chan<- *prometheus.Desc) {

	c.nodeSentBytesTotal.Describe(ch)
	c.nodeRecvBytesTotal.Describe(ch)
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *NetworkCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	stats, err := c.client.Network.GetStats(ctx)
	if err != nil {
		c.errors.WithLabelValues("network_stat").Add(1)
		fmt.Print(err)
		// level.Warn(c.logger).Log("msg", "can't query /network/stat", "err", err)
	}
	c.nodeSentBytesTotal.WithLabelValues("identity").Add(float64(stats.TotalBytesSent))
	c.nodeSentBytesTotal.Collect(ch)
	c.nodeRecvBytesTotal.WithLabelValues("identity").Add(float64(stats.TotalBytesRecv))
	c.nodeRecvBytesTotal.Collect(ch)
	//ch <- prometheus.New
	//get data from rpc
	//{ "total_sent": "76511326", "total_recv": "393777480", "current_inflow": 844,
	//  "current_outflow": 192 }
}
