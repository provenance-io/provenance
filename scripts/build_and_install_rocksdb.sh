#!/bin/bash

# This script was written for use within github actions.
# It downloads and builds a provided version of rocksdb.
# A tar will be downloaded and unpacked in your current working directory.
# Usage: build_rocksdb.sh <version>
# As of writing this (Feb 7, 2022), the current version is 6.28.2

if [[ -z "$1" || "$1" == '-h' || "$1" == '--help' || "$1" == 'help' ]]; then
    echo "Usage: $( basename $0 ) <version>"
    echo 'See https://github.com/facebook/rocksdb/releases for version info'
    exit 1
fi

set -ex

ROCKS_DB_VERSION="$1"
wget "https://github.com/facebook/rocksdb/archive/refs/tags/v${ROCKS_DB_VERSION}.tar.gz"
tar zxf "v${ROCKS_DB_VERSION}.tar.gz"
cd "rocksdb-${ROCKS_DB_VERSION}"
export DEBUG_LEVEL=0
make -j$(nproc) shared_lib
sudo make install-shared

