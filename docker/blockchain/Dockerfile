FROM golang:1.17-buster as build
ARG VERSION
ARG ROCKSDB_VERSION
ARG ROCKSDB_JOBS

WORKDIR /go/src/github.com/provenance-io/provenance

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl jq && \
    apt-get install -y libleveldb-dev && \
    apt-get install -y gcc libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev

COPY client/ ./client/
COPY app/ ./app/
COPY go.* ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY x/ ./x/
COPY vendor/ ./vendor/
COPY scripts/rocksdb_build_and_install.sh ./scripts/
COPY Makefile sims.mk ./
COPY testutil/ ./testutil/
COPY .git/ ./.git/
COPY contrib/ ./contrib/


# Build and install rocksdb
ENV ROCKSDB_VERSION=$ROCKSDB_VERSION
ENV ROCKSDB_JOBS=$ROCKSDB_JOBS
RUN make rocksdb

# Build and install provenanced
ENV VERSION=$VERSION
RUN make VERSION=${VERSION} install

###
FROM debian:buster-slim as run

RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/

COPY --from=build /go/src/github.com/provenance-io/provenance/vendor/github.com/CosmWasm/wasmvm/api/libwasmvm.so /lib/x86_64-linux-gnu/libwasmvm.so
COPY --from=build /go/bin/provenanced /usr/bin/provenanced

ENV PIO_HOME=/home/provenance
WORKDIR /home/provenance

EXPOSE 1317 9090 26656 26657
CMD ["/usr/bin/provenanced"]

