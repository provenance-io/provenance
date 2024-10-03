#!/usr/bin/env bash

# see pre-requests:
# - https://grpc.io/docs/languages/go/quickstart/
# - gocosmos plugin is automatically installed during scaffolding.

set -eo pipefail

if [ "$1" == '-v' ] || [ "$1" == '--verbose' ]; then
    VERBOSE=1
fi

# Find all of our proto files that have our go_package name.
proto_files=$( find ./proto -type f -name '*.proto' -print0 | xargs -0 grep -l 'option go_package.*provenance' )
for file in $proto_files; do
  [ "$VERBOSE" ] && printf 'Generating proto code for %s\n' "$file"
  buf generate --template proto/buf.gen.gogo.yaml "$file"
done

# move proto files to the right places
cp -r github.com/provenance-io/provenance/* ./
rm -rf github.com

