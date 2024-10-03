#!/usr/bin/env bash

if [ "$1" == '-v' ] || [ "$1" == '--verbose' ]; then
  VERBOSE=1
fi

set -eo pipefail

mkdir -p ./tmp-swagger-gen
proto_files=$( find ./proto -type f -name '*.proto' -print0 | xargs -0 grep -El '^service +[^ ]+ +\{' )
for file in $proto_files; do
  [ -n "$VERBOSE" ] && printf 'Generating swagger file for [%s].\n' "$file"
  buf generate --template proto/buf.gen.swagger.yaml "$file"
done

[ -n "$VERBOSE" ] && printf 'Combining swagger files.\n'

swagger_files=$( find ./tmp-swagger-gen -type f -name '*.swagger.json' )
for file in $swagger_files; do
  if ! grep -Fq "\"url\": \"$file\"" ./client/docs/config.json; then
      printf '\033[93mWARNING\033[0m: "%s" not referenced in ./client/docs/config.json\n' "$file"
  fi
done

# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine ./client/docs/config.json -o ./tmp-swagger-gen/swagger-new.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# Strip buf appended Query and Service tags from the resulting swagger.
# While this isn't the cleanest approach unfortunately buf doesn't support a
# configuration to remove or prevent appending these extra tags.

grep -v '        - Query' "./tmp-swagger-gen/swagger-new.yaml" | grep -v '        - Service' > "./client/docs/swagger-ui/swagger.yaml"
# clean swagger files
[ -n "$VERBOSE" ] && printf 'Deleting ./tmp-swagger-gen\n'
rm -rf ./tmp-swagger-gen
