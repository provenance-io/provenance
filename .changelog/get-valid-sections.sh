#!/bin/bash
# This script will output all of the valid changelog section options.
# It extracts this list from the comment at the top of the CHANGELOG.md file.
# Any double-quoted string at the start of a line in that comment will be treated as a valid section header.
# Those strings are lower-cased and spaces turned to dashes.
# E.g the line `"Bug Fixes" for any bug fixes.` will result in a valid section of "bug-fixes".
# It's assumed that they're listed in the order that they should appear as sections.

# Assume that this script is in the {repo_root}.changelog/ dir and that the CHANGELOG.md file is directly in {repo_root}.
where_i_am="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"
cl_file="$( dirname "$where_i_am" )/CHANGELOG.md"
if [[ ! -f "$cl_file" ]]; then
    printf 'Changelog file does not exist: %s\n' "$cl_file" >&2
    exit 1
fi
awk '{ if (in_com) { if (/^".*"/) { sub(/^"/,""); sub(/".*$/,""); sub(/^[[:space:]]+/,""); sub(/[[:space:]]$/,""); gsub(/[[:space:]]+/,"-"); print $0; } else if (/-->/) { exit 0; } }; if (/<!--/) { in_com=1; }; }' "$cl_file" | tr '[:upper:]' '[:lower:]'

