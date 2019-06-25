# Tezos Exporter for Prometheus (WIP)

[![CircleCI](https://circleci.com/gh/ecadlabs/tezos_exporter.svg?style=svg)](https://circleci.com/gh/ecadlabs/tezos_exporter)

`tezos_exporter` produces metrics by querying the RPC methods of a Tezos node.

## Getting Started using docker

To get started using the provided docker images, run:

```sh
docker run -it --rm --name tezos_exporter ecadlabs/tezos_exporter \
    -tezos-node-url http://YOUR_TEZOS_NODE:8732/
```

You will need to configure a prometheus server to scrape the metrics from your
newly running exporter. Add the following scrape job to your `promethus.yml`
configuration file. 

```yaml
  - job_name: tezos
    scrape_interval: 30s
    file_sd_configs:
    static_configs:
      - targets: ['EXPORTER_ADDRESS:9489']
```

Restart promethues, and you should see the new job named `tezos` by looking at
`Targets` via the prometheus UI.

Metric names are as follows;

* tezos_node_bootstrapped
* tezos_node_connections
* tezos_node_mempool_operations
* tezos_node_peers
* tezos_node_points
* tezos_node_recv_bytes_total
* tezos_node_sent_bytes_total
* tezos_rpc_failed

To request a new metric be added, please file a new feature request Issue in
the github tracker, or submit a Pull Request. Contributors welcome!

## Reporting issues/feature requests

Please use the [GitHub issue
tracker](https://github.com/ecadlabs/go-tezos/issues) to report bugs or request
features.

## Contributions

To contribute, please check the issue tracker to see if an existing issue
exists for your planned contribution. If there's no Issue, please create one
first, and then submit a pull request with your contribution.

For a contribution to be merged, it must be well documented, come with unit
tests, and integration tests where appropriate. Submitting a "Work in progress"
pull request is welcome!

## Reporting Security Issues

If a security vulnerability in this project is discovered, please report the
issue to security@ecadlabs.com or to `jevonearth` on keybase.io

Reports may be encrypted using keys published on keybase.io
