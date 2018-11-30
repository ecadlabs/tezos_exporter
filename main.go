package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	tezos "github.com/ecadlabs/go-tezos"
	"github.com/ecadlabs/tezos_exporter/collector"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const defaultTimeout = 5 * time.Second

var rpcFailedDesc = prometheus.NewDesc(
	"tezos_rpc_failed",
	"A gauge that is set to 1 when a metrics collection RPC failed during the current scrape, 0 otherwise.",
	[]string{"rpc"},
	nil)

func main() {
	metricsAddr := flag.String("metrics-listen-addr", ":9489", "TCP address on which to serve Prometheus metrics.")
	tezosAddr := flag.String("tezos-node-url", "http://localhost:8732", "URL of Tezos node to monitor.")

	flag.Parse()

	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)

	client, err := tezos.NewRPCClient(nil, *tezosAddr)
	if err != nil {
		level.Error(logger).Log("msg", "error initializing Tezos RPC client", "err", err)
		os.Exit(1)
	}
	service := &tezos.Service{Client: client}

	reportRPCResult := func(rpc string, err error, ch chan<- prometheus.Metric) {
		var val float64
		if err != nil {
			val = 1
		}
		ch <- prometheus.MustNewConstMetric(rpcFailedDesc, prometheus.GaugeValue, val, rpc)
		if err != nil {
			level.Warn(logger).Log("msg", "error querying RPC", "rpc", rpc, "err", err)
		}
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(collector.NewNetworkCollector(reportRPCResult, service, defaultTimeout))

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	if err := http.ListenAndServe(*metricsAddr, nil); err != nil {
		level.Error(logger).Log("msg", "error starting webserver", "err", err)
		os.Exit(1)
	}
}
