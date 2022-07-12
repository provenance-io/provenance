#!/usr/bin/env bash
set -ex

#
# Download third_party proto files from the versions declared in go.mod
#

cat << EOF
Updates third_party Protobuf files

Usage: ./proto-update-deps.sh [dest]

The dest is optional.
  The default location is ./third_party/proto/
  If a dest is supplied then the path becomes ./<dest>/third_party/proto/.

EOF

dir="$( cd "$( dirname "${BASH_SOURCE:-$0}" )/.."; pwd -P )"
dest="$dir${1:+/$1}"

EXT_PROTO_DIR="$dest"/third_party

echo "$EXT_PROTO_DIR"

# Retrieve versions from go.mod (single source of truth)
CONFIO_PROTO_URL=https://raw.githubusercontent.com/confio/ics23/go/$(go list -m github.com/confio/ics23/go | sed 's:.* ::')/proofs.proto
GOGO_PROTO_URL=https://raw.githubusercontent.com/regen-network/protobuf/$(go list -m github.com/gogo/protobuf | sed 's:.* ::')/gogoproto/gogo.proto
COSMOS_PROTO_URL=https://raw.githubusercontent.com/regen-network/cosmos-proto/master/cosmos.proto
COSMWASM_TARBALL_URL=github.com/CosmWasm/wasmd/tarball/v0.17.0  # Backwards compatibility. Needed to serialize/deserialize older wasmd protos.
WASMD_TARBALL_URL=$(go list -m github.com/CosmWasm/wasmd | sed 's:.* => ::' | sed 's/ /\/tarball\//')
IBC_GO_TARBALL_URL=$(go list -m github.com/cosmos/ibc-go/v2 | sed 's:.* => ::' | sed 's/\/v2//' | sed 's/ /\/tarball\//')
COSMOS_TARBALL_URL=$(go list -m github.com/cosmos/cosmos-sdk | sed 's:.* => ::' | sed 's/ /\/tarball\//')
TM_TARBALL_URL=$(go list -m github.com/tendermint/tendermint | sed 's:.* => ::' | sed 's/ /\/tarball\//')

# Download third_party protos
mkdir -p "$EXT_PROTO_DIR"/proto
#cp -r "$dir"/third_party/proto/google "$EXT_PROTO_DIR"/proto || exit $?
cd "$EXT_PROTO_DIR"
PROTO_EXPR="*/proto/**/*.proto"

# gnu tar on ubuntu requires the '--wildcards' flag
tar='tar zx --strip-components 1'
if tar --version 2> /dev/null | grep GNU > /dev/null 2>&1; then
  tar="$tar --wildcards"
fi

curl -f -sSL "$CONFIO_PROTO_URL" -o proto/proofs.proto.orig --create-dirs
curl -f -sSL "$GOGO_PROTO_URL" -o proto/gogoproto/gogo.proto --create-dirs
curl -f -sSL "$COSMOS_PROTO_URL" -o proto/cosmos_proto/cosmos.proto --create-dirs
curl -f -sSL "$COSMWASM_TARBALL_URL" | $tar --exclude="*/third_party" "$PROTO_EXPR"
curl -f -sSL "$WASMD_TARBALL_URL" | $tar --exclude="*/third_party" --exclude="*/proto/ibc" "$PROTO_EXPR"
curl -f -sSL "$IBC_GO_TARBALL_URL" | $tar --exclude="*/third_party" "$PROTO_EXPR"
curl -f -sSL "$COSMOS_TARBALL_URL" | $tar --exclude="*/third_party" --exclude="*/testutil" "$PROTO_EXPR"
curl -f -sSL "$TM_TARBALL_URL" | $tar --exclude="*/third_party" "$PROTO_EXPR"

## insert go, java package option into proofs.proto file
## Issue link: https://github.com/confio/ics23/issues/32 (instead of a simple sed we need 4 lines cause bsd sed -i is incompatible)
CONFIO_TYPES="$EXT_PROTO_DIR"/proto
head -n3 "$CONFIO_TYPES"/proofs.proto.orig > "$CONFIO_TYPES"/proofs.proto
# See: https://github.com/koalaman/shellcheck/wiki/SC2129
{
  echo 'option go_package = "github.com/confio/ics23/go";'
  echo 'option java_package = "tech.confio.ics23";'
  echo 'option java_multiple_files = true;'
  tail -n+4 "$CONFIO_TYPES"/proofs.proto.orig
} >> "$CONFIO_TYPES"/proofs.proto
rm "$CONFIO_TYPES"/proofs.proto.orig
