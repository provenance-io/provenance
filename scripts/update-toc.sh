#!/bin/bash
# This file can either be executed or sourced.
# If sourced, the update_toc function will be added to the environment.
# If executed, it will be run using the provided parameters (provide --help for usage).
# This depends on the generate-toc.sh file being in the same directory.

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && update_toc_sourced='YES' || update_toc_sourced='NO'

__update_toc_usage () {
  cat << EOF
update-toc: Updates the table of contents in a markdown file.

Usage: ./update-toc.sh <filename>

If the provided filename contains a TOC comment line, the Table of
contents will be put there, replacing the existing one if it exists.

If the provided file does not have a TOC comment line then the table of
contents will be placed above the first heading with level 2 or greater.

If the provided file has neither a TOC comment line, nor a heading with level
2 or greater, the table of contents will be placed at the end of the file.

A TOC comment line has the following format:
<!-- TOC [<max depth>|<min depth> <max depth>] -->

The min and max depths are optional and describe which level headings to
include. For example, a depth of 2 refers to level 2 headings (that start
with 2 pound signs). If no numbers are present, the min depth is 2, and
the max is 3. If only one number is given, it is used as the max depth.
The min depth is the lesser of 2 and that number. If two numbers are given,
they are the min and max depth (in any order).

TOC comment lines should be placed on their own lines.

Default TOC comment line:
<!-- TOC -->
This is equal to:
<!-- TOC 3 -->
And also equal to:
<!-- TOC 2 3 -->

TOC comment line to index headings with levels 2, 3, and 4:
<!-- TOC 4 -->

TOC comment line to index headings with level 1 and 2:
<!-- TOC 1 2 -->

TOC comment line to index all headings:
<!-- TOC 1 5 -->

EOF
}

# Updates (or adds) a TOC in a markdown file.
# Usage: update_toc "<filename>"
# See __update_toc_usage for details.
update_toc () {
  local filename
  filename="$1"
  shift

  if [[ ! -f "$filename" ]]; then
    printf 'File not found: %s\n' "$filename" >&2
    return 1
  fi

  local toc_loc_regex tempfile has_toc_loc line toc_params toc_included in_old_toc
  toc_loc_regex="^[[:space:]]*<\!--[[:space:]]*TOC.*-->"
  tempfile="$( mktemp -t "$( sed 's/\//-/g' <<< 'x/metadata/spec/03_messages.md' )" )"
  has_toc_loc="$( grep -q "$toc_loc_regex" "$filename" && printf 'YES' )"

  while IFS= read -r line; do
    if [[ -n "$in_old_toc" ]] && ! grep -q "^[[:space:]]*- \[" <<< "$line"; then
      in_old_toc=''
    fi
    if [[ -n "$has_toc_loc" ]] && grep -q "$toc_loc_regex" <<< "$line"; then
      # Get just the TOC comment piece,
      # replace all non-digits with spaces, trim leading and trailing spaces.
      # Expected result examples: "", "3", "1 3"
      toc_params="$(
        grep -o "$toc_loc_regex" <<< "$line" \
        | sed -E 's/[^[:digit:]]+/ /g; s/^[[:space:]]+//; s/[[:space:]]+$//;'
      )"

      printf '%s\n' "$line" >> "$tempfile"
      # Note: $toc_params is specifically not in quotes so that if it's zero or two numbers, it becomes zero or two args.
      generate_toc "$filename" $toc_params >> "$tempfile"
      toc_included='YES'
      in_old_toc='YES'
    elif [[ -z "$toc_included" && -z "$has_toc_loc" ]] && grep -q '^##' <<< "$line"; then
      printf '<!-- TOC -->\n' >> "$tempfile"
      generate_toc "$filename" >> "$tempfile"
      printf '\n\n\n' >> "$tempfile"
      printf '%s\n' "$line" >> "$tempfile"
      toc_included='YES'
    elif [[ -z "$in_old_toc" ]]; then
      printf '%s\n' "$line" >> "$tempfile"
    fi
  done < "$filename"

  if [[ -z "$toc_included" ]]; then
      printf '<!-- TOC -->\n' >> "$tempfile"
      generate_toc "$filename" >> "$tempfile"
      printf '\n' >> "$tempfile"
  fi

  # Using cp and rm here to preserve permissions.
  cp "$tempfile" "$filename"
  rm "$tempfile"
}

__source_generate_toc () {
  local source_file
  source_file="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )/generate-toc.sh"
  if [[ ! -f "$source_file" ]]; then
    printf 'File not found: %s\n' "$source_file" >&2
    return 1
  fi
  source "$source_file"
}

if [[ "$update_toc_sourced" == 'YES' ]]; then
  __source_generate_toc || return $?
else
  __source_generate_toc || exit $?
fi

# If not sourced, do the stuff!
if [[ "$update_toc_sourced" != 'YES' ]]; then
  if [[ "$#" -eq 0 ]]; then
    __update_toc_usage
    exit 0
  fi

  for a in "$@"; do
    if [[ "$a" == '-h' || "$a" == '--help' || "$a" == "help" ]]; then
      __update_toc_usage
      exit 0
    fi
  done

  update_toc "$@"
  exit $?
fi

# It was sourced, clean up some environment stuff.
unset update_toc_sourced
unset -f __update_toc_usage
unset -f __source_generate_toc

return 0