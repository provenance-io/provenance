FROM golang:1.15-buster as build
ARG VERSION

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
RUN echo "Completed go build"

RUN go install ./cmd/provenanced

RUN echo "Initializing provenance"
# Initialize everything for provenance to run
WORKDIR /test
RUN provenanced -t --home ./run/provenanced init --chain-id=testing testing
RUN provenanced -t --home ./run/provenanced keys add validator --keyring-backend test
RUN provenanced -t --home ./run/provenanced add-genesis-root-name validator pio --keyring-backend test
RUN provenanced -t --home ./run/provenanced add-genesis-root-name validator pb --restrict=false --keyring-backend test
RUN provenanced -t --home ./run/provenanced add-genesis-root-name validator io --restrict --keyring-backend test
RUN provenanced -t --home ./run/provenanced add-genesis-root-name validator provenance --keyring-backend test
RUN provenanced -t --home ./run/provenanced add-genesis-account validator 100000000000000000000nhash --keyring-backend test
RUN provenanced -t --home ./run/provenanced gentx validator 1000000000000000nhash --keyring-backend test --chain-id=testing
RUN provenanced -t --home ./run/provenanced add-genesis-marker 100000000000000000000nhash --manager validator --access mint,burn,admin,withdraw,deposit --activate --keyring-backend test
RUN provenanced -t --home ./run/provenanced collect-gentxs

# lets see what is inside of there?
RUN ls /test/run/provenanced > tmpfile.txt ; cat tmpfile.txt

FROM alpine
RUN apk add gcc libc-dev python3 bash --no-cache

ENV PATH=$PATH:/bin

# Requires a mounted volume with the config in it.
VOLUME [ "/provenance" ]
WORKDIR /provenance
EXPOSE 26656 26657 1317 9090

RUN echo "Copying files from build"
COPY --from=build /go/src/github.com/provenance-io/provenance/vendor/github.com/CosmWasm/wasmvm/api/libwasmvm.so /lib/x86_64-linux-gnu/libwasmvm.so
COPY --from=build /go/bin/provenanced /usr/bin/provenanced
COPY --from=build /test/run/provenanced /test/run/provenanced

RUN echo "Finished copying files from build"

ENV PIO_HOME=/home/provenance
WORKDIR /rosetta
COPY ./client/rosetta/configuration ./
RUN chmod +x run_tests.sh
RUN chmod +x send_funds.sh
RUN chmod +x faucet.py

RUN chmod -R 0777 ./

EXPOSE 9090 26656 26657
