#!/bin/bash

# This script will download, compile, and install leveldb, then clean up.
DEFAULT_CLEVELDB_VERSION='1.23'

can_sudo='false'
command -v sudo > /dev/null 2>&1 && can_sudo='true'

if [[ "$1" == '-h' || "$1" == '--help' || "$1" == 'help' ]]; then
    echo "Usage: $( basename $0 ) [<version>]"
    echo 'See https://github.com/facebook/leveldb/releases for version info.'
    echo 'The arguments can also be defined using environment variables:'
    echo "  CLEVELDBDB_VERSION for the <version>. Default is $DEFAULT_CLEVELDB_VERSION."
    echo 'Additional parameters definable using environment variables:'
    echo "  CLEVELDB_JOBS is the number of parallel jobs for make to use. Default is the result of nproc (=$( nproc )), or 1 if nproc isn't availble."
    echo '  CLEVELDB_DO_BUILD controls whether to build. Default is true.'
    echo '  CLEVELDB_DO_INSTALL controls whether to install. Default is true.'
    echo "  CLEVELDB_SUDO controls whether to use sudo when installing the built libraries. Default is $can_sudo."
    echo '  CLEVELDB_DO_CLEANUP controls whether to delete the downloaded and unpacked repo. Default is true.'
    exit 0
fi

# Order of precedence for leveldb version: command line arg 1, env var, default.
if [[ -n "$1" ]]; then
    CLEVELDB_VERSION="$1"
elif [[ -z "$CLEVELDB_VERSION" ]]; then
    CLEVELDB_VERSION="$DEFAULT_CLEVELDB_VERSION"
fi
if [[ -n "$CLEVELDB_VERSION" && "$CLEVELDB_VERSION" =~ ^v ]]; then
    echo "Illegal version: [$CLEVELDB_VERSION]. Must not start with 'v'." >&2
    exit 1
fi

if [[ -z "$CLEVELDB_JOBS" ]]; then
    if command -v nproc > /dev/null 2>&1; then
        CLEVELDB_JOBS="$( nproc )"
    else
        CLEVELDB_JOBS=1
    fi
fi

if [[ -n "$CLEVELDB_JOBS" && ( "$CLEVELDB_JOBS" =~ [^[:digit:]] || $CLEVELDB_JOBS -lt '1' ) ]]; then
    echo "Illegal jobs count: [$CLEVELDB_JOBS]. Must only contain digits. Must be at least 1." >&2
    exit 1
fi

# Usage: trueFalseOrDefault <variable name> <default value>
trueFalseOrDefault () {
    local k v d
    k="$1"
    v="${!1}"
    d="$2"
    if [[ -n "$v" ]]; then
        if [[ "$v" =~ ^[tT]([rR][uU][eE])?$ ]]; then
            printf 'true'
        elif [[ "$v" =~ ^[fF]([aA][lL][sS][eE])?$ ]]; then
            printf 'false'
        else
            echo "Illegal $k value: '$v'. Must be either 'true' or 'false'." >&2
            printf '%s' "$v"
            return 1
        fi
    else
        printf '%s' "$d"
    fi
    return 0
}

CLEVELDB_SUDO="$( trueFalseOrDefault CLEVELDB_SUDO "$can_sudo" )" || exit $?
CLEVELDB_DO_CLEANUP="$( trueFalseOrDefault CLEVELDB_DO_CLEANUP true )" || exit $?
CLEVELDB_DO_BUILD="$( trueFalseOrDefault CLEVELDB_DO_BUILD true )" || exit $?
CLEVELDB_DO_INSTALL="$( trueFalseOrDefault CLEVELDB_DO_INSTALL true )" || exit $?

# The github action runners need sudo when installing libraries. Brew sometimes does (even though it complains).
# In some situations, though, the sudo program isn't availble. If you've got sudo, this'll default to using it.
# You'll need sudo if the install command fails due to permissions (might manifest as a file does not exist error).
SUDO=''
if [[ "$CLEVELDB_SUDO" == 'true' ]]; then
    SUDO='sudo'
fi

# These are defined in the leveldb CMakeLists.txt file. We don't care about them for this.
LEVELDB_BUILD_TESTS="${LEVELDB_BUILD_TESTS:-OFF}"
LEVELDB_BUILD_BENCHMARKS="${LEVELDB_BUILD_BENCHMARKS:-OFF}"

set -ex

# These lines look dumb, but they're here so that the values are clearly in the output (because of set -x).
CLEVELDB_VERSION="$CLEVELDB_VERSION"
CLEVELDB_JOBS="$CLEVELDB_JOBS"
CLEVELDB_SUDO="$CLEVELDB_SUDO"
CLEVELDB_DO_CLEANUP="$CLEVELDB_DO_CLEANUP"
CLEVELDB_DO_BUILD="$CLEVELDB_DO_BUILD"
CLEVELDB_DO_INSTALL="$CLEVELDB_DO_INSTALL"
export LEVELDB_BUILD_TESTS="$LEVELDB_BUILD_TESTS"
export LEVELDB_BUILD_BENCHMARKS="$LEVELDB_BUILD_BENCHMARKS"
TAR_FILE="leveldb-${CLEVELDB_VERSION}.tar.gz"

if [[ ! -e "$TAR_FILE" ]]; then
    wget "https://github.com/google/leveldb/archive/${CLEVELDB_VERSION}.tar.gz" -O "$TAR_FILE"
    tar zxf "$TAR_FILE"
fi
TAR_DIR="$( tar --exclude='./*/*/*' -tf "$TAR_FILE" | head -n 1 )"
cd "$TAR_DIR"
[[ -d 'build' ]] || mkdir build
cd build
[[ "$CLEVELDB_DO_BUILD" == 'true' ]] && \
    cmake -DCMAKE_BUILD_TYPE=Release \
          -DLEVELDB_BUILD_TESTS="$LEVELDB_BUILD_TESTS" \
          -DLEVELDB_BUILD_BENCHMARKS="$LEVELDB_BUILD_BENCHMARKS" \
          -DBUILD_SHARED_LIBS=ON \
          .. && \
    cmake --build . -j$CLEVELDB_JOBS
[[ "$CLEVELDB_DO_INSTALL" == 'true' ]] && $SUDO cmake --install .
cd ..
cd ..
if [[ "$CLEVELDB_DO_CLEANUP" == 'true' ]]; then
    rm "$TAR_FILE"
    rm -rf "$TAR_DIR"
fi
