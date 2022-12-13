FROM golang:1.18-bullseye as build
ARG VERSION
ARG ARCH

WORKDIR /go/src/github.com/provenance-io/provenance

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y libleveldb-dev

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


# Build and install provenanced
ENV VERSION=$VERSION
RUN make VERSION=${VERSION} install

###
FROM debian:bullseye-slim as run
ARG ARCH
ENV LD_LIBRARY_PATH="/usr/local/lib"
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl jq libleveldb-dev && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/

COPY --from=build /go/src/github.com/provenance-io/provenance/vendor/github.com/CosmWasm/wasmvm/internal/api/libwasmvm.$ARCH.so /usr/local/lib
COPY --from=build /go/bin/provenanced /usr/bin/provenanced

ENV PIO_HOME=/home/provenance
WORKDIR /home/provenance

EXPOSE 1317 9090 26656 26657
ENTRYPOINT ["/usr/bin/provenanced"]

