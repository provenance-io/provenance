FROM golang:1.15-buster as build
ARG VERSION

WORKDIR /go/src/github.com/provenance-io/provenance
ENV GOPRIVATE=github.com/provenance-io
RUN apt-get update && apt-get upgrade -y && apt-get install -y libleveldb-dev

COPY client/ ./client/
COPY app/ ./app/
COPY go.* ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY x/ ./x/
COPY vendor/ ./vendor/

### Not sure if we need these copys or not?
COPY testutil/ ./testutil/
COPY .git/ ./.git/
COPY contrib/ ./contrib/
COPY Makefile sims.mk ./

# Build the binaries
RUN go build -mod vendor \
    -tags cleveldb \
    -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/version.Name=Provenance-Blockchain' \
    -o /go/bin/ ./cmd/...
RUN echo "Completed go build"

RUN echo "Initializing provenance"

# Initialize everything for provenance to run
WORKDIR /testrosetta
RUN provenanced testnet -t --v 1 -o . --starting-ip-address 0.0.0.0 --keyring-backend=test --chain-id=chain-local

# generate some data for the config
RUN provenanced -t keys add fd --keyring-backend=test
RUN addr=$(provenanced -t keys show fd -a --keyring-backend=test)
RUN echo "$addr"
RUN provenanced -t add-genesis-account "$addr" 1000000000000nhash --keyring-backend=test
RUN provenanced -t gentx fd 10000000nhash --chain-id testing --keyring-backend=test

###
FROM debian:buster-slim
ENV LOCALNET=1

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl jq libleveldb-dev && \
    apt-get clean && \
    apt-get install -y python3 && \
    rm -rf /var/lib/apt/lists/

# Source binaries from the build above
COPY --from=build /go/src/github.com/provenance-io/provenance/vendor/github.com/CosmWasm/wasmvm/api/libwasmvm.so /lib/x86_64-linux-gnu/libwasmvm.so
COPY --from=build /go/bin/provenanced /usr/bin/provenanced
COPY --from=build /testrosetta /testrosetta

WORKDIR /rosetta
COPY ./client/rosetta/configuration ./
RUN chmod +x run_tests.sh
RUN chmod +x send_funds.sh
RUN chmod +x faucet.py

RUN chmod -R 0777 ./

# Requires a mounted volume with the config in it.
VOLUME [ "/provenance" ]
WORKDIR /provenance
EXPOSE 26656 26657 1317 9090

ENV PATH=$PATH:/bin
ENV PIO_HOME="/provenance/testrosetta"