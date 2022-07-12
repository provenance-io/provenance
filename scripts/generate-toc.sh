#!/bin/bash
# This file can either be executed or sourced.
# If sourced, the generate_toc function will be added to the environment.
# If executed, it will be run using the provided parameters (provide --help for usage).

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && generate_toc_sourced='YES' || generate_toc_sourced='NO'

__generate_toc_usage () {
  cat << EOF
generate-toc: Creates the table of contents for a markdown file.

Usage: ./update-toc.sh "<filename>" [<max level>|<min level> <max level>]

The filename is required.

The min level and max level are optional.
  If two are provided, the smallest is the min and the largest is the max.
  If exactly one is provided, it is treated as the max level.
    The min level is then the smaller of that max level and 2.
    I.e. Providing just "1" results in both a min and max level of 1.
  If no levels are provided, the min is 2 and max is 3.
  Both have a minimum value of 1 and a maximum value of 5 and are adjusted if needed.
  A "level" is referring to the number of pound signs in the heading markdown.

EOF
}

# Creates a table of contents as markdown lines from a provided markdown file.
# Usage: generate_toc "<filename>" [<max level>|<min level> <max level>]
# See __generate_toc_usage for details.
generate_toc () {
  local filename level
  filename="$1"
  shift
  if [[ ! -f "$filename" ]]; then
    printf 'File not found: %s\n' "$filename" >&2
    return 1
  fi

  if [[ "$#" -ge '2' ]]; then
      level1=$1
      level2=$2
      shift
      shift
  elif [[ "$#" -ge '1' ]]; then
      level2=$1
      shift
      if [[ "$level2" -lt '2' ]]; then
        level1=$level2
      else
        level1=2
      fi
  else
    level1=2
    level2=3
  fi
  if [[ "$level1" -lt '1' ]]; then
    level1=1
  elif [[ "$level1" -gt '5' ]]; then
    level1=5
  fi
  if [[ "$level2" -lt '1' ]]; then
    level2=1
  elif [[ "$level2" -gt '5' ]]; then
    level2=5
  fi
  if [[ "$level1" -gt "$level2" ]]; then
    local temp
    temp=$level1
    level1=$level2
    level2=$temp
  fi

  # Use Grep to get all heading lines with the desired levels.
  # Use sed to remove any pound below the min level.
  # Use sed to delimit the markdown from the heading, and a duplicate of the heading
  # Use awk to do some conversion on each piece and output each final entry line.
  grep -E "^#{${level1},${level2}} " "$filename" \
  | sed "s/^#\{${level1}\}/#/" \
  | sed -E 's/(#+) (.+)/\1~\2~\2/g' \
  | awk -F "~" \
      '{
          gsub(/#/,"  ",$1);
          gsub(/[^[:alnum:]_]+/,"-",$3);
          gsub(/-+$/,"",$3);
          print $1 "- [" $2 "](#" tolower($3) ")";
        }'
}

# If not sourced, do the stuff!
if [[ "$generate_toc_sourced" != 'YES' ]]; then
  if [[ "$#" -eq 0 ]]; then
    __generate_toc_usage
    exit 0
  fi

  for a in "$@"; do
    if [[ "$a" == '-h' || "$a" == '--help' || "$a" == "help" ]]; then
      __generate_toc_usage
      exit 0
    fi
  done

  generate_toc "$@"
  exit $?
fi

# It was sourced, clean up some environment stuff.
unset generate_toc_sourced
unset -f __generate_toc_usage

return 0
