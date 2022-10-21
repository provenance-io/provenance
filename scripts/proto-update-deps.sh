#!/usr/bin/env bash
set -ex

#
# Download third_party proto files from the versions declared in go.mod
#

cat << EOF
Updates third_party Protobuf files

Usage: ./proto-update-deps.sh [dest]

The [dest] argument is the optional download destination.
  Default is {repo root}/third_party/

EOF

# This assumes that this script is located in {repo root}/scripts.
# Basically: "If an argument was provided, use that, otherwise get the full path to this repo's root and append /third_party to it."
DEST="${1:-$( cd "$( dirname "${BASH_SOURCE:-$0}" )/.."; pwd -P )/third_party}"

# Retrieve versions from go.mod (single source of truth)
# With dragonberry, we had to "bump" ics23 to v0.8.0 with a replace line. But confio doesn't have such a version yet.
# So until we can get rid of the ics23 replace line in go.mo, we need to just hard-code it in here to the most recent version they have.
CONFIO_PROTO_URL="https://raw.githubusercontent.com/confio/ics23/go/v0.7.0/proofs.proto"
GOGO_PROTO_URL="https://raw.githubusercontent.com/regen-network/protobuf/$( go list -m github.com/gogo/protobuf | sed 's:.* ::' )/gogoproto/gogo.proto"
COSMOS_PROTO_URL="raw.githubusercontent.com/cosmos/cosmos-proto/$( go list -m github.com/cosmos/cosmos-proto | sed 's:.* ::' )/proto/cosmos_proto/cosmos.proto"
COSMWASM_V1BETA1_TARBALL_URL='github.com/CosmWasm/wasmd/tarball/v0.17.0'  # Backwards compatibility. Needed to serialize/deserialize older wasmd protos.
COSMWASM_CUR_TARBALL_URL="$( go list -m github.com/CosmWasm/wasmd | sed 's:.* => ::; s: :/tarball/:;' )"
IBC_PORT_V1_QUERY_URL='https://raw.githubusercontent.com/cosmos/ibc-go/v2.3.1/proto/ibc/core/port/v1/query.proto' # Backwards compatibility.
IBC_GO_TARBALL_URL="$( go list -m github.com/cosmos/ibc-go/v5 | sed 's:.* => ::; s: :/tarball/:; s:/v5::;')"
COSMOS_TARBALL_URL="$( go list -m github.com/cosmos/cosmos-sdk | sed 's:.* => ::; s: :/tarball/:;' )"
TM_TARBALL_URL="$( go list -m github.com/tendermint/tendermint | sed 's:.* => ::; s: :/tarball/:;' )"

# gnu tar on ubuntu requires the '--wildcards' flag
tar='tar zx --strip-components 1'
if tar --version 2> /dev/null | grep GNU > /dev/null 2>&1; then
  tar="$tar --wildcards"
fi

mkdir -p "$DEST/proto"
cd "$DEST"
PROTO_EXPR='*/proto/**/*.proto'

# Refresh third_party protos
CONFIO_FILE='proto/proofs.proto'
rm -f "$CONFIO_FILE" "$CONFIO_FILE.orig"
curl -f -sSL "$CONFIO_PROTO_URL" -o "$CONFIO_FILE.orig" --create-dirs

GOGO_FILE='proto/gogoproto/gogo.proto'
rm -f "$GOGO_FILE"
curl -f -sSL "$GOGO_PROTO_URL" -o "$GOGO_FILE" --create-dirs

COSMOS_FILE='proto/cosmos_proto/cosmos.proto'
rm -f "$COSMOS_FILE"
curl -f -sSL "$COSMOS_PROTO_URL" -o "$COSMOS_FILE" --create-dirs

rm -rf 'proto/cosmwasm'
curl -f -sSL "$COSMWASM_V1BETA1_TARBALL_URL" | $tar --exclude='*/third_party' "$PROTO_EXPR"
curl -f -sSL "$COSMWASM_CUR_TARBALL_URL" | $tar --exclude='*/third_party' --exclude='*/proto/ibc' "$PROTO_EXPR"

rm -rf 'proto/ibc'
curl -f -sSL "$IBC_GO_TARBALL_URL" | $tar --exclude='*/third_party' "$PROTO_EXPR"
IBC_PORT_QUERY_FILE='proto/ibc/core/port/v1/query.proto'
if [ ! -f "$IBC_PORT_QUERY_FILE" ]; then
    curl -f -sSL "$IBC_PORT_V1_QUERY_URL" -o "$IBC_PORT_QUERY_FILE" --create-dirs
fi

rm -rf 'proto/cosmos'
curl -f -sSL "$COSMOS_TARBALL_URL" | $tar --exclude='*/third_party' --exclude='*/testutil' "$PROTO_EXPR"

rm -rf 'proto/tendermint'
curl -f -sSL "$TM_TARBALL_URL" | $tar --exclude='*/third_party' "$PROTO_EXPR"

## insert go, java package option into proofs.proto file
## Issue link: https://github.com/confio/ics23/issues/32 (instead of a simple sed we need 4 lines cause bsd sed -i is incompatible)
# See: https://github.com/koalaman/shellcheck/wiki/SC2129
{
  head -n3 "$CONFIO_FILE.orig"
  printf 'option go_package = "github.com/confio/ics23/go";\n'
  printf 'option java_package = "tech.confio.ics23";\n'
  printf 'option java_multiple_files = true;\n'
  tail -n+4 "$CONFIO_FILE.orig"
} > "$CONFIO_FILE"
rm "$CONFIO_FILE.orig"
