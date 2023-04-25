# This docker file builds a test image using binaries built locally.
# Use `make docker-build-local` to build or `make localnet-start` to
# start the test network and `make localnet-stop` to stop it.

## Build provenance for x86_64
FROM golang:1.18-bullseye as build-x86_64
WORKDIR /go/src/github.com/provenance-io/provenance
ENV GOPRIVATE=github.com/provenance-io
RUN apt-get update && apt-get upgrade -y

COPY client/ ./client/
COPY app/ ./app/
COPY go.* ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY x/ ./x/
COPY vendor/ ./vendor/

# Build the binaries
RUN go build -mod vendor -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/version.Name=Provenance-Blockchain' -o /go/bin/ ./cmd/...

## Run provenance for x86_64
FROM debian:bullseye-slim as provenance-x86_64
ENV LOCALNET=1
ENV LD_LIBRARY_PATH="/usr/local/lib"

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl jq && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/

# Requires a mounted volume with the config in it.
VOLUME [ "/provenance" ]
WORKDIR /provenance
EXPOSE 26656 26657 1317 9090 2345

# Source binaries from the build above
COPY --from=build-x86_64 /go/src/github.com/provenance-io/provenance/vendor/github.com/CosmWasm/wasmvm/internal/api/libwasmvm.x86_64.so /usr/local/lib/
COPY --from=build-x86_64 /go/bin/provenanced /usr/bin/provenanced

COPY networks/ibc/blockchain-ibc/entrypoint.sh /usr/bin/entrypoint.sh

STOPSIGNAL SIGTERM

ENTRYPOINT ["/usr/bin/entrypoint.sh"]
CMD ["start"]






## Build provenance for ARM
FROM golang:1.18-bullseye as build-arm64
WORKDIR /go/src/github.com/provenance-io/provenance
ENV GOPRIVATE=github.com/provenance-io
RUN apt-get update && apt-get upgrade -y

COPY client/ ./client/
COPY app/ ./app/
COPY go.* ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY x/ ./x/
COPY vendor/ ./vendor/

# Build the binaries
RUN go build -mod vendor -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/version.Name=Provenance-Blockchain' -o /go/bin/ ./cmd/...

## Run provenance for ARM
FROM debian:bullseye-slim as provenance-arm64
ENV LOCALNET=1
ENV LD_LIBRARY_PATH="/usr/local/lib"

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl jq && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/

# Requires a mounted volume with the config in it.
VOLUME [ "/provenance" ]
WORKDIR /provenance
EXPOSE 26656 26657 1317 9090 2346

# Source binaries from the build above
COPY --from=build-arm64 /go/src/github.com/provenance-io/provenance/vendor/github.com/CosmWasm/wasmvm/internal/api/libwasmvm.aarch64.so /usr/local/lib/

COPY --from=build-arm64 /go/bin/provenanced /usr/bin/provenanced

COPY networks/ibc/blockchain-ibc/entrypoint.sh /usr/bin/entrypoint.sh

STOPSIGNAL SIGTERM

ENTRYPOINT ["/usr/bin/entrypoint.sh"]
CMD ["start"]
