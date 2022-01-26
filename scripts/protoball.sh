#!/usr/bin/env bash
set -e

if [ "$1" == "" ]; then
	echo "Usage: $0 <output-file>"
	exit 1
fi

PROTO_DIR=proto
EXT_PROTO_DIR=third_party
BUILD_DIR=build

# Retrieve versions from go.mod (single source of truth)
CONFIO_PROTO_URL=https://raw.githubusercontent.com/confio/ics23/$(go list -m github.com/confio/ics23/go | sed 's:.* ::')/proofs.proto
GOGO_PROTO_URL=https://raw.githubusercontent.com/regen-network/protobuf/$(go list -m github.com/confio/ics23/go | sed 's:.* ::')/gogoproto/gogo.proto
COSMOS_PROTO_URL=https://raw.githubusercontent.com/regen-network/cosmos-proto/master/cosmos.proto
COSMWASM_TARBALL_URL=github.com/CosmWasm/wasmd/tarball/v0.17.0  # Backwards compatibility. Needed to serialize/deserialize older wasmd protos.
WASMD_TARBALL_URL=$(go list -m github.com/CosmWasm/wasmd | sed 's:.* => ::' | sed 's/ /\/tarball\//')
IBC_GO_TARBALL_URL=$(go list -m github.com/cosmos/ibc-go/v2 | sed 's:.* => ::' | sed 's/\/v2//' | sed 's/ /\/tarball\//')
COSMOS_TARBALL_URL=$(go list -m github.com/cosmos/cosmos-sdk | sed 's:.* => ::' | sed 's/ /\/tarball\//')
TM_TARBALL_URL=$(go list -m github.com/tendermint/tendermint | sed 's:.* => ::' | sed 's/ /\/tarball\//')

dir="$(pwd)"
zip="${1}"

rm -f "${zip}"

# Colorize the ouput.
red='\e[0;31m'
green='\e[0;32m'
lite_blue='\e[1;34m'
lite_red='\e[1;31m'
yellow='\e[1;33m'
off='\e[0m'

# Download third_party protos
mkdir -p "$dir/$BUILD_DIR/$EXT_PROTO_DIR/proto" || exit 1
cp -r $EXT_PROTO_DIR/proto/google $dir/$BUILD_DIR/$EXT_PROTO_DIR/proto || exit 1
cd "${dir}/$BUILD_DIR/$EXT_PROTO_DIR" || exit 1
curl -sSL $CONFIO_PROTO_URL -o proto/confio/proofs.proto.orig --create-dirs || exit 1
curl -sSL $GOGO_PROTO_URL -o proto/gogoproto/proofs.proto --create-dirs || exit 1
curl -sSL $COSMOS_PROTO_URL -o proto/cosmos_proto/cosmos.proto --create-dirs || exit 1
curl -sSL $COSMWASM_TARBALL_URL | tar zx --strip-components 1 --exclude="*/third_party" || exit 1
curl -sSL $WASMD_TARBALL_URL | tar zx --strip-components 1 --exclude="*/third_party" --exclude="*/proto/ibc" || exit 1
curl -sSL $IBC_GO_TARBALL_URL | tar zx --strip-components 1 --exclude="*/third_party" || exit 1
curl -sSL $COSMOS_TARBALL_URL | tar zx --strip-components 1 --exclude="*/third_party" --exclude="*/testutil" || exit 1
curl -sSL $TM_TARBALL_URL | tar zx --strip-components 1 --exclude="*/third_party" || exit 1

## insert go, java package option into proofs.proto file
## Issue link: https://github.com/confio/ics23/issues/32 (instead of a simple sed we need 4 lines cause bsd sed -i is incompatible)

CONFIO_TYPES=$dir/$BUILD_DIR/$EXT_PROTO_DIR/proto/confio
head -n3 $CONFIO_TYPES/proofs.proto.orig > $CONFIO_TYPES/proofs.proto
echo 'option go_package = "github.com/confio/ics23/go";' >> $CONFIO_TYPES/proofs.proto
echo 'option java_package = "tech.confio.ics23";' >> $CONFIO_TYPES/proofs.proto
echo 'option java_multiple_files = true;' >> $CONFIO_TYPES/proofs.proto
tail -n+4 $CONFIO_TYPES/proofs.proto.orig >> $CONFIO_TYPES/proofs.proto
rm $CONFIO_TYPES/proofs.proto.orig

# Include all third_party protos in the final zipball.
cd "${dir}/$BUILD_DIR/$EXT_PROTO_DIR" || exit 1
find "." -name \*.proto -print0 | while read -rd $'\0' d; do
	echo -en " * Adding ${lite_red}external${off} proto ${yellow}${d}${off} ... "
    if find "${d}" -name \*.proto | zip "${zip}" -@ >/dev/null; then
		  echo -e "[${green}OK${off}]"
	  else
		  echo -e "[${red}!!${off}]"
	fi
done

# Include all provenance protos in the final zipball.
cd "${dir}" || exit 1
find "${PROTO_DIR}" -name \*.proto -print0 | while read -rd $'\0' d; do
	echo -en " * Adding ${lite_blue}internal${off} protos ${yellow}${d}${off} ... "
	if find "${d}" -name \*.proto | zip "${zip}" -@ >/dev/null; then
    echo -e "[${green}OK${off}]"
  else
    echo -e "[${red}!!${off}]"
  fi
done

# Clean up
cd "${dir}/$BUILD_DIR"
rm -rf $EXT_PROTO_DIR

# Formatted for gh workflow action
echo
echo "::set-output name=protos::$zip"
