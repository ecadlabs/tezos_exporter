# Tezos Exporter for Prometheus (WIP)

This is a metrics exporter that queries tezos node's RPC API, calculates
metrics, and exports them via HTTP for prometheus consumption.

## Getting Started

### To run it from current directory

```
go build
./tezos_exporter -tezos-rpc http://your_tezos_node:8732/ 
```


