FROM golang:1.17-buster as build
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

# Build the binaries
RUN go build -mod vendor \
    -tags cleveldb \
    -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/version.Name=Provenance-Blockchain' \
    -o /go/bin/ ./cmd/...
RUN echo "Completed go build"

RUN echo "Initializing provenance"

# Initialize provenance to run with the default node configuration
WORKDIR /testrosetta
RUN provenanced testnet -t --v 1 -o . --starting-ip-address=0.0.0.0 --keyring-backend=test --chain-id=chain-local

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
COPY --from=build /go/src/github.com/provenance-io/provenance/vendor/github.com/CosmWasm/wasmvm/api/libwasmvm.x86_64.so /lib/x86_64-linux-gnu/libwasmvm.x86_64.so
COPY --from=build /go/bin/provenanced /usr/bin/provenanced
COPY --from=build /testrosetta /testrosetta

WORKDIR /rosetta
COPY ./client/rosetta/configuration ./
RUN chmod +x run_tests.sh
RUN chmod +x send_funds.sh
RUN chmod +x faucet.py

RUN chmod -R 0777 ./

ENV PATH=$PATH:/bin
ENV PIO_HOME="/provenance/testrosetta"