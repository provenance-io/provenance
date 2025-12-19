#!/usr/bin/env bash

# see pre-requests:
# - https://grpc.io/docs/languages/go/quickstart/
# - gocosmos plugin is automatically installed during scaffolding.

set -eo pipefail

if [ "$1" = '-v' ] || [ "$1" = '--verbose' ]; then
    VERBOSE=1
fi

# Find all of our proto files that have our go_package name.
proto_files=$(find ./proto ./legacy_protos -name '*.proto' -exec grep -l 'option go_package.*provenance' {} \;)
if [ -z "$proto_files" ]; then
    echo "No proto files found with go_package containing 'provenance'"
    exit 0
fi

echo "Found $(echo "$proto_files" | wc -l) proto files to generate"

for file in $proto_files; do
  [ "$VERBOSE" ] && printf 'Generating proto code for %s\n' "$file"
  buf generate --template proto/buf.gen.gogo.yaml "$file"
done

# Move proto files if generated under GOPATH-style paths (local dev)
if [ -d "github.com/provenance-io/provenance" ]; then
    echo "Detected GOPATH-style generation output..."
    cp -r github.com/provenance-io/provenance/* ./
    rm -rf github.com
    echo "Proto generation completed successfully."
else
    echo "No github.com directory found; assuming buf output is already in the correct place."
    echo "Proto generation completed successfully."
fi


