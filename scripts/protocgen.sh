#!/usr/bin/env bash

# see pre-requests:
# - https://grpc.io/docs/languages/go/quickstart/
# - gocosmos plugin is automatically installed during scaffolding.

set -eo pipefail

protoc_gen_gocosmos() {
  if ! grep "github.com/cosmos/gogoproto" go.mod &>/dev/null ; then
    echo -e "\tPlease run this command from somewhere inside the cosmos-sdk folder."
    return 1
  fi

  go get github.com/cosmos/gogoproto/protoc-gen-gocosmos@latest
}

protoc_gen_gocosmos

proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  proto_files=$(find "${dir}" -maxdepth 1 -name '*.proto')
  for file in $proto_files; do
    # Check if the go_package in the file is pointing to provenance
    if grep -q "option go_package.*provenance" "$file"; then
      buf generate --template proto/buf.gen.gogo.yaml "$file"
    fi
  done
done

# move proto files to the right places
cp -r github.com/provenance-io/provenance/* ./
rm -rf github.com

