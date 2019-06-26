# build stage
FROM golang:alpine AS build-env
WORKDIR  /tezos_exporter
ADD . .
RUN apk --no-cache add git
RUN go get -d ./...
RUN go build

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /tezos_exporter/tezos_exporter /usr/bin/tezos_exporter
ENTRYPOINT ["/usr/bin/tezos_exporter"]
CMD ["-tezos-node-url" "http://localhost:8732/"]
