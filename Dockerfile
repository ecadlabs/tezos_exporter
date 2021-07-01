FROM alpine

ENTRYPOINT ["/usr/bin/tezos_exporter"]
CMD ["-tezos-node-url" "http://localhost:8732/"]


COPY tezos_exporter /usr/bin/tezos_exporter
