FROM golang:1.21-bullseye as build
ARG VERSION

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
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" != "x86_64" ] && [ "$ARCH" != "aarch64" ]; then \
        echo "Unsupported architecture (required: x86_64 or aarch64): $ARCH"; \
        exit 1; \
    fi && \
    echo "Building and installing provenance for Arch: $ARCH"; \
    make VERSION=${VERSION} install

###
FROM debian:bullseye-slim as run
ENV LD_LIBRARY_PATH="/usr/local/lib"
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl jq libleveldb-dev && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/
    

COPY --from=build /go/src/github.com/provenance-io/provenance/vendor/github.com/CosmWasm/wasmvm/v2/internal/api/libwasmvm.*.so /tmp
COPY --from=build /go/bin/provenanced /usr/bin/provenanced

RUN ARCH=$(uname -m) && \
    if [ "$ARCH" != "x86_64" ] && [ "$ARCH" != "aarch64" ]; then \
        echo "Unsupported architecture (required: x86_64 or aarch64): $ARCH"; \
        exit 1; \
    fi && \
    cp /"tmp/libwasmvm.$ARCH.so /usr/local/lib/." && \
    rm /tmp/libwasmvm.*.so

ENV PIO_HOME=/home/provenance
WORKDIR /home/provenance

EXPOSE 1317 9090 26656 26657
CMD ["/usr/bin/provenanced"]

