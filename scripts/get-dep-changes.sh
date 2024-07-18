#!/bin/bash
# This script will git diff go.mod and identify the changes to it,
# outputting the info in a format ready for the changelog.

temp_dir="$( mktemp -d -t link-updates )"
[[ -n "$verbose" ]] && printf 'Created temp dir: %s\n' "$temp_dir"
full_diff_file="$temp_dir/go.mod.main.diff"
adds_file="$temp_dir/go.mod.adds"
rems_file="$temp_dir/go.mod.rems"

git diff -U0 main -- go.mod > "$full_diff_file"
grep -E '^\+' "$full_diff_file" | grep -Ev '^\+\+\+ ' | sed -E 's/^\+[[:space:]]*//; s/[[:space:]]*(\/\/.*)?$//;' | sort > "$adds_file"
grep -E '^-' "$full_diff_file" | grep -Ev '^--- ' | sed -E 's/^-[[:space:]]*//; s/[[:space:]]*(\/\/.*)?$//;' | sort > "$rems_file"
