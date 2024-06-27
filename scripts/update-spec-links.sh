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
    --no-clean|--no-cleanup)
      no_clean='YES'
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

########################################################################################################################
########################################  Identify files that will be updated.  ########################################
########################################################################################################################

link_rx='^\+\+\+ https://github\.com/provenance-io/provenance'
declare files=()

if [[ "${#files_in[@]}" -eq '0' && "${#dirs_in[@]}" -eq '0' ]]; then
  dirs_in+=( '.' )
fi

i=0
for file in "${files_in[@]}"; do
  i=$(( i + 1 ))
  [[ -n "$verbose" ]] && printf '[%d/%d]: Checking file: %s ... ' "$i" "${#files_in[@]}" "$file"
  if grep -E --no-messages --binary-file=without-match --silent "$link_rx" "$file" > /dev/null 2>&1; then
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

i=0
for folder in "${dirs_in[@]}"; do
  i=$(( i + 1 ))
  [[ -n "$verbose" ]] && printf '[%d/%d]: Checking directory for files with links: %s ... ' "$i" "${#dirs_in[@]}" "$folder"
  declare new_files=()
  set -o noglob
  IFS=$'\n' new_files+=( $( grep -E --files-with-matches --recursive --no-messages --binary-file=without-match "$link_rx" "$folder" | sort ) )
  set +o noglob
  if [[ "${#new_files[@]}" -ne '0' ]]; then
    [[ -n "$verbose" ]] && printf 'Found %d file(s) with links:\n' "${#new_files[@]}"
    j=0
    for file in "${new_files[@]}"; do
      j=$(( j + 1 ))
      [[ -n "$verbose" ]] && printf '[%d/%d|%d/%d]: %s' "$i" "${#dirs_in[@]}" "$j" "${#new_files[@]}" "$file"
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

if [[ "${#files[@]}" -eq '0' ]]; then
  printf 'No files found to update.\n'
  exit 0
fi

[[ -z "$quiet" ]] && printf 'Updating links in %d file(s).\n' "${#files[@]}"
[[ -n "$verbose" ]] && printf '  %s\n' "${files[@]}"

########################################################################################################################
###########################################  Identify proto files involved.  ###########################################
########################################################################################################################

[[ -n "$verbose" ]] && printf 'Identifying proto files linked to from %d files.\n' "${#files[@]}"

declare protos_linked=()
i=0
for file in "${files[@]}"; do
  i=$(( i + 1 ))
  [[ -n "$verbose" ]] && printf '[%d/%d]: Checking %s ... ' "$i" "${#files[@]}" "$file"
  declare new_proto_files=()
  set -o noglob
  IFS=$'\n' new_proto_files+=( $( grep -E "$link_rx" "$file" | sed 's|^.*/proto/|proto/|; s/#.*$//;' | sort -u ) )
  set +o noglob
  if [[ "${#new_proto_files[@]}" -ne '0' ]]; then
    [[ -n "$verbose" ]] && printf 'Found %d:\n' "${#new_proto_files[@]}" && printf '[%d/%d]:   %s\n' "$i" "${#files[@]}" "${new_proto_files[@]}"
    protos_linked+=( "${new_proto_files[@]}" )
  else
    [[ -n "$verbose" ]] && printf 'None found.\n'
  fi
done

[[ -n "$verbose" ]] && printf 'All linked protos (with dups) (%d):\n' "${#protos_linked[@]}"
[[ -n "$verbose" ]] && printf '  %s\n' "${protos_linked[@]}"

# Deduplicate entries linked from multiple files.
declare protos=()
set -o noglob
IFS=$'\n' protos+=( $( printf '%s\n' "${protos_linked[@]}" | sort -u ) )
set +o noglob

[[ -n "$verbose" ]] && printf 'All linked protos (no dups) (%d):\n' "${#protos[@]}"
[[ -n "$verbose" ]] && printf '  %s\n' "${protos[@]}"

########################################################################################################################
#####################################  Put all needed proto files in a temp dir.  ######################################
########################################################################################################################

temp_dir="$( mktemp -d -t link-updates )"
[[ -n "$verbose" ]] && printf 'Created temp dir for protos: %s\n' "$temp_dir"

# Usage: safe_exit [code]
safe_exit () {
  local ec
  ec="${1:-0}"

  if [[ -n "$temp_dir" && -d "$temp_dir" ]]; then
    if [[ -n "$no_clean" ]]; then
      [[ -z "$quiet" ]] && printf 'NOT deleting temporary directory: %s\n' "$temp_dir"
    else
      [[ -n "$verbose" ]] && printf 'Deleting temporary directory: %s\n' "$temp_dir"
      rm -rf "$temp_dir"
      temp_dir=''
    fi
  fi

  exit "$ec"
}

[[ -n "$verbose" ]] && printf 'Getting %d proto files.\n' "${#protos[@]}"

repo_root="$( git rev-parse --show-toplevel )"

i=0
for file in "${protos[@]}"; do
  i=$(( i + 1 ))
  [[ -n "$verbose" ]] && printf '[%d/%d] Getting: %s ... ' "$i" "${#protos[@]}" "$file"
  file_dir="$( dirname "$file" )"
  dest_dir="${temp_dir}/${file_dir}"
  mkdir -p "$dest_dir"

  if [[ -n "$source_ref" ]]; then
    dest_file="${temp_dir}/${file}"
    header_file="$dest_file.header"
    url="https://raw.githubusercontent.com/provenance-io/provenance/${source_ref}/$file"
    [[ -n "$verbose" ]] && printf 'From: %s\n' "$url"
    curl --silent -o "$dest_file" --dump-header "$header_file" "$url"
    if ! head -n 1 "$header_file" | grep -q '200[[:space:]]*$' > /dev/null 2>&1; then
      printf 'Source file not found: %s\n' "$url"
      stop_early='YES'
    fi
  else
    [[ -n "$verbose" ]] && printf 'From: %s\n' "$source_file"
    source_file="${repo_root}/${file}"
    cp "$source_file" "$dest_dir" || stop_early='YES'
  fi
done

if [[ -n "$stop_early" ]]; then
  safe_exit 1
fi


safe_exit 0