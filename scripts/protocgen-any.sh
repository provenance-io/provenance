#!/usr/bin/env bash

# This script generates a custom wrapper for google.protobuf.Any in
# codec/types/any.pb.go with a custom generated struct that lives in
# codec/types/any.go

set -eo pipefail

go install github.com/cosmos/gogoproto/protoc-gen-gogotypes

buf generate --template third_party/proto/buf.gen.gogo.yaml third_party/proto/google/protobuf/any.proto

exit 0

mv codec/types/google/protobuf/any.pb.go codec/types
rm -rf codec/types/third_party

# This removes the call to RegisterType in the custom generated Any wrapper
# so that only the Any type provided by gogo protobuf is registered in the
# global gogo protobuf type registry, which we aren't actually using
sed '/proto\.RegisterType/d' codec/types/any.pb.go > tmp && mv tmp codec/types/any.pb.go
