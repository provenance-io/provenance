#!/usr/bin/env bash

verbose=''
if [[ "$1" == '-v' || "$1" == '--verbose' ]]; then
  verbose='--verbose'
fi

set -eo pipefail

mkdir -p ./tmp-swagger-gen
proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do

  # generate swagger files for the queries.
  query_file=$(find "${dir}" -maxdepth 1 -name 'query.proto')
  if [[ -n "$query_file" ]]; then
    [[ -n "$verbose" ]] && printf 'Generating swagger file for [%s].\n' "$query_file"
    buf generate --template proto/buf.gen.swagger.yaml "$query_file"
  fi
  # generate swagger files for the transactions.
  tx_file=$(find "${dir}" -maxdepth 1 -name 'tx.proto')
  if [[ -n "$tx_file" ]]; then
    [[ -n "$verbose" ]] && printf 'Generating swagger file for [%s].\n' "$tx_file"
    buf generate --template proto/buf.gen.swagger.yaml "$tx_file"
  fi
done

[[ -n "$verbose" ]] && printf 'Combining swagger files.\n'
# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine ./client/docs/config.json -o ./tmp-swagger-gen/swagger-new.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# Strip buf appended Query and Service tags from the resulting swagger.
# While this isn't the cleanest approach unfortunately buf doesn't support a
# configuration to remove or prevent appending these extra tags.

grep -v '        - Query' "./tmp-swagger-gen/swagger-new.yaml" | grep -v '        - Service' > "./client/docs/swagger-ui/swagger.yaml"
# clean swagger files
[[ -n "$verbose" ]] && printf 'Deleting ./tmp-swagger-gen\n'
rm -rf ./tmp-swagger-gen
