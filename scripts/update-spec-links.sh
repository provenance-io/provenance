#!/usr/bin/env bash
# This script will update all the proto reference links in our spec docs to have a new <ref> and correct line numbers.
# The following scripts must be in the same directory as this one:
#   identify-links.awk
#   identify-endpoints.awk
#   identify-messages.awk

simple_usage='Usage: ./update-spec-links.sh <ref> [<flags>] [<file or dir> ...]'

show_usage () {
  cat << EOF
This script will update all the proto reference links in our spec docs.

$simple_usage
[<flags>]: [--help] [--source (<source ref>|--ref)|--source-is-ref] [--no-clean] [--quiet|--verbose]

This script will identify the desired content of the links and update both the ref and line numbers in each link.

The <ref> is the branch, tag, or commit that the updated links should have.
It is the part of the url after '/blob/' but before the file's path.
Example <ref> values and the a link they might produce:
  Tag: 'v1.19.0' -> https://github.com/provenance-io/provenance/blob/v1.19.0/proto/provenance/marker/v1/marker.proto#L14-L26
  Branch: 'release/v1.19.x' -> https://github.com/provenance-io/provenance/blob/release/v1.19.x/proto/provenance/marker/v1/marker.proto#L14-L26
  Commit: '17b9b8c3e655b6a56840841b83252b241d01ff14' -> https://github.com/provenance-io/provenance/blob/17b9b8c3e655b6a56840841b83252b241d01ff14/proto/provenance/marker/v1/marker.proto#L14-L26
The first non-flag argument is taken to be the <ref>, but it can also be preceded by the --ref flag if desired.

By default the currently checked-out proto files are used to identify message line numbers and endpoint messages.
To use a different version as the source for this identification, provide the --source <source ref> option.
The <source ref> can be a branch, tag, or commit.
If you want to use the same <source ref> as the <ref>, you can either provide the <source ref> as "--ref"
(e.g. with the args: --source --ref) or with just the --source-is-ref flag.
Keep in mind that using the same source and ref means that the ref must already be available for download.

By default, this will update all files under the current directory (recursively).
To limit this update to specific files or directories (recursively), provide them as extra arguments.

To reduce output, you can provide the --quiet or -q flag.
For extra processing output, you can provide the --verbose or -v flag.
If multiple of --quiet -q --verbose and/or -v are given, the one last provided is used.

A temporary directory is used for the source proto files and some helper/processing files.
By default, the source proto files will be copies of what's currently in your repo locally.
If a <source ref> is identified, curl will be used to download those proto files from github.

By default, the temp directory is deleted when the script ends (either successfully or with an error).
To keep the directory around regardless of outcome, supply the --no-clean flag.
To keep the directory around only in the case of errors, supply the --no-clean-on-error flag.
To always delete the directory (same as default behavior), supply the --always-clean flag.
These can also be provided as --clean never --clean ok or --clean always respectively.
If multiple --no-clean --no-clean-on-error --always-clean and/or --clean flags are provided, the last one is used.

Example proto reference link line:
+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/marker.proto#L14-L25

In this example, the <ref> is '22740319ba4b3ba268b3720d4bee36d6c6b06b40'.

For documentation on how the desired content of a link is determined, see identify-links.awk.

EOF

}

# This script requires a few other scripts that must be in the same directory.
# To consistently find them, we'll need to know the absolute path to the dir with this script.
where_i_am="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"

