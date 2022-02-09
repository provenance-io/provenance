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

if [[ -n "$1" && "$1" =~ ^v ]]; then
    echo "Illegal version: [$1]. Must not start with 'v'." >&2
    exit 1
fi

if [[ -n "$2" && "$2" =~ [^[:digit:]] ]]; then
    echo "Illegal jobs count: [$2]. Must only contain digits. Must be at least 1." >&2
    exit 1
fi

# In order to install the compiled libraries:
#  * For linux (at least in the github action runners), sudo is required.
#  * From inside a docker container, sudo isn't needed (and isn't even available), so we cannot use sudo there.
#  * On a mac, it ends up using brew, which complains if you use sudo (security concerns).
# So basically, if sudo is available and brew is not, use sudo for the install.
SUDO=''
if command -v sudo > /dev/null 2>&1 && ! command -v brew > /dev/null 2>&1; then
    SUDO="sudo"
fi

set -ex

ROCKS_DB_VERSION="${1:-6.28.2}"
JOBS="${2:-$( nproc )}"
TAR_FILE="v${ROCKS_DB_VERSION}.tar.gz"

[[ ! -e "$TAR_FILE" ]] || rm "$TAR_FILE"
wget "https://github.com/facebook/rocksdb/archive/refs/tags/$TAR_FILE"
tar zxf "$TAR_FILE"
ROCKS_DB_DIR="$( tar --exclude='./*/*/*' -tf "$TAR_FILE" | head -n 1 )"
cd "$ROCKS_DB_DIR"
export DEBUG_LEVEL=0
make -j${JOBS} shared_lib
$SUDO make install-shared

cd ..
rm "$TAR_FILE"
rm -rf "$ROCKS_DB_DIR"