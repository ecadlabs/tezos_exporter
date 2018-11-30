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

	errors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tezos_rpc_errors_total",
		Help: "The total number of errors per RPC metric collector.",
	}, []string{"collector"})

	reg := prometheus.NewRegistry()
	reg.Register(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	reg.Register(prometheus.NewGoCollector())
	reg.Register(collector.NewNetworkCollector(logger, errors, service, defaultTimeout))
	reg.Register(errors)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	if err := http.ListenAndServe(*metricsAddr, nil); err != nil {
		level.Error(logger).Log("msg", "error starting webserver", "err", err)
		os.Exit(1)
	}
}
