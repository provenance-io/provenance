#!/usr/bin/env bash
# This script will update all the proto reference links in our spec docs.

show_usage () {
  cat << EOF
This script will update all the proto reference links in our spec docs.

Usage: ./update-spec-links.sh <ref> [--source <source ref>] [<file or dir> ...]

This script will identify the desired content of the links and update both the ref and line numbers in each link.

The <ref> can be a branch, tag, or commit. It is used for the update links and is the single part of the url after '/blob/...'

By default the currently checked-out proto files are used to identify the line numbers.
To use a different version for the source, provide the --source <source ref> option.
<source ref> can be a branch, tag, or commit.

By default, this will update all files under the current directory (recusively).
To limit this update to specific files or directories (recursively), provide them as extra arguments.

Example proto reference link line:
+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/marker.proto#L14-L25

In this example, the <ref> is '22740319ba4b3ba268b3720d4bee36d6c6b06b40'.

TODO: document how this identifies the desired content of a link.

EOF

}

declare files_in=()
declare dirs_in=()

while [[ "$#" -gt '0' ]]; do
  case "$1" in
    -h|--help|help)
      show_usage
      exit 0
      ;;
    -s|--source|--source-ref)
      if [[ -z "$2" ]]; then
        printf 'No argument provided after %s.\n' "$1"
        exit 1
      fi
      source_ref="$2"
      shift
      ;;
    -r|--ref)
      if [[ -z "$2" ]]; then
        printf 'No argument provided after %s.\n' "$1"
        exit 1
      fi
      ref="$1"
      ;;
    -q|--quiet)
      quiet='YES'
      verbose=''
      ;;
    --not-quiet)
      quiet=''
      ;;
    -v|--verbose)
      verbose='YES'
      quiet=''
      ;;
    --not-verbose)
      verbose=''
      ;;
    *)
      if [[ -z "$ref" ]]; then
        ref="$1"
      elif [[ -d "$1" ]]; then
        dirs_in+=( "$1" )
      elif [[ -f "$1" ]]; then
        files_in+=( "$1" )
      else
        printf 'File or directory not found: %s\n' "$1"
        stop_early='YES'
      fi
      ;;
  esac
  shift
done

if [[ -z "$ref" ]]; then
  printf 'No ref provided.\n'
  exit 1
fi

if [[ -n "$stop_early" ]]; then
  exit 1
fi

if [[ -n "$verbose" ]]; then
  if [[ "${#files_in[@]}" -eq '0' ]]; then
    printf 'No files provided.\n'
  else
    printf 'Files provided (%d):\n' "${#files_in[@]}"
    printf '  %s\n' "${files_in[@]}"
  fi
  if [[ "${#dirs_in[@]}" -eq '0' ]]; then
    printf 'No directories provided.\n'
  else
    printf 'Directories provided (%d):\n' "${#dirs_in[@]}"
    printf '  %s\n' "${dirs_in[@]}"
  fi
fi

declare files=()

if [[ "${#files_in[@]}" -eq '0' && "${#dirs_in[@]}" -eq '0' ]]; then
  dirs_in+=( '.' )
fi

for file in "${files_in[@]}"; do
  [[ -n "$verbose" ]] && printf 'Checking file: %s ... ' "$file"
  if grep -E --no-messages --binary-file=without-match --silent '^\+\+\+ ' "$file" > /dev/null 2>&1; then
    files+=( "$file" )
    [[ -n "$verbose" ]] && printf 'File has links.\n'
  else
    if [[ -n "$verbose" ]]; then
      printf 'No links found.\n'
    elif [[ -z "$quiet" ]]; then
     printf 'File does not have any links: %s\n' "$file"
   fi
  fi
done

for folder in "${dirs_in[@]}"; do
  [[ -n "$verbose" ]] && printf 'Checking directory for files with links: %s ... ' "$folder"
  set -o noglob
  declare new_files=()
  IFS=$'\n' new_files+=( $( grep -E --files-with-matches --recursive --no-messages --binary-file=without-match '^\+\+\+ ' "$folder" | sort ) )
  set +o noglob
  if [[ "${#new_files[@]}" -ne '0' ]]; then
    [[ -n "$verbose" ]] && printf 'Found %d file(s) with links:\n' "${#new_files[@]}"
    for file in "${new_files[@]}"; do
      [[ -n "$verbose" ]] && printf '  %s' "$file"
      if [[ "$file" =~ \.md$ ]]; then
        files+=( "$file" )
        [[ -n "$verbose" ]] && printf '\n'
      else
        [[ -n "$verbose" ]] && printf ' (ignored, not a markdown file)\n'
      fi
    done
  else
    if [[ -n "$verbose" ]]; then
      printf 'None found.\n'
    elif [[ -z "$quiet" ]]; then
      printf 'Directory does not have any files with links: %s\n' "$folder"
    fi
  fi
done

file_count="${#files[@]}"

if [[ "$file_count" -eq '0' ]]; then
  printf 'No files found to update.\n'
  exit 0
fi

[[ -z "$quiet" ]] && printf 'Updating links in %d file(s).\n' "$file_count"
