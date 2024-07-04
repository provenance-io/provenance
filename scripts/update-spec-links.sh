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

repo_root="$( git rev-parse --show-toplevel )" || exit $?
where_i_am="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"

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

link_rx='^\+\+\+ https://github\.com/provenance-io/provenance.*/proto/.*\.proto'
link_rx_esc="$( sed 's|\\|\\\\|g;' <<< "$link_rx" )"
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

[[ -z "$quiet" ]] && printf 'Identifying proto files linked to from %d files.\n' "${#files[@]}"

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
    if [[ -n "$verbose" ]]; then
      printf 'Found %d:\n' "${#new_proto_files[@]}"
      j=0
      for new_file in "${new_proto_files[@]}"; do
        j=$(( j + 1 ))
        printf '[%d/%d|%d/%d]: %s\n' "$i" "${#files[@]}" "$j" "${#new_proto_files[@]}" "$new_file"
      done
    fi
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

[[ -n "$verbose" ]] && printf 'All linked protos (no dups) (%d):\n' "${#protos[@]}" && printf '  %s\n' "${protos[@]}"

########################################################################################################################
#####################################  Put all needed proto files in a temp dir.  ######################################
########################################################################################################################

[[ -z "$quiet" ]] && printf 'Getting %d proto files.\n' "${#protos[@]}"

temp_dir="$( mktemp -d -t link-updates )"
[[ -n "$verbose" ]] && printf 'Created temp dir for protos: %s\n' "$temp_dir"

# safe_exit handles any needed cleanup before exiting with the provided code.
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

########################################################################################################################
#####################################  Identify Line Numbers in the Proto Files.  ######################################
########################################################################################################################

[[ -z "$quiet" ]] && printf 'Identifying line numbers for messages in the proto files.\n'

# Get a count of all the lines in a file.
# Usage: get_line_count <file>
#   or   <stuff> | get_line_count
# Note: A newline is added automatically when using a heredoc.
#       So get_line_count <<< '' will return 1 instead of the expected 0.
get_line_count () {
  # Not using wc because it won't count a line if there's no ending newline.
  awk 'END { print FNR; }' "$@"
}

# Each line of the message summary file is expected to have this format:
#     <message name>=<proto file>#L<start>-L<end>
message_summary_file="${temp_dir}/message_summary.txt"
[[ -n "$verbose" ]] && printf 'Creating summary of all message and enums: %s\n' "$message_summary_file"
find "$temp_dir" -type f -name '*.proto' -print0 | xargs -0 awk -f "${where_i_am}/identify-messages.awk" >> "$message_summary_file" || safe_exit $?
[[ -n "$verbose" ]] && printf 'Found %d messages/enums in the proto files.\n' "$( get_line_count "$message_summary_file" )"

########################################################################################################################
#################################################  Identify Endpoints  #################################################
########################################################################################################################

[[ -z "$quiet" ]] && printf 'Identifying endpoint messages in the proto files.\n'

# Each line of the endpoint summary file is expected to have this format:
#     rpc:<endpoint>:(Request|Response):<proto file>;<message name>=<proto file>
endpoint_summary_file="${temp_dir}/endpoint_summary.txt"
[[ -n "$verbose" ]] && printf 'Creating summary of all endpoint messages: %s\n' "$endpoint_summary_file"
find "$temp_dir" -type f -name '*.proto' -print0 | xargs -0 awk -f "${where_i_am}/identify-endpoints.awk" >> "$endpoint_summary_file" || safe_exit $?
[[ -n "$verbose" ]] && printf 'Found %d endpoint messages in the proto files.\n' "$( get_line_count "$endpoint_summary_file" )"



########################################################################################################################
############################################  Identify Links and Content.  #############################################
########################################################################################################################

[[ -z "$quiet" ]] && printf 'Identifying proto links and their content in %d markdown files.\n' "${#files[@]}"

# First pass, identify all the links and their content.
# The lines in the initial link info file are expected to each have one of the following formats:
#       <markdown file>:<line number>;<message name>=<proto file>
#       <markdown file>:<line number>;rpc:<endpoint>:(Request|Response):<proto file>
#       <markdown file>:<line number>;ERROR: <error message>: <context>

initial_link_info_file="${temp_dir}/initial_link_info.txt"
[[ -n "$verbose" ]] && printf 'Identifying link content, initial pass of %d files: %s\n' "${#files[@]}" "$initial_link_info_file"
i=0
for file in "${files[@]}"; do
  i=$(( i + 1 ))
  [[ -n "$verbose" ]] && printf '[%d/%d] Processing: %s\n' "$i" "${#files[@]}" "$file"
  awk -v LinkRx="$link_rx_esc" -f "${where_i_am}/identify-links.awk" "$file" >> "$initial_link_info_file" || safe_exit $?
