#!/bin/bash

# This script will download, compile, and install rocksdb.
# It downloads and unpacks a tar in the current working directory, then clean them up when done.
# Usage: build_rocksdb.sh <version>
# As of writing this (Feb 7, 2022), the current version is 6.28.2

if [[ "$1" == '-h' || "$1" == '--help' || "$1" == 'help' ]]; then
    echo "Usage: $( basename $0 ) <version> [<jobs>]"
    echo 'See https://github.com/facebook/rocksdb/releases for version info.'
    echo '<jobs> is the number of parallel jobs for make to use. The default comes from the nproc command.'
    exit 0
fi

# Order of precedence for rocksdb version: command line arg 1, env var, default.
if [[ -n "$1" ]]; then
    ROCKSDB_VERSION="$1"
elif [[ -z "$ROCKSDB_VERSION" ]]; then
    ROCKSDB_VERSION='6.28.2'
fi
if [[ -n "$ROCKSDB_VERSION" && "$ROCKSDB_VERSION" =~ ^v ]]; then
    echo "Illegal version: [$ROCKSDB_VERSION]. Must not start with 'v'." >&2
    exit 1
fi

# Order of precedence for rocksdb jobs count: command line arg 2, env var, default.
if [[ -n "$2" ]]; then
    ROCKSDB_JOBS="$2"
elif [[ -z "$ROCKSDB_JOBS" ]]; then
    ROCKSDB_JOBS="$( nproc )"
fi

if [[ -n "$ROCKSDB_JOBS" && "$ROCKSDB_JOBS" =~ [^[:digit:]] ]]; then
    echo "Illegal jobs count: [$ROCKSDB_JOBS]. Must only contain digits. Must be at least 1." >&2
    exit 1
fi

# In order to install the compiled libraries:
#  * For linux (at least in the github action runners), sudo is required.
#  * From inside a docker container, sudo isn't needed (and isn't even available), so we cannot use sudo there.
#  * On a mac, it ends up using brew, which complains if you use sudo (security concerns).
# So basically, if sudo is available and brew is not, use sudo for the install.
# This is overrideable by setting the ROCKSDB_SUDO environment variable to either 'yes' or 'no'.
SUDO=''
if [[ -n "$ROCKSDB_SUDO" ]]; then
    if [[ "$ROCKSDB_SUDO" =~ ^[yY]([eE][sS])?$ ]]; then
        SUDO='sudo'
    elif [[ ! "$ROCKSDB_SUDO" =~ ^[nN]([oO])?$ ]]; then
        echo "Illegal ROCKSDB_SUDO value: [$ROCKSDB_SUDO]. Must be either 'yes' or 'no'." >&2
        exit 1
    fi
elif command -v sudo > /dev/null 2>&1 && ! command -v brew > /dev/null 2>&1; then
    SUDO="sudo"
fi

set -ex

# These lines look dumb, but they're here so that the values are clearly in the output (because of set -e).
ROCKSDB_VERSION="$ROCKSDB_VERSION"
ROCKSDB_JOBS="$ROCKSDB_JOBS"
TAR_FILE="v${ROCKSDB_VERSION}.tar.gz"

[[ ! -e "$TAR_FILE" ]] || rm "$TAR_FILE"
wget "https://github.com/facebook/rocksdb/archive/refs/tags/$TAR_FILE"
tar zxf "$TAR_FILE"
ROCKS_DB_DIR="$( tar --exclude='./*/*/*' -tf "$TAR_FILE" | head -n 1 )"
cd "$ROCKS_DB_DIR"
export DEBUG_LEVEL=0
make -j${ROCKSDB_JOBS} shared_lib
$SUDO make install-shared

cd ..
rm "$TAR_FILE"
rm -rf "$ROCKS_DB_DIR"