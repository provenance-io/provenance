#!/bin/bash

# This script will download, compile, and install rocksdb, then clean up.
DEFAULT_ROCKSDB_VERSION='6.29.5'

can_sudo='false'
command -v sudo > /dev/null 2>&1 && can_sudo='true'

if [[ "$1" == '-h' || "$1" == '--help' || "$1" == 'help' ]]; then
    echo "Usage: $( basename $0 ) [<version>]"
    echo 'See https://github.com/facebook/rocksdb/releases for version info.'
    echo 'The arguments can also be defined using environment variables:'
    echo "  ROCKSDB_VERSION for the <version>. Default is $DEFAULT_ROCKSDB_VERSION."
    echo 'Additional parameters definable using environment variables:'
    echo "  ROCKSDB_JOBS is the number of parallel jobs for make to use. Default is the result of nproc (=$( nproc )), or 1 if nproc isn't availble."
    echo '  ROCKSDB_WITH_SHARED controls whether to build and install the shared library. Default is true.'
    echo '  ROCKSDB_WITH_STATIC controls whether to build and install the static library. Default is false.'
    echo '  ROCKSDB_DO_BUILD controls whether to build. Default is true.'
    echo '  ROCKSDB_DO_INSTALL controls whether to install. Default is true.'
    echo "  ROCKSDB_SUDO controls whether to use sudo when installing the built libraries. Default is $can_sudo."
    echo '  ROCKSDB_DO_CLEANUP controls whether to delete the downloaded and unpacked repo. Default is true.'
    exit 0
fi

# Order of precedence for rocksdb version: command line arg 1, env var, default.
if [[ -n "$1" ]]; then
    ROCKSDB_VERSION="$1"
elif [[ -z "$ROCKSDB_VERSION" ]]; then
    ROCKSDB_VERSION="$DEFAULT_ROCKSDB_VERSION"
fi
if [[ -n "$ROCKSDB_VERSION" && "$ROCKSDB_VERSION" =~ ^v ]]; then
    echo "Illegal version: [$ROCKSDB_VERSION]. Must not start with 'v'." >&2
    exit 1
fi

if [[ -z "$ROCKSDB_JOBS" ]]; then
    if command -v nproc > /dev/null 2>&1; then
        ROCKSDB_JOBS="$( nproc )"
    else
        ROCKSDB_JOBS=1
    fi
fi

if [[ -n "$ROCKSDB_JOBS" && ( "$ROCKSDB_JOBS" =~ [^[:digit:]] || $ROCKSDB_JOBS -lt '1' ) ]]; then
    echo "Illegal jobs count: [$ROCKSDB_JOBS]. Must only contain digits. Must be at least 1." >&2
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

ROCKSDB_SUDO="$( trueFalseOrDefault ROCKSDB_SUDO "$can_sudo" )" || exit $?
ROCKSDB_WITH_SHARED="$( trueFalseOrDefault ROCKSDB_WITH_SHARED true )" || exit $?
ROCKSDB_WITH_STATIC="$( trueFalseOrDefault ROCKSDB_WITH_STATIC false )" || exit $?
ROCKSDB_DO_CLEANUP="$( trueFalseOrDefault ROCKSDB_DO_CLEANUP true )" || exit $?
ROCKSDB_DO_BUILD="$( trueFalseOrDefault ROCKSDB_DO_BUILD true )" || exit $?
ROCKSDB_DO_INSTALL="$( trueFalseOrDefault ROCKSDB_DO_INSTALL true )" || exit $?

# The github action runners need sudo when installing libraries. Brew sometimes does (even though it complains).
# In some situations, though, the sudo program isn't availble. If you've got sudo, this'll default to using it.
# You'll need sudo if the install command fails due to permissions (might manifest as a file does not exist error).
SUDO=''
if [[ "$ROCKSDB_SUDO" == 'true' ]]; then
    SUDO='sudo'
fi

BUILD_TARGETS=()
INSTALL_TARGETS=()
if [[ "$ROCKSDB_WITH_SHARED" == 'true' ]]; then
    BUILD_TARGETS+=( shared_lib )
    INSTALL_TARGETS+=( install-shared )
fi
if [[ "$ROCKSDB_WITH_STATIC" == 'true' ]]; then
    BUILD_TARGETS+=( static_lib )
    INSTALL_TARGETS+=( install-static )
fi

set -ex

# These lines look dumb, but they're here so that the values are clearly in the output (because of set -x).
ROCKSDB_VERSION="$ROCKSDB_VERSION"
ROCKSDB_JOBS="$ROCKSDB_JOBS"
ROCKSDB_WITH_SHARED="$ROCKSDB_WITH_SHARED"
ROCKSDB_WITH_STATIC="$ROCKSDB_WITH_STATIC"
ROCKSDB_SUDO="$ROCKSDB_SUDO"
ROCKSDB_DO_CLEANUP="$ROCKSDB_DO_CLEANUP"
ROCKSDB_DO_BUILD="$ROCKSDB_DO_BUILD"
ROCKSDB_DO_INSTALL="$ROCKSDB_DO_INSTALL"
TAR_FILE="rocksdb-${ROCKSDB_VERSION}.tar.gz"

if [[ ! -e "$TAR_FILE" ]]; then
    wget "https://github.com/facebook/rocksdb/archive/refs/tags/v${ROCKSDB_VERSION}.tar.gz" -O "$TAR_FILE"
    tar zxf "$TAR_FILE"
fi
TAR_DIR="$( tar --exclude='./*/*/*' -tf "$TAR_FILE" | head -n 1 )"
cd "$TAR_DIR"
export DEBUG_LEVEL=0
[[ "$ROCKSDB_DO_BUILD" == 'true' && "${#BUILD_TARGETS[@]}" > '0' ]] && make -j${ROCKSDB_JOBS} "${BUILD_TARGETS[@]}"
[[ "$ROCKSDB_DO_INSTALL" == 'true' && "${#INSTALL_TARGETS[@]}" > '0' ]] && $SUDO make "${INSTALL_TARGETS[@]}"
cd ..
if [[ "$ROCKSDB_DO_CLEANUP" == 'true' ]]; then
    rm "$TAR_FILE"
    rm -rf "$TAR_DIR"
fi
