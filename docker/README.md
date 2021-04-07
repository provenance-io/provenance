# Quick reference

- **Maintained by**: [The Provenance Team](https://github.com/provenance-io/provenance)
- **Where to file issues**: [Provenance Issue Tracker](https://github.com/provenance-io/provenance/issues)

# Supported tags and respective `Dockerfile` links:

- [`v1.0.0`](https://github.com/provenance-io/provenance/blob/release/v1.0.0/docker/blockchain/Dockerfile)
- [`v0.3.0`](https://github.com/provenance-io/provenance/blob/release/v0.3.0/docker/blockchain/Dockerfile)
- [`v0.2.1`](https://github.com/provenance-io/provenance/blob/release/v0.2.1/docker/blockchain/Dockerfile)
- [`v0.2.0`](https://github.com/provenance-io/provenance/blob/release/v0.2.0/docker/blockchain/Dockerfile)

# Quick reference (cont.)

- **Supported architectures**: [`amd64`]
  - **Why not more?**: Upstream dependencies currently lock us into amd64 (namely libwasm). There are future plans for other architectures.
- **Source of this description**: [docs](https://github.com/provenance-io/provenance/blob/main/docker/README.md)

# What is this image?

The `provenanceio/provenance` images are used to quickly access the `provenanced` binary to run queries and transactions against the provenance testnet and mainnet.

# How to use this image

Querying an account balance

- Queries do not require access to an existing keyring.

```console
$ docker run --rm -it provenanceio/provenance provenanced -t q bank balances tp1hwrh8c3z4x4s9zl9442ee9syvyvle5ulx5jye2 --node tcp://rpc-1.test.provenance.io:26657
```

Submitting a transaction entirely within this docker image.

- Transactions require accessing an existing keyring via volume mount.
- Never use `--keyring-backend test` in a production environment! There are better ways to access your keys, than using the test backend. We use it in this example for simplicity.
- The following commands assume you have set up a keyring in the folder `${keyring_dir}/keyring-test`.
  - keyring_dir: the directory where your keyring is held.
  - from: the account within the keyring to sign and send the transaction as.

```console
# Submitting a bank transfer
$ docker run --rm -it \
    -v ${keyring_dir}/keyring-test:/home/provenance/keyring-test \
  provenanceio/provenance \
    provenanced -t \
      tx bank send \
        ${from} tp12hm0dlpz0jqm6wlm2e8mxtqvxuvwk85pxsfgnu 1000nhash \
      --node tcp://rpc-1.test.provenance.io:26657 \
      --keyring-backend test \
      --gas auto \
      --fees 5000nhash \
      --chain-id pio-testnet-1 \
      --broadcast-mode block \
      -y
```
