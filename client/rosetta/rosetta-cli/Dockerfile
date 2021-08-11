FROM golang:1.15-alpine as build

RUN apk add git gcc libc-dev bash --no-cache

ARG ROSETTA_VERSION="v0.6.7"

# build rosetta CLI
WORKDIR /rosetta
RUN git clone https://github.com/coinbase/rosetta-cli .
RUN git checkout tags/$ROSETTA_VERSION
RUN go build -o rosetta-cli ./main.go
RUN echo "Completed go build"

FROM alpine
RUN apk add gcc libc-dev python3 --no-cache

ENV PATH=$PATH:/bin

COPY --from=build /rosetta/rosetta-cli /bin/rosetta-cli

#FROM debian:buster-slim as run
#
#RUN apt-get update && \
#    apt-get upgrade -y && \
#    apt-get install -y curl jq libleveldb-dev && \
#    apt-get install python3 && \
#    apt-get clean && \
#    rm -rf /var/lib/apt/lists/