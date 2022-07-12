<div align="center">
<img src="./docs/logo.svg" alt="Provenance"/>
</div>
<br/><br/>

# Provenance Blockchain

[Provenance] is a distributed, proof of stake blockchain designed for the financial services industry.

For more information about [Provenance Inc](https://provenance.io) visit https://provenance.io


The Provenance app is the core blockchain application for running a node on the Provenance Network.  The node
software is based on the open source [Tendermint](https://tendermint.com) consensus engine combined with the 
[Cosmos SDK](https://cosmos.network) and custom modules to support apis for financial services. [Figure](https://figure.com)
is the first and primary user of the Provenance Blockchain.

## Status

[![Latest Release][release-badge]][release-latest]
[![Apache 2.0 License][license-badge]][license-url]
[![Go Report][goreport-badge]][goreport-url]
[![Code Coverage][cover-badge]][cover-report]
[![LOC][loc-badge]][loc-report]
![Lint Status][lint-badge]


[license-badge]: https://img.shields.io/github/license/provenance-io/provenance.svg
[license-url]: https://github.com/provenance-io/provenance/blob/main/LICENSE
[release-badge]: https://img.shields.io/github/tag/provenance-io/provenance.svg
[release-latest]: https://github.com/provenance-io/provenance/releases/latest
[goreport-badge]: https://goreportcard.com/badge/github.com/provenance-io/provenance
[goreport-url]: https://goreportcard.com/report/github.com/provenance-io/provenance
[cover-badge]: https://codecov.io/gh/provenance-io/provenance/branch/main/graph/badge.svg
[cover-report]: https://codecov.io/gh/provenance-io/provenance
[loc-badge]: https://tokei.rs/b1/github/provenance-io/provenance
[loc-report]: https://github.com/provenance-io/provenance
[lint-badge]: https://github.com/provenance-io/provenance/workflows/Lint/badge.svg
[provenance]: https://provenance.io/#overview

The Provenance networks are based on work from the private [Figure Technologies](https://figure.com) blockchain launched in 2018.

## Quick Start

The Provenance Blockchain is based on Cosmos, the [sdk introduction](https://github.com/cosmos/cosmos-sdk/blob/master/docs/intro/overview.md)
is a useful starting point.

Developers can use a local checkout and the make targets `make run` and `make localnet-start` to run a local development network.

Note: Requires [Go 1.17+](https://golang.org/dl/)

See Also: [Building](docs/Building.md)

## Active Networks

There are two active public Provenance networks, [testnet](https://github.com/provenance-io/testnet) and [mainnet](https://github.com/provenance-io/mainnet).
