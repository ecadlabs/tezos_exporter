package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tezos "github.com/ecadlabs/go-tezos"
	"github.com/ecadlabs/tezos_exporter/collector"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/oklog/pkg/group"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	metricsAddr := flag.String("metrics-listen-addr", ":9489", "TCP address on which to serve Prometheus metrics")
	tezosAddr := flag.String("tezos-rpc", "http://localhost:8732", "TCP address of tezos node to monitor")
	_ = tezosAddr

	flag.Parse()

	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)

	// metrics := newTezosMetrics()
	client, err := tezos.NewClient(nil, *tezosAddr)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprintf("Can't initlize tezos client: %v", err))
	}

	errors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tezos_rpc_errors_total",
		Help: "The total number of errors per rpc metric",
	}, []string{"collector"})

	timeout := time.Duration(600) * time.Millisecond

	reg := prometheus.NewRegistry()
	reg.Register(prometheus.NewProcessCollector(os.Getpid(), ""))
	reg.Register(prometheus.NewGoCollector())
	reg.Register(collector.NewNetworkCollector(errors, client, timeout))
	// reg.Register(metrics)

	ctx, cancel := context.WithCancel(context.Background())

	var g group.Group
	{
		// SIGTERM termination signal handler
		term := make(chan os.Signal)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)
		g.Add(
			func() error {
				select {
				case <-term:
					level.Warn(logger).Log("msg", "Recieved SIGTERM, exiting gracefully...")
				case <-ctx.Done():
					break
				}
				return nil
			},
			func(err error) {
				cancel()
			},
		)
	}
	{
		// Prometheus metrics HTTP server handler.
		g.Add(
			func() error {
				http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
				s := http.Server{
					Addr: *metricsAddr,
				}
				errCh := make(chan error)
				go func() {
					errCh <- s.ListenAndServe()
				}()

				select {
				case err := <-errCh:
					return fmt.Errorf("error serving Prometheus metrics: %v", err)
				case <-ctx.Done():
					level.Info(logger).Log("msg", "Shutting down metrics HTTP server...")
					s.Shutdown(ctx)
					return nil
				}
			},
			func(err error) {
				cancel()
			},
		)
	}

	if err := g.Run(); err != nil {
		level.Error(logger).Log("msg", "Error running exporter", "err", err)
	}
	level.Info(logger).Log("msg", "Shutdown complete!")

}