# Define paths to each of the required scripts.
# Note that each of these can alternatively be defined as environment variables
# primarily so that it's easier to test/develop alternative versions of each one.
identify_messages_awk="${IDENTIFY_MESSAGES_AWK:-${where_i_am}/identify-messages.awk}"
identify_endpoints_awk="${IDENTIFY_ENDPOINTS_AWK:-${where_i_am}/identify-endpoints.awk}"
identify_links_awk="${IDENTIFY_LINKS_AWK:-${where_i_am}/identify-links.awk}"

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
    --source-is-ref|--source--ref)
      source_ref='--ref'
      ;;
    -r|--ref)
      if [[ -z "$2" ]]; then
        printf 'No argument provided after %s.\n' "$1"
        exit 1
      fi
      if [[ -n "$ref" ]]; then
        printf 'two <ref>s provided: %s and %s\n' "$ref" "$2"
        exit 1
      fi
      ref="$2"
      shift
      ;;
    -q|--quiet)
      quiet='YES'
      verbose=''
      ;;
    -v|--verbose)
      verbose='YES'
      quiet=''
      ;;
    --output)
      if [[ -z "$2" ]]; then
        printf 'No argument provided after %s.\n' "$1"
        exit 1
      fi
      case "$2" in
        q|quiet)
          quiet='YES'
          verbose=''
          ;;
        v|verbose)
          quiet=''
          verbose='YES'
          ;;
        n|normal|d|default)
          quiet=''
          verbose=''
          ;;
        *)
          printf 'Unknown %s value: "%s". Must be one of: "quiet", "verbose", or "normal".\n' "$1" "$2"
          exit 1
          ;;
      esac
      shift
      ;;
    --no-clean|--no-cleanup)
      clean='never'
      ;;
    --no-clean-on-error|--no-cleanup-on-error)
      clean='ok'
      ;;
    --always-clean|--always-cleanup)
      clean='always'
      ;;
    --clean|--cleanup)
      if [[ -z "$2" ]]; then
        printf 'No argument provided after %s.\n' "$1"
        exit 1
      elif [[ "$2" != 'never' && "$2" != 'ok' && "$2" != 'always' ]]; then
        printf 'Unknown %s value: "%s". Must be one of: "never", "ok", or "always".\n' "$1" "$2"
        exit 1
      fi
      clean="$2"
      shift
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
        exit 1
      fi
      ;;
  esac
  shift
done

if [[ -z "$ref" ]]; then
  printf '%s\n' "$simple_usage"
  printf 'No ref provided.\n'
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

if [[ "$source_ref" == '--ref' ]]; then
  source_ref="$ref"
fi

if [[ ! -f "$identify_messages_awk" ]]; then
  printf 'Could not find the identify-messages awk script: %s\n' "$identify_messages_awk"
  stop_early='1'
fi
if [[ ! -f "$identify_endpoints_awk" ]]; then
  printf 'Could not find the identify-endpoints awk script: %s\n' "$identify_endpoints_awk"
  stop_early='1'
fi
if [[ ! -f "$identify_links_awk" ]]; then
  printf 'Could not find the identify-links awk script: %s\n' "$identify_links_awk"
  stop_early='1'
fi

if [[ -n "$stop_early" ]]; then
  exit 1
fi


########################################################################################################################
########################################  Identify files that will be updated.  ########################################
########################################################################################################################

link_prefix="+++ https://github.com/provenance-io/provenance/blob/$ref/"
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
      [[ -n "$verbose" ]] && printf '[%d/%d|%d/%d]:   %s' "$i" "${#dirs_in[@]}" "$j" "${#new_files[@]}" "$file"
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

[[ -z "$quiet" ]] && printf 'Identifying proto files linked to from %d markdown files.\n' "${#files[@]}"

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
        printf '[%d/%d|%d/%d]:   %s\n' "$i" "${#files[@]}" "$j" "${#new_proto_files[@]}" "$new_file"
      done
    fi
    protos_linked+=( "${new_proto_files[@]}" )
  else
    [[ -n "$verbose" ]] && printf 'None found.\n'
  fi
done

[[ -n "$verbose" ]] && printf 'All linked protos (with dups) (%d):\n' "${#protos_linked[@]}" && printf '  %s\n' "${protos_linked[@]}"

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

if [[ -z "$source_ref" ]]; then
  # If there isn't a source ref, then we're copying the proto files from the local version.
  # But all we'll know is the path relative to the repo root. So, below, we'll convert that
  # relative path to an absolute path for the cp command so that this script from any directory
  # in this repo. We're doing this before trying to create the temp directory so that we don't
  # have to worry about the extra cleanup if this fails.
  repo_root="$( git rev-parse --show-toplevel )" || exit $?
fi

temp_dir="$( mktemp -d -t link-updates )"
[[ -n "$verbose" ]] && printf 'Created temp dir for protos: %s\n' "$temp_dir"

