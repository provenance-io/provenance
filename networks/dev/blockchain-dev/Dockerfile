# This docker file builds a test image using binaries built locally.
# Use `make docker-build-local` to build or `make localnet-start` to
# start the test network and `make localnet-stop` to stop it.
FROM golang:1.17-buster as build
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
RUN go build -mod vendor \
    -tags cleveldb \
    -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/version.Name=Provenance-Blockchain' \
    -o /go/bin/ ./cmd/...

###
FROM debian:buster-slim
ENV LOCALNET=1

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
COPY --from=build /go/src/github.com/provenance-io/provenance/vendor/github.com/CosmWasm/wasmvm/api/libwasmvm.x86_64.so /lib/x86_64-linux-gnu/libwasmvm.x86_64.so
COPY --from=build /go/bin/provenanced /usr/bin/provenanced

COPY networks/dev/mnemonics/ /mnemonics/

ENTRYPOINT ["/usr/bin/entrypoint.sh"]
CMD ["start"]
STOPSIGNAL SIGTERM

COPY networks/dev/blockchain-dev/entrypoint.sh /usr/bin/entrypoint.sh
