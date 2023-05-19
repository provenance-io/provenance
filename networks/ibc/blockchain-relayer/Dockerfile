## Build provenance for ARM
FROM golang:1.20-bullseye as build
WORKDIR /
ARG VERSION
RUN apt-get update && apt-get upgrade -y && apt-get install -y git
RUN git clone https://github.com/cosmos/relayer.git /relayer && cd /relayer && git checkout "$VERSION" && make install && cd / && rm -rf /relayer

## Run provenance for ARM
FROM build as relayer
ARG MNEMONIC
ENV LD_LIBRARY_PATH="/usr/local/lib"

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/

# Requires a mounted volume with the config in it.
VOLUME [ "/relayer" ]
WORKDIR /relayer
EXPOSE 26656 26657 1317 9090

COPY --from=build /go/bin/rly /usr/bin/rly

# We need to setup the relayer
RUN rly config init --home "/relayer"

# Copy over the network config
COPY networks/ibc/blockchain-relayer/config.yaml /relayer/config/config.yaml

# Generate the keys
RUN rly keys restore local default "$MNEMONIC" --home "/relayer"
RUN rly keys restore local2 default "$MNEMONIC" --home "/relayer"

COPY networks/ibc/blockchain-relayer/entrypoint.sh /usr/bin/entrypoint.sh

STOPSIGNAL SIGTERM

ENTRYPOINT ["/usr/bin/entrypoint.sh"]
CMD ["start"]
