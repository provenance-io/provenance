#!/bin/bash
# This file can either be executed or sourced.
# If sourced, the update_toc and update_tocs functions will be added to the environment.
# If sourced, the generate_toc.sh file will also be sourced.
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

Usage: ./update-toc.sh [--force] <filename> [<filename>...]

By default, only .md files can be updated. The --force flag bypasses this restriction.

If the provided filename is a directory, all files in that directory (recursively)
that have a TOC comment will be updated.

If the provided filename contains a TOC comment line, the Table of
contents will be put there, replacing the existing one if it exists.

If the provided file does not have a TOC comment line then the table of
contents will be placed above the first heading with level 2 or greater.

If the provided file has neither a TOC comment line, nor a heading with level
2 or greater, a TOC placed at the top of the file.

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

# TOC_LOC_REGEX is a regex for finding a TOC comment.
TOC_LOC_REGEX='^[[:space:]]*<\!--[[:space:]]*TOC.*-->'

# Calls the update_toc function on multiple files or directories.
# Usage: update_toc "<filename or directory>" [<filename or directory> ...]
# See __update_toc_usage for details.
update_tocs () {
  local filenames useforce filesdone filename isdone filedone
  filenames=()
  while [[ "$#" -gt '0' ]]; do
    if [[ "$1" == '--force' ]]; then
      useforce='YES'
    elif [[ -d "$1" ]]; then
      filenames+=( $( grep -rl "$TOC_LOC_REGEX" "$( sed 's/\/$//' <<< "$1" )" 2> /dev/null | sort ) )
    elif [[ -f "$1" ]]; then
      filenames+=( "$1" )
    else
      printf 'Unknown argument: [%s]\n' "$1"
    fi
    shift
  done

  filesdone=()
  for filename in "${filenames[@]}"; do
    # Only do .md files.
    # This prevents unwanted stuff if launched like:  scripts/update-toc.sh x/*
    # The "x/*" would get expanded in the shell before getting to this script and would include all files.
    if [[ -z "$useforce" ]] && ! grep -q "\.md$" 2> /dev/null <<< "$filename"; then
      printf 'Skipping [%s] because it is not a .md file.\n' "$filename"
      continue
    fi
    # Only do a file once.
    isdone=''
    for filedone in "${filesdone[@]}"; do
      if [[ "$filename" == "$filedone" ]]; then
        isdone='YES'
        break
      fi
    done
    if [[ -n "$isdone" ]]; then
      printf 'Skipping [%s] because it is a duplicate.\n' "$filename"
      continue
    fi
    # Do it!
    printf 'Updating TOC in [%s] ... ' "$filename"
    update_toc "$filename"
    printf 'done\n'
    filesdone+=( "$filename" )
  done

  if [[ "${#filesdone[@]}" -eq '0' ]]; then
    printf 'No files were updated.\n'
    return 0
  fi

  return 0
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

  local has_toc_loc has_heading_two tempfile line toc_params toc_included in_old_toc
  has_toc_loc="$( grep -q "$TOC_LOC_REGEX" "$filename" && printf 'YES' )"
  has_heading_two="$( grep -q '^##' "$filename" && printf 'YES' )"

  tempfile="$( mktemp -t "$( sed 's/\//-/g' <<< 'x/metadata/spec/03_messages.md' )" )"

  # if there's no pre-defined TOC location, and no level 2 heading, put the TOC at the top.
  if [[ -z "$has_toc_loc" && -z "$has_heading_two" ]]; then
      printf '<!-- TOC 1 -->\n' >> "$tempfile"
      generate_toc "$filename" '1' >> "$tempfile"
      printf '\n' >> "$tempfile"
      toc_included='YES'
  fi

  while IFS= read -r line; do
    if [[ -n "$in_old_toc" ]] && ! grep -q "^[[:space:]]*- \[" <<< "$line"; then
      in_old_toc=''
    fi
    if [[ -n "$has_toc_loc" ]] && grep -q "$TOC_LOC_REGEX" <<< "$line"; then
      # Get just the TOC comment piece,
      # replace all non-digits with spaces, trim leading and trailing spaces.
      # Expected result examples: "", "3", "1 3"
      toc_params="$(
        grep -o "$TOC_LOC_REGEX" <<< "$line" \
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
  done <<< "$( cat "$filename" )"
  # Note: The above uses <<< "$( cat "$filename" )" because:
  # 1) Using just < "$filename" results in loss of the last line if the last line doesn't end in a newline.
  # 2) Starting the loop with cat "$filename" | while ... can cause weird variable behavior in the loop.

  # Using cp and rm here (instead of mv) to preserve permissions.
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

# If not sourced, do the stuff!
if [[ "$update_toc_sourced" != 'YES' ]]; then
  # Print help if no args are given.
  if [[ "$#" -eq 0 ]]; then
    __update_toc_usage
    exit 0
  fi
  # Print help if any of the args are -h, or --help (or h, --h, help, or -help)
  for a in "$@"; do
    if [[ "$a" =~ ^-?-?h(elp)?$ ]]; then
      __update_toc_usage
      exit 0
    fi
  done

  # Pull in the generate_toc function.
  __source_generate_toc || exit $?

  # Finally, Do the stuff!
  update_tocs "$@"
  exit $?
fi

# It was sourced, do some clean up and also source the generate_toc file so that function's available too.
unset update_toc_sourced
unset -f __update_toc_usage
__source_generate_toc || return $?
unset -f __source_generate_toc

return 0