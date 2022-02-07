#!/bin/bash

# This script was written for use within github actions.
# It downloads and builds a provided version of rocksdb.
# A tar will be downloaded and unpacked in your current working directory.
# Usage: build_rocksdb.sh <version>
# As of writing this (Feb 7, 2022), the current version is 6.28.2

if [[ -z "$1" || "$1" == '-h' || "$1" == '--help' || "$1" == 'help' ]]; then
    echo "Usage: $( basename $0 ) <version> [<jobs>]"
    echo 'See https://github.com/facebook/rocksdb/releases for version info.'
    echo 'Default <jobs> comes from the nproc command.'
    exit 1
fi

# The github action runners have sudo, and it's required when installing rocksdb.
# But when building for docker, there is no sudo command.
# So, check if sudo is available, and use that for the installation if we can.
# The build purposefully does NOT use sudo due to security concerns.
SUDO_MAKE='make'
if command -v sudo > /dev/null 2>&1; then
    SUDO_MAKE="sudo ${SUDO_MAKE}"
fi

set -ex

ROCKS_DB_VERSION="$1"
JOBS="${2:-$( nproc )}"
wget "https://github.com/facebook/rocksdb/archive/refs/tags/v${ROCKS_DB_VERSION}.tar.gz"
tar zxf "v${ROCKS_DB_VERSION}.tar.gz"
cd "rocksdb-${ROCKS_DB_VERSION}"
export DEBUG_LEVEL=0
make -j${JOBS} shared_lib
${SUDO_MAKE} install-shared
