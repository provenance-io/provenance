# This docker file builds a test image using binaries built locally.
# Use `make docker-build-local` to build or `make localnet-start` to
# start the test network and `make localnet-stop` to stop it.

## Build provenance for x86_64
FROM golang:1.17-buster as build-x86_64
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

# Build the binaries
RUN go build -mod vendor -tags=cleveldb -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/version.Name=Provenance-Blockchain' -o /go/bin/ ./cmd/...

## Run provenance for x86_64
FROM debian:buster-slim as provenance-x86_64
ENV LOCALNET=1
ENV LD_LIBRARY_PATH="/usr/local/lib"

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl jq libleveldb-dev && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/

# Requires a mounted volume with the config in it.
VOLUME [ "/provenance" ]
WORKDIR /provenance
EXPOSE 26656 26657 1317 9090

# Source binaries from the build above
COPY --from=build-x86_64 /go/src/github.com/provenance-io/provenance/vendor/github.com/CosmWasm/wasmvm/api/libwasmvm.x86_64.so /usr/local/lib/
COPY --from=build-x86_64 /go/bin/provenanced /usr/bin/provenanced

COPY networks/local/blockchain-local/entrypoint.sh /usr/bin/entrypoint.sh

STOPSIGNAL SIGTERM

ENTRYPOINT ["/usr/bin/entrypoint.sh"]
CMD ["start"]






## Build provenance for ARM
FROM golang:1.17-buster as build-arm64
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

# Build and install librdkafka
RUN git clone --depth 1 --branch v1.8.2 https://github.com/edenhill/librdkafka.git && cd librdkafka && ./configure --enable-static && make && make install

# Build the binaries
RUN go build -mod vendor -tags=cleveldb,dynamic -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/version.Name=Provenance-Blockchain' -o /go/bin/ ./cmd/...

## Run provenance for ARM
FROM debian:buster-slim as provenance-arm64
ENV LOCALNET=1
ENV LD_LIBRARY_PATH="/usr/local/lib"
# This is for M1 to find package config
ENV PKG_CONFIG_PATH="/usr/local/lib"

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl jq libleveldb-dev && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/

# Setup librdkafka
RUN apt-get update
RUN apt-get install -y pkg-config
COPY --from=build-arm64 /usr/local/include/librdkafka /usr/local/include/
COPY --from=build-arm64 /usr/local/lib/librdkafka* /usr/local/lib/

# Requires a mounted volume with the config in it.
VOLUME [ "/provenance" ]
WORKDIR /provenance
EXPOSE 26656 26657 1317 9090

# Source binaries from the build above
COPY --from=build-arm64 /go/src/github.com/provenance-io/provenance/vendor/github.com/CosmWasm/wasmvm/api/libwasmvm.aarch64.so /usr/local/lib/

COPY --from=build-arm64 /go/bin/provenanced /usr/bin/provenanced

COPY networks/local/blockchain-local/entrypoint.sh /usr/bin/entrypoint.sh

STOPSIGNAL SIGTERM

ENTRYPOINT ["/usr/bin/entrypoint.sh"]
CMD ["start"]
