#!/bin/bash

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

TOC_LOC_REGEX="^[[:space:]]*<\!--[[:space:]]*TOC.*-->"

# Usage: __update_toc "<filename>"
__update_toc () {
  local filename
  filename="$1"
  shift

  if [[ ! -f "$filename" ]]; then
    printf 'File not found: %s\n' "$filename" >&2
    return 1
  fi

  local tempfile has_toc_loc line toc_params toc_included in_old_toc
  tempfile="$( mktemp -t "$( sed 's/\//-/g' <<< 'x/metadata/spec/03_messages.md' )" )"
  has_toc_loc="$( grep -q "$TOC_LOC_REGEX" "$filename" && printf 'YES' )"

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
      __generate_toc "$filename" $toc_params >> "$tempfile"
      toc_included='YES'
      in_old_toc='YES'
    elif [[ -z "$toc_included" && -z "$has_toc_loc" ]] && grep -q '^##' <<< "$line"; then
      printf '<!-- TOC -->\n' >> "$tempfile"
      __generate_toc "$filename" >> "$tempfile"
      printf '\n\n\n' >> "$tempfile"
      printf '%s\n' "$line" >> "$tempfile"
      toc_included='YES'
    elif [[ -z "$in_old_toc" ]]; then
      printf '%s\n' "$line" >> "$tempfile"
    fi
  done < "$filename"

  if [[ -z "$toc_included" ]]; then
      printf '<!-- TOC -->\n' >> "$tempfile"
      __generate_toc "$filename" >> "$tempfile"
      printf '\n' >> "$tempfile"
  fi

  # Using cp and rm here to preserve permissions.
  cp "$tempfile" "$filename"
  rm "$tempfile"
}

# Creates the actual table of contents text.
# Usage: __generate_toc "<filename>" [<max level>|<min level> <max level>]
# <filename> is the file to make the table of contents for
# The min and max levels are optional.
# If exactly one is provided, it is treated as the max level.
# If two are provided, the smallest is the min and the largest is the max.
# Default min level is 2.
# Default max level is 3.
# Both have a minimum value of 1 and a maximum value of 5.
# These represent the number of pound signs in the heading markdown.
__generate_toc () {
  local filename level
  filename="$1"
  shift
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
  # Use Grep to get all heading lines up to the desired level.
  # Use sed to remove any pound below the min level.
  # Use sed to delimit the markdown from the heading, and a duplicate of the heading
  # Use awk to do some conversion on each piece and output each final entry line.
  grep -E "^#{${level1},${level2}} " "$filename" \
  | sed "s/^#\{${level1}\}/#/" \
  | sed -E 's/(#+) (.+)/\1~\2~\2/g' \
  | awk -F "~" \
      '{
          gsub(/#/,"  ",$1);
          gsub(/[^[:alnum:]_]/,"-",$3);
          print $1 "- [" $2 "](#" tolower($3) ")";
        }'
}

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

__update_toc "$@"
exit $?
