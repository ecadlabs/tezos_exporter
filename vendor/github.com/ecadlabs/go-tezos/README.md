# go-tezos

go-tezos is a Go client library that is used to interact with a Tezos' nodes
RPC methods.

[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/2349/badge)](https://bestpractices.coreinfrastructure.org/projects/2349)
[![CircleCI](https://circleci.com/gh/ecadlabs/go-tezos/tree/master.svg?style=svg)](https://circleci.com/gh/ecadlabs/go-tezos/tree/master)


## Work in progress, contributions welcome.

This client RPC library is in development and should be considered alpha.
Contributors are welcome. We will start to tag releases of this library in
November 2018.

The library will be useful to anyone wanting to build tools, products or
services on top of the Tezos RPC API in go.

The library will be:

* Well tested
* Nightly Integration tests against official Tezos docker images
* Written in Idiomatic Go
* Aim to have complete coverage of the Tezos API and stay up to date with new
  RPCs or changes to existing RPCs

# Documentation

Library documentation lives in the code as godoc comments. Readers can view
up-to-date documentation here: https://godoc.org/github.com/ecadlabs/go-tezos

# Contributions

## Reporting issues/feature requests

Please use the [GitHub issue
tracker](https://github.com/ecadlabs/go-tezos/issues) to report bugs or request
features.

## Contribution

To contribute, please check the issue tracker to see if an existing issue
exists for your planned contribution. If there's no Issue, please create one
first, and then submit a pull request with your contribution. 

For a contribution to be merged, it must be well documented, come with unit
tests, and integration tests where appropriate. Submitting a "Work in progress"
pull request is welcome!

## Reporting Security Issues

If a security vulnerabiltiy in this project is discovered, please report the
issue to go-tezos@ecadlabs.com or to `jevonearth` on keybase.io

Reports may be encrypted using keys published on keybase.io

# Tezos RPC API documentation

The best known RPC API docs are available here: http://tezos.gitlab.io/mainnet/

# Users of `go-tezos`

* A prometheus metrics exporter for a Tezos node https://github.com/ecadlabs/tezos_exporter

## Development

### Running a tezos RPC node using docker-compose

To run a local tezos RPC node using docker, run the following command:

`docker-compose up`

The node will generate an identity and then, it will the chain from other nodes
on the network. The process of synchronizing or downloading the chain can take
some time, but most of the RPC will work while this process completes.

The `alphanet` image tag means you are not interacting with the live `mainnet`.
You can connect to `mainnet` with the `tezos/tezos:mainnet` image, but it takes
longer to sync.

The `docker-compose.yml` file uses volumes, so when you restart the node, it
won't have to regenerate an identity, or sync the entire chain.

### Running a tezos RPC node using docker

If you want to run a tezos node quickly, without using `docker-compose` try:

`docker run -it --rm --name tezos_node -p 8732:8732 tezos/tezos:alphanet tezos-node`

### Interacting with tezos RPC

With the tezos-node docker image, you can test that the RPC interface is
working:

`curl localhost:8732/network/stat`

The tezos-client cli is available in the docker image, and can be run as
follows:

`docker exec -it tezos_node tezos-client -A 0.0.0.0 man`

`docker exec -it tezos_node tezos-client -A 0.0.0.0 rpc list`

Create a shell alias that you can run from your docker host for convenience;

`alias tezos-client='sudo docker exec -it -e TEZOS_CLIENT_UNSAFE_DISABLE_DISCLAIMER=Y tezos_node tezos-client -A 0.0.0.0'`
