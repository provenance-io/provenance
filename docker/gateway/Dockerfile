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

# Build the binaries
RUN go build -mod vendor -tags cleveldb -o /go/bin/ ./cmd/...

###
FROM debian:buster-slim

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl jq libleveldb-dev && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/

## Fixes chain panic.  Should be removed once cosmwasm has fixed their release
## TODO Remove it: https://github.com/CosmWasm/cosmwasm/pull/1223
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0-beta6/libwasmvm_muslc.a /lib/libwasmvm_muslc.a
RUN sha256sum /lib/libwasmvm_muslc.a | grep e9cb9517585ce3477905e2d4e37553d85f6eac29bdc3b9c25c37c8f5e554045c

COPY --from=build /go/src/github.com/provenance-io/provenance/vendor/github.com/CosmWasm/wasmvm/api/libwasmvm.so /lib/x86_64-linux-gnu/libwasmvm.so
COPY --from=build /go/bin/provenanced /usr/bin/provenanced

ENV PIO_HOME=/home/provenance
WORKDIR /home/provenance

EXPOSE 1317
CMD ["/usr/bin/provenanced"]
