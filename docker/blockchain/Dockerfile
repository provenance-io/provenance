FROM golang:1.17-buster as build

WORKDIR /go/src/github.com/provenance-io/provenance

RUN apt-get update && apt-get upgrade -y && apt-get install -y libleveldb-dev

COPY client/ ./client/
COPY app/ ./app/
COPY go.* ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY x/ ./x/
COPY vendor/ ./vendor/
COPY testutil/ ./testutil/
COPY .git/ ./.git/
COPY contrib/ ./contrib/
COPY Makefile sims.mk ./

# Build the binaries
RUN go build -mod vendor \
    -tags cleveldb \
    -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/version.Name=Provenance-Blockchain' \
    -o /go/bin/ ./cmd/...

###
FROM debian:buster-slim as run

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl jq libleveldb-dev && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/

COPY --from=build /go/src/github.com/provenance-io/provenance/vendor/github.com/CosmWasm/wasmvm/api/libwasmvm.so /lib/x86_64-linux-gnu/libwasmvm.so
COPY --from=build /go/bin/provenanced /usr/bin/provenanced

ENV PIO_HOME=/home/provenance
WORKDIR /home/provenance

EXPOSE 9090 26656 26657
CMD ["/usr/bin/provenanced"]

