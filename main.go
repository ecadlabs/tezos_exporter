package main

import (
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ecadlabs/tezos_exporter/collector"
	tezos "github.com/ecadlabs/tezos_exporter/go-tezos"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	metricsAddr := flag.String("metrics-listen-addr", ":9489", "TCP address on which to serve Prometheus metrics")
	tezosAddr := flag.String("tezos-node-url", "http://localhost:8732", "URL of Tezos node to monitor")
	chainID := flag.String("chain-id", "main", "ID of chain about which to report chain-related stats")
	rpcTimeout := flag.Duration("rpc-timeout", 10*time.Second, "Timeout for connecting to tezos RPCs")
	noHealthEp := flag.Bool("disable-health-endpoint", false, "Disable /health endpoint")
	isBootstrappedPollInterval := flag.Duration("bootstraped-poll-interval", 10*time.Second, "is_bootstrapped endpoint polling interval")
	isBootstrappedThreshold := flag.Int("bootstraped-threshold", 3, "Report is_bootstrapped change after N samples of the same value")
	mempoolRetryInterval := flag.Duration("mempool-retry-delay", 30*time.Second, "Retry mempool monitoring after a delay in case of an error")
	pools := flag.String("mempool-pools", "applied,branch_refused,refused,branch_delayed", "Mempool pools")

	flag.Parse()

	client, err := tezos.NewRPCClient(*tezosAddr)
	if err != nil {
		log.WithError(err).Error("error initializing Tezos RPC client")
		os.Exit(1)
	}

	service := &tezos.Service{Client: client}

	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(collector.NewBuildInfoCollector(""))
	reg.MustRegister(collector.NewNetworkCollector(service, *rpcTimeout, *chainID))
	reg.MustRegister(collector.NewMempoolOperationsCollectorCollector(service, *chainID, strings.Split(*pools, ","), *mempoolRetryInterval))

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	if !*noHealthEp {
		http.Handle("/health", NewHealthHandler(service, *chainID, *isBootstrappedPollInterval, *isBootstrappedThreshold))
	}

	log.WithField("address", *metricsAddr).Info("tezos_exporter starting...")

	if err := http.ListenAndServe(*metricsAddr, nil); err != nil {
		log.WithError(err).Error("error starting webserver")
		os.Exit(1)
	}
}
