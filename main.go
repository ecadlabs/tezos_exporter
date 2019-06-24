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

func main() {
	metricsAddr := flag.String("metrics-listen-addr", ":9489", "TCP address on which to serve Prometheus metrics.")
	tezosAddr := flag.String("tezos-node-url", "http://localhost:8732", "URL of Tezos node to monitor.")
	chainID := flag.String("chain-id", "main", "ID of chain about which to report chain-related stats.")
	rpcTimeout := flag.Duration("rpc-timeout", 10*time.Second, "Timeout for connecting to tezos RPCs")

	flag.Parse()

	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)

	client, err := tezos.NewRPCClient(nil, *tezosAddr)
	if err != nil {
		level.Error(logger).Log("msg", "error initializing Tezos RPC client", "err", err)
		os.Exit(1)
	}

	service := &tezos.Service{Client: client}

	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(collector.NewNetworkCollector(service, *rpcTimeout, *chainID))
	reg.MustRegister(collector.NewMempoolOperationsCollectorCollector(service, *chainID, []string{
		"applied",
		"branch_refused",
		"refused",
		"branch_delayed",
	}))

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	level.Info(logger).Log("msg", "tezos_exporter starting...", "address", metricsAddr)

	if err := http.ListenAndServe(*metricsAddr, nil); err != nil {
		level.Error(logger).Log("msg", "error starting webserver", "err", err)
		os.Exit(1)
	}
}