done
rpc_count="$( grep ';rpc:' "$initial_link_info_file" | get_line_count )"
if [[ -n "$verbose" ]]; then
  printf 'Found %d links in %d files.\n' "$( get_line_count "$initial_link_info_file" )" "${#files[@]}"
  printf 'Identified the message name for %d links.\n' "$( grep -F '=' "$initial_link_info_file" | get_line_count )"
  printf 'Identified the endpoint and type for %d links.\n' "$rpc_count"
  printf 'Found %d problems.\n' "$( grep -F ';ERROR' "$initial_link_info_file" | get_line_count )"
fi
# We'll move onto the second pass even if there were errors because new errors might be found;
# this will provide us with a more complete picture of problems.

# Second pass, change all of the rpc lines with the format:
#       <markdown file>:<line number>;rpc:<endpoint>:(Request|Response):<proto file>
# To a message line with the format:
#       <markdown file>:<line number>;<message name>=<proto file>

link_info_file="${temp_dir}/link_info.txt"
[[ -n "$verbose" ]] && printf 'Identifying message name for %d links: %s\n' "$rpc_count" "$link_info_file"

i=0
while IFS="" read -r line || [[ -n "$line" ]]; do
    if [[ ! "$line" =~ ';rpc:' ]]; then
      # Not an rpc-lookup line, just pass it on.
      printf '%s\n' "$line" >> "$link_info_file"
    else
        i=$(( i + 1 ))
        [[ -n "$verbose" ]] && printf '[%d\%d]: Processing: %s\n' "$i" "$rpc_count" "$line"
        # The line has this format:
        #   <markdown file>:<line number>;rpc:<endpoint>:(Request|Response):<proto file>
        # Split the line into the line number and the rest.
        lead="$( sed 's/;.*$//' <<< "$line" )"
        [[ -n "$verbose" ]] && printf '[%d\%d]: lead: %s\n' "$i" "$rpc_count" "$lead"
        to_find="$( sed 's/^[^;]*;//' <<< "$line" )"
        [[ -n "$verbose" ]] && printf '[%d\%d]: to_find: %s\n' "$i" "$rpc_count" "$to_find"

        # Look for a line in the endpoint_summary_file that starts with to_find followed by a semi-colon.
        # The endpoint_summary_file lines have this format:
        #     rpc:<endpoint>:(Request|Response):<proto file>;<message name>=<proto file>
        found_lines="$( grep -F "${to_find};" "$endpoint_summary_file" )"
        found_lines_count="$( get_line_count <<< "$found_lines" )"
        [[ -n "$verbose" ]] && printf '[%d\%d]: found_lines (%d): %s\n' "$i" "$rpc_count" "$found_lines_count" "$found_lines"

        if [[ -z "$found_lines" || "$found_lines" =~ ^[[:space:]]*$ ]]; then
            [[ -n "$verbose" ]] && printf '[%d\%d]: Result: Error: not found.\n' "$i" "$rpc_count"
            printf '%s;ERROR: could not find endpoint message: %s\n' "$lead" "$to_find" >> "$link_info_file"
        elif [[ "$found_lines_count" -eq '1' ]]; then
            [[ -n "$verbose" ]] && printf '[%d\%d]: Result: Message identified.\n' "$i" "$rpc_count"
            message="$( sed 's/^[^;]*;//' <<< "$found_lines" )"
            printf '%s;%s\n' "$lead" "$message" >> "$link_info_file"
        else
            [[ -n "$verbose" ]] && printf '[%d\%d]: Result: Multiple messages identified.\n' "$i" "$rpc_count"
            printf '%s;ERROR: found %d endpoint messages: %s\n' "$lead" "$found_lines_count" "$to_find" >> "$link_info_file"
        fi
    fi
done < "$initial_link_info_file"

link_count="$( get_line_count "$link_info_file" )"
problem_count="$( grep -F ';ERROR' "$link_info_file" | get_line_count )"
if [[ -n "$verbose" ]]; then
  printf 'Found %d links in %d files.\n' "$link_count" "${#files[@]}"
  printf 'Know the message name for %d links.\n' "$( grep '=' "$link_info_file" | get_line_count )"
  printf 'Know the endpoint and type for %d links.\n' "$( grep ';rpc:' "$link_info_file" | get_line_count )"
  printf 'Found %d problems.\n' "$problem_count"
fi

# Now, we'll check for errors and stop if any are found.
[[ -n "$verbose" ]] && printf 'Checking for link identification errors: %s\n' "$link_info_file"
if [[ "$problem_count" -ne '0' ]]; then
    printf 'Found %d problematic links in the markdown files.\n' "$problem_count"
    grep -F ';ERROR' "$link_info_file"
    safe_exit 1
fi

########################################################################################################################
#############################################  Update the markdown files.  #############################################
########################################################################################################################

[[ -z "$quiet" ]] && printf 'Updating %d links in %d files.\n' "$link_count" "${#files[@]}"

# TODO: Look through each line of the link_info_file and update each link.
# Will use something like sed "$( printf '%d c\\\n%s' "$line_number" "$new_entry" )"$'\n' <markdown file> > <temp file>
# Then copy the temp file over the original.


safe_exit 0