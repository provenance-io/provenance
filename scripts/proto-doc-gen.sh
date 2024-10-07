#!/usr/bin/env bash

# Prerequisites:
# - protoc: https://grpc.io/docs/protoc-installation/
# - protoc-gen-doc: https://github.com/pseudomuto/protoc-gen-doc

# Note:
# I couldn't get this to work using the docker image.
# Here's the command I think is the closest to what's needed:
#
# docker run --rm -v docs:/out -v .:/protos pseudomuto/protoc-gen-doc \
#   -I 'protos/proto' \
#   -I 'protos/third_party/proto' \
#   --doc_opt='./docs/protodoc-markdown.tmpl,proto-docs.md' \
#   $( find "proto" -name '*.proto' | sed 's/^/protos\//' )
#
# The problem with that is that the image's entrypoint invokes protoc with the -Iprotos option.
# That ends up giving protoc all the protos twice, so it complains about everything being defined twice.
# If I try to use our proto dir for the /protos volume (-v proto:/protos), it complains about not finding
# the third_party stuff. If I leave off either -I "protos/proto" or -I "protos/third_party/proto" it
# complains about things not being found.

ec=0

if ! command -v protoc > /dev/null 2>&1; then
  printf 'Command not found: protoc\n' >&2
  printf 'See: https://grpc.io/docs/protoc-installation/\n' >&2
  ec=1
fi

if ! command -v protoc-gen-doc > /dev/null 2>&1; then
  printf 'Plugin not found: protoc-gen-doc\n' >&2
  printf 'See: https://github.com/pseudomuto/protoc-gen-doc\n' >&2
  ec=1
fi

if [ "$ec" != '0' ]; then
  exit "$ec"
fi

# It's assumed that this script is in <repo root>/<some folder>/proto-doc-gen.sh.
# We need to be in <repo root> for this to work, so start where this script is and go up one.
where_i_am="$( cd "$( dirname "${BASH_SOURCE:-$0}" )" || exit $?; pwd -P )" || ec=$?
if [ -z "$where_i_am" ] || [ "$ec" != '0' ]; then
  printf 'Could not identify where this script is located: %s\n' "$0" >&2
  ec=1
fi

cd "$where_i_am/.." || ec=$?
if [ "$ec" != '0' ]; then
  printf 'Could not cd %s\n' "$where_i_am/.." >&2
  exit 1
fi

if ! [ -d ./proto ]; then
  printf 'Directory not found: ./proto\n' >&2
  ec=1
fi
if ! [ -d ./third_party ]; then
  printf 'Directory not found: ./third_party\n' >&2
  ec=1
fi
if ! [ -f ./docs/protodoc-markdown.tmpl ]; then
  printf 'File not found: ./docs/protodoc-markdown.tmpl\n' >&2
  ec=1
fi

if [ "$ec" != '0' ]; then
  printf 'Base directory . = %s\n' "`pwd`" >&2
  exit "$ec"
fi

# Generate the docs/proto-docs.md file.
protoc \
  -I "proto" \
  -I "third_party/proto" \
  --doc_out=./docs \
  --doc_opt=./docs/protodoc-markdown.tmpl,proto-docs.md \
  $( find "proto" -name '*.proto' )