# clean_exit handles any needed cleanup before exiting with the provided code.
# Usage: clean_exit [code]
clean_exit () {
  local ec
  ec="${1:-0}"

  if [[ -n "$temp_dir" && -d "$temp_dir" ]]; then
    if [[ "$clean" == 'always' || ( "$clean" == 'ok' && "$ec" -eq '0' ) ]]; then
      [[ -n "$verbose" ]] && printf 'Deleting temporary directory: %s\n' "$temp_dir"
      rm -rf "$temp_dir"
      temp_dir=''
    else
      [[ -z "$quiet" ]] && printf 'NOT deleting temporary directory: %s\n' "$temp_dir"
    fi
  fi

  exit "$ec"
}

i=0
for file in "${protos[@]}"; do
  i=$(( i + 1 ))
  [[ -n "$verbose" ]] && printf '[%d/%d]: Getting: %s ... ' "$i" "${#protos[@]}" "$file"
  file_dir="$( dirname "$file" )"
  dest_dir="${temp_dir}/${file_dir}"
  if ! mkdir -p "$dest_dir"; then
    printf 'ERROR: Command failed: mkdir -p "%s"\n' "$dest_dir"
    clean_exit 1
  fi

  if [[ -n "$source_ref" ]]; then
    dest_file="${temp_dir}/${file}"
    header_file="$dest_file.header"
    url="https://raw.githubusercontent.com/provenance-io/provenance/${source_ref}/$file"
    [[ -n "$verbose" ]] && printf 'From url: %s\n' "$url"
    curl --silent -o "$dest_file" --dump-header "$header_file" "$url"
    if ! head -n 1 "$header_file" | grep -q '200[[:space:]]*$' > /dev/null 2>&1; then
      printf 'ERROR: Source file not found: %s\n' "$url"
      stop_early='YES'
    fi
  else
    # We know the path relative to the repo's root, convert it to an absolute path.
    # Otherwise, this script would only work if run from the repo's root.
    source_file="${repo_root}/${file}"
    [[ -n "$verbose" ]] && printf 'From file: %s\n' "$source_file"
    if ! cp "$source_file" "$dest_dir"; then
      printf 'ERROR: Command failed: cp "%s" "%s"\n' "$source_file" "$dest_dir"
      stop_early='YES'
    fi
  fi
done

if [[ -n "$stop_early" ]]; then
  clean_exit 1
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
#     ;<message name>=<proto file>#L<start>-L<end>
message_summary_file="${temp_dir}/message-summary.txt"
[[ -n "$verbose" ]] && printf 'Creating summary of all message and enums: %s\n' "$message_summary_file"
find "$temp_dir" -type f -name '*.proto' -print0 | xargs -0 awk -f "$identify_messages_awk" >> "$message_summary_file" || clean_exit $?
[[ -n "$verbose" ]] && printf 'Found %d messages/enums in the proto files.\n' "$( get_line_count "$message_summary_file" )"

########################################################################################################################
#######################################  Identify Endpoints in the Proto Files.  #######################################
########################################################################################################################

[[ -z "$quiet" ]] && printf 'Identifying endpoint messages in the proto files.\n'

# Each line of the endpoint summary file is expected to have this format:
#     rpc:<endpoint>:(Request|Response):<proto file>;<message name>=<proto file>
endpoint_summary_file="${temp_dir}/endpoint-summary.txt"
[[ -n "$verbose" ]] && printf 'Creating summary of all endpoint messages: %s\n' "$endpoint_summary_file"
find "$temp_dir" -type f -name '*.proto' -print0 | xargs -0 awk -f "$identify_endpoints_awk" >> "$endpoint_summary_file" || clean_exit $?
[[ -n "$verbose" ]] && printf 'Found %d endpoint messages in the proto files.\n' "$( get_line_count "$endpoint_summary_file" )"

########################################################################################################################
#################################  Identify Links and Content in the Markdown Files.  ##################################
########################################################################################################################

[[ -z "$quiet" ]] && printf 'Identifying proto links and their content in %d markdown files.\n' "${#files[@]}"

# This is done in three steps:
# 1. Process the markdown files and identify all links and their content.
# 2. Convert all endpoint entries into message entries.
# 3. Identify the correct line numbers, and create the updated links.
#
# At each step, it's possible for new errors to be introduced.
# However, each step will pass previous errors along, and errors will only be looked for after the third step.
# This way, if there are errors, we get more of them all at once instead of having to fix a bunch before finding
# out if there are more. The fix for these errors will usually involve updating the proto files.

# First step: Process the markdown files and identify all links and their content.
# The lines in the initial link info file are expected to each have one of the following formats:
#       <markdown file>:<line number>;<message name>=<proto file>
#       <markdown file>:<line number>;rpc:<endpoint>:(Request|Response):<proto file>
#       <markdown file>:<line number>;ERROR: <error message>: <context>

link_info_file="${temp_dir}/1-link-info.txt"
[[ -n "$verbose" ]] && printf 'Identifying link content in %d files: %s\n' "${#files[@]}" "$link_info_file"
i=0
for file in "${files[@]}"; do
  i=$(( i + 1 ))
  [[ -n "$verbose" ]] && printf '[1:%d/%d]: Processing: %s\n' "$i" "${#files[@]}" "$file"
  awk -v LinkRx="$link_rx_esc" -f "$identify_links_awk" "$file" >> "$link_info_file" || clean_exit $?
done

rpc_count="$( grep ';rpc:' "$link_info_file" | get_line_count )"
if [[ -n "$verbose" ]]; then
  printf 'Found %d links in %d files.\n' "$( get_line_count "$link_info_file" )" "${#files[@]}"
  printf 'Identified the message name for %d links.\n' "$( grep -F '=' "$link_info_file" | get_line_count )"
  printf 'Identified the endpoint and type for %d links.\n' "$rpc_count"
  printf 'Found %d problems.\n' "$( grep -F ';ERROR' "$link_info_file" | get_line_count )"
fi

# Second step: Convert all endpoint entries into message entries.
# Basically, change all of the rpc lines with the format:
#       <markdown file>:<line number>;rpc:<endpoint>:(Request|Response):<proto file>
# To a message line with the format:
#       <markdown file>:<line number>;<message name>=<proto file>
# Lines with other formats are passed forward untouched.

link_message_info_file="${temp_dir}/2-link-message-info.txt"
[[ -n "$verbose" ]] && printf 'Identifying message name for %d links: %s\n' "$rpc_count" "$link_message_info_file"

i=0
while IFS="" read -r line || [[ -n "$line" ]]; do
  if [[ ! "$line" =~ ';rpc:' ]]; then
    # Not an rpc-lookup line, just pass it on.
    printf '%s\n' "$line" >> "$link_message_info_file"
    continue
  fi

  i=$(( i + 1 ))
  [[ -n "$verbose" ]] && printf '[2:%d/%d]: Processing: %s\n' "$i" "$rpc_count" "$line"
  # The line has this format:
  #   <markdown file>:<line number>;rpc:<endpoint>:(Request|Response):<proto file>
  # Split the line into:
  #   The lead: <markdown file>:<line number>
  #   And the endpoint lookup string: rpc:<endpoint>:(Request|Response):<proto file>
  lead="$( sed 's/;.*$//' <<< "$line" )"
  [[ -n "$verbose" ]] && printf '[2:%d/%d]: lead: %s\n' "$i" "$rpc_count" "$lead"
  to_find="$( sed 's/^[^;]*;//' <<< "$line" )"
  [[ -n "$verbose" ]] && printf '[2:%d/%d]: to_find: %s\n' "$i" "$rpc_count" "$to_find"
  if [[ -z "$lead" || -z "$to_find" || "$lead" == "$to_find" ]]; then
    [[ -n "$verbose" ]] && printf '[2:%d/%d]: Result: Error: Could not split lead and to_find.\n' "$i" "$ok_count"
    printf '%s;ERROR: could not parse as endpoint lookup line\n' "$line" >> "$link_message_info_file"
    continue
  fi

  # Look for a line in the endpoint_summary_file that starts with to_find followed by a semi-colon.
  # The endpoint_summary_file lines have this format:
  #     rpc:<endpoint>:(Request|Response):<proto file>;<message name>=<proto file>
  found_lines="$( grep -F "${to_find};" "$endpoint_summary_file" )"
  found_lines_count="$( get_line_count <<< "$found_lines" )"
  [[ -z "$found_lines" || "$found_lines" =~ ^[[:space:]]*$ ]] && found_lines_count='0'
  [[ -n "$verbose" ]] && printf '[2:%d/%d]: found_lines (%d): %q\n' "$i" "$rpc_count" "$found_lines_count" "$found_lines"

  if [[ "$found_lines_count" -eq '1' ]]; then
    # Extract the combined message and file from what we found: <message name>=<proto file>
    message="$( sed -E 's/^[^;]+;//' <<< "$found_lines" )"
    [[ -n "$verbose" ]] && printf '[2:%d/%d]: message: %s\n' "$i" "$rpc_count" "$message"
    if [[ -n "$message" && "$message" != "$found_lines" ]]; then
      [[ -n "$verbose" ]] && printf '[2:%d/%d]: Result: Success: Message identified.\n' "$i" "$rpc_count"
      printf '%s;%s\n' "$lead" "$message" >> "$link_message_info_file"
    else
      [[ -n "$verbose" ]] && printf '[2:%d/%d]: Result: Error: Could not parse [%s].\n' "$i" "$rpc_count" "$found_lines"
      printf '%s;ERROR: could not parse message from %s for %s\n' "$lead" "$found_lines" "$to_find" >> "$link_message_info_file"
    fi
  elif [[ "$found_lines_count" -eq '0' ]]; then
    [[ -n "$verbose" ]] && printf '[2:%d/%d]: Result: Error: Not found.\n' "$i" "$rpc_count"
    # If you get this error, check that the section has the correct endpoint name and not the name of a message.
    printf '%s;ERROR: could not find endpoint message: %s\n' "$lead" "$to_find" >> "$link_message_info_file"
  else
    [[ -n "$verbose" ]] && printf '[2:%d/%d]: Result: Error: Multiple messages identified.\n' "$i" "$rpc_count"
    printf '%s;ERROR: found %d endpoint messages: %s\n' "$lead" "$found_lines_count" "$to_find" >> "$link_message_info_file"
  fi
done < "$link_info_file"

link_count="$( get_line_count "$link_message_info_file" )"
problem_count="$( grep -F ';ERROR' "$link_message_info_file" | get_line_count )"
if [[ -n "$verbose" ]]; then
  printf 'Found %d links in %d files.\n' "$link_count" "${#files[@]}"
  printf 'Found the message name for %d links.\n' "$( grep '=' "$link_message_info_file" | get_line_count )"
  printf 'Found %d problems.\n' "$problem_count"
fi

# Third step: Identify the correct line numbers, and create the updated links.
# Basically convert all of these lines:
#     <markdown file>:<line number>;<message name>=<proto file>
# into this format:
#     <markdown file>:<line number>;<updated link line>
# where <updated link line> has this format:
#     +++ <link prefix><proto file>#L<start>-L<end>
# Lines with other formats are passed forward untouched.

ok_count=$(( link_count - problem_count ))
new_links_file="${temp_dir}/3-new-links.txt"
[[ -n "$verbose" ]] && printf 'Creating %d new links: %s\n' "$ok_count" "$new_links_file"

i=0
while IFS="" read -r line || [[ -n "$line" ]]; do
  if [[ "$line" =~ ';ERROR' ]]; then
    # Pass on previous errors.
    printf '%s\n' "$line" >> "$new_links_file"
    continue
  fi

  i=$(( i + 1 ))
  [[ -n "$verbose" ]] && printf '[3:%d/%d]: Processing: %s\n' "$i" "$ok_count" "$line"

  # All lines here should have this format:
  #     <markdown file>:<line number>;<message name>=<proto file>
  # Split the line into:
  #   The lead: <markdown file>:<line number>
  #   The message and file to find: <message name>=<proto file>
  lead="$( sed -E 's/;.*$//' <<< "$line" )"
  [[ -n "$verbose" ]] && printf '[3:%d/%d]: lead: %s\n' "$i" "$ok_count" "$lead"
  to_find="$( sed -E 's/^[^;]+;//' <<< "$line" )"
  [[ -n "$verbose" ]] && printf '[3:%d/%d]: to_find: %s\n' "$i" "$ok_count" "$to_find"
  if [[ -z "$lead" || -z "$to_find" || "$lead" == "$to_find" ]]; then
    [[ -n "$verbose" ]] && printf '[3:%d/%d]: Result: Error: Could not split lead and to_find.\n' "$i" "$ok_count"
    printf '%s;ERROR: could not parse as link info line\n' "$line" >> "$new_links_file"
    continue
  fi

  # Lines in the message_summary_file have this format:
  #     ;<message name>=<proto file>#L<start>-L<end>
  # And in the "$to_find" variable, we have this format:
  #     <message name>=<proto file>
  # The ; is used so that I have specific characters bounding the <message name>.
  # This lets me provide "$to_find" to grep with the -F flag while only matching on
  # the exact <message name>, without also matching other message names that have this one as a prefix.
  # E.g. "Order=proto/..." needs to not also match "AskOrder=proto/...".
  # Without the ;, I'd have to omit the -F flag, escape stuff in "$to_find" and add a ^ to the front of the regex.
  # But I like to avoid such escaping if possible, so I went with the ; route in here.
  # The # is needed in the link anyway, but also provides a handy ending bound on "$to_find".

  found_lines="$( grep -F ";${to_find}#" "$message_summary_file" )"
  found_lines_count="$( get_line_count <<< "$found_lines" )"
  [[ -z "$found_lines" || "$found_lines" =~ ^[[:space:]]*$ ]] && found_lines_count='0'
  [[ -n "$verbose" ]] && printf '[3:%d/%d]: found_lines (%d): %q\n' "$i" "$ok_count" "$found_lines_count" "$found_lines"

  if [[ "$found_lines_count" -eq '1' ]]; then
    # A found line will have this format: ;<message name>=<proto file>#L<start>-L<end>
    # Extract the relative link: <proto file>#L<start>-L<end>
    relative_link="$( sed -E 's/^[^=]+=//' <<< "$found_lines" )"
    [[ -n "$verbose" ]] && printf '[3:%d/%d]: relative_link: %s\n' "$i" "$ok_count" "$relative_link"
    if [[ -n "$relative_link" && "$found_lines" != "$relative_link" ]]; then
      [[ -n "$verbose" ]] && printf '[3:%d/%d]: Result: Success: New link created.\n' "$i" "$ok_count"
      printf '%s;%s\n' "$lead" "${link_prefix}${relative_link}" >> "$new_links_file"
    else
      [[ -n "$verbose" ]] && printf '[3:%d/%d]: Result: Error: Could not parse [%s].\n' "$i" "$ok_count" "$found_lines"
      printf '%s;ERROR: could not parse relative link for %s from %s\n' "$lead" "$to_find" "$found_lines" >> "$new_links_file"
    fi
  elif [[ "$found_lines_count" -eq '0' ]]; then
    [[ -n "$verbose" ]] && printf '[3:%d/%d]: Result: Error: Not found.\n' "$i" "$ok_count"
    # If you get this error, the <message name> here probably does not exist. You'll either want to update the
    # section header that the link is in, or else add a link comment.
    printf '%s;ERROR: could not find message line numbers: %s\n' "$lead" "$to_find" >> "$new_links_file"
  else
    [[ -n "$verbose" ]] && printf '[3:%d/%d]: Result: Error: Multiple message line number entries identified.\n' "$i" "$ok_count"
    printf '%s;ERROR: found %d message line number entries: %s\n' "$lead" "$found_lines_count" "$to_find" >> "$new_links_file"
  fi
done < "$link_message_info_file"

# At this point, all of the lines should have one of these formats:
#     <markdown file>:<line number>;<updated link line>
#     <markdown file>:<line number>;ERROR <error message>

link_count="$( get_line_count "$new_links_file" )"
ok_count="$( grep -E ';\+\+\+ ' "$new_links_file" | get_line_count )"
problem_count="$( grep -F ';ERROR' "$new_links_file" | get_line_count )"
if [[ -n "$verbose" ]]; then
  printf 'Found %d links in %d files.\n' "$link_count" "${#files[@]}"
  printf 'Created %d updated links.\n' "$ok_count"
  printf 'Found %d problems.\n' "$problem_count"
fi

# Finally, if there are any errors, stop and output them.
problem_count="$( grep ';ERROR' "$new_links_file" | get_line_count )"
[[ -n "$verbose" ]] && printf 'Checking for link identification errors: %s\n' "$new_links_file"
if [[ "$problem_count" -ne '0' ]]; then
  printf 'Found %d problematic links in the markdown files.\n' "$problem_count"
  grep -F ';ERROR' "$new_links_file"
  clean_exit 1
fi

# Also make sure that every line in new_links_file is an updated link line.
if [[ "$link_count" -ne "$ok_count" ]]; then
  printf 'Could not create new links (%d) for every current link (%d).\n' "$ok_count" "$link_count"
  grep -Ev ';\+\+\+ ' "$new_links_file"
  clean_exit 1
fi

# Finally, all lines should now have this format:
#     <markdown file>:<line number>;<updated link line>

########################################################################################################################
#############################################  Update the markdown files.  #############################################
########################################################################################################################

[[ -z "$quiet" ]] && printf 'Updating %d links in %d files.\n' "$link_count" "${#files[@]}"

tfile="$temp_dir/temp.md"
problems=0
ec=0
i=0
while IFS="" read -r line || [[ -n "$line" ]]; do
  i=$(( i + 1 ))
  [[ -n "$verbose" ]] && printf '[%d/%d]: Applying update: %s\n' "$i" "$link_count" "$line"

  # Each line should have this format:
  #   <markdown file>:<line number>;<updated link line>
  md_file="$( sed -E 's/^([^:]+):.+$/\1/' <<< "$line" )"
  [[ -n "$verbose" ]] && printf '[%d/%d]: md_file: %s\n' "$i" "$link_count" "$md_file"
  line_number="$( sed -E 's/^[^:]+:([[:digit:]]+);.+$/\1/' <<< "$line" )"
  [[ -n "$verbose" ]] && printf '[%d/%d]: line_number: %s\n' "$i" "$link_count" "$line_number"
  new_link="$( sed -E 's/^[^:]+:[[:digit:]]+;//' <<< "$line" )"
  [[ -n "$verbose" ]] && printf '[%d/%d]: new_link: %s\n' "$i" "$link_count" "$new_link"

  if [[ -z "$md_file" || -z "$line_number" || -z "$new_link" || "$md_file" == "$line" || "$line_number" == "$line" || "$new_link" == "$line" || "${md_file}:${line_number};${new_link}" != "$line" ]]; then
    printf '[%d/%d]: ERROR: Could not parse new link line: %s\n' "$i" "$link_count" "$line"
    problems=$(( problems + 1 ))
    continue
  fi

  # This sed command uses a line number to identify which line to update, and the c directive to
  # replace the entire line with new content. It's a little clunky because it insists that the c
  # is followed by a \ and then the new content must all be on its own line. And it has to end in a
  # newline because sed doesn't automatically include that in the replacement; i.e., without it,
  # the line below gets appended to the new content and the file ultimately has one less line.
  # The ending newline can't be included in the printf because "$( )" might strip it.
  if ! sed "$( printf '%d c\\\n%s' "$line_number" "$new_link" )"$'\n' "$md_file" > "$tfile"; then
    printf '[%d/%d]: ERROR: Command failed: sed to update line %d in %s with new link: %s\n' "$i" "$link_count" "$line_number" "$md_file" "$new_link"
    problems=$(( problems + 1 ))
    continue
  fi

  if ! mv "$tfile" "$md_file"; then
    printf '[%d/%d]: ERROR: Command failed: mv "%s" "%s"\n' "$i" "$link_count" "$tfile" "$md_file"
    problems=$(( problems + 1 ))
    continue
  fi

  [[ -n "$verbose" ]] && printf '[%d/%d]: Success.\n' "$i" "$link_count"
done < "$new_links_file"

if [[ "$problems" -ne '0' ]]; then
  printf 'Failed to update %d links.\n' "$problems"
  clean_exit 1
fi

[[ -z "$quiet" ]] && printf 'Done.\n'
clean_exit 0
