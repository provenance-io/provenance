#!/bin/bash

TEMP=/tmp/librdkafka
LIB_PATH=/usr/local/lib
LIB_RDKAFKA=librdkafka.so
VERSION=v1.8.2

# Check if we have /usr/local/lib/librdkafka.so
if [ -f "$LIB_PATH/$LIB_RDKAFKA" ]; then
    echo "librdkafka is already installed"
else
    echo "Installing librdkafka..."
    # Build librdkafka
    git clone -b $VERSION https://github.com/edenhill/librdkafka.git $TEMP
    cd /tmp/librdkafka
    ./configure
    make
    make install
    cd ${PWD}

    # Cleanup
    rm -rf $TEMP
fi

# Check if we have the env varibale set
if ! tr ':' '\n' <<< "$LD_LIBRARY_PATH" | grep -xFq "$LIB_PATH"; then
    echo 'LD_LIBRARY_PATH is not set. Please set it with export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib'
else
    echo "LD_LIBRARY_PATH is already set."
fi