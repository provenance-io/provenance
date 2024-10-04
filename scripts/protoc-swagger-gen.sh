#!/usr/bin/env bash

if [ "$1" = '-v' ] || [ "$1" = '--verbose' ]; then
  VERBOSE=1
fi
proto_dir='./proto'
temp_dir='./tmp-swagger-gen'
temp_file="$temp_dir/swagger-new.yaml"
template_file='./proto/buf.gen.swagger.yaml'
config_file='./client/docs/config.json'
output_file='./client/docs/swagger-ui/swagger.yaml'

set -eo pipefail

mkdir -p "$temp_dir"
proto_files=$( find "$proto_dir" -type f -name '*.proto' -print0 | xargs -0 grep -El '^service +[^ ]+ +\{' )
for file in $proto_files; do
  [ -n "$VERBOSE" ] && printf 'Generating swagger file for [%s].\n' "$file"
  buf generate --template "$template_file" "$file"
done

[ -n "$VERBOSE" ] && printf 'Combining swagger files.\n'

# Output warnings for all generated files not referenced in the config.
# If it's not in the config, it won't be included in the final file, which is usually not what we want.
swagger_files=$( find "$temp_dir" -type f -name '*.swagger.json' )
for file in $swagger_files; do
  if ! grep -Fq "\"url\": \"$file\"" "$config_file"; then
      printf '\033[93mWARNING\033[0m: "%s" not referenced in %s\n' "$file" "$config_file"
  fi
done

# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine "$config_file" -o "$temp_file" -f yaml --continueOnConflictingPaths true --includeDefinitions true

# Strip buf appended Query and Service tags from the resulting swagger.
# While this isn't the cleanest approach unfortunately buf doesn't support a
# configuration to remove or prevent appending these extra tags.

grep -v -e '        - Query' -e '        - Service' "$temp_file" > "$output_file"
# clean swagger files
[ -n "$VERBOSE" ] && printf 'Deleting %s\n' "$temp_dir"
rm -rf "$temp_dir"
