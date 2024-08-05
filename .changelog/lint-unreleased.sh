#!/bin/bash
# This script will check that the dirs in unreleased are all valid.

where_i_am="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"
ur_dir="${where_i_am}/unreleased"
if [[ ! -d "$ur_dir" ]]; then
    printf 'Unreleased changes Directory does not exist: %s\n' "$ur_dir"
    exit 1
fi

valid_sections=( $( "${where_i_am}/get-valid-sections.sh" ) )
# Usage: is_valid_section <section>
# Returns with exit code 0 if it is valid, or 1 if not.
is_valid_section () {
    local s
    for s in "${valid_sections[@]}"; do
        if [[ "$s" == "$1" ]]; then
            return 0
        fi
    done
    return 1
}

ec=0
bad_sections=()
for section in $( find "$ur_dir" -type d -depth 1 | sed 's|^.*/||' ); do
    if ! is_valid_section "$section"; then
        bad_sections+=( "$section" )
    fi
done

if [[ "${#bad_sections[@]}" -ne '0' ]]; then
    printf 'Invalid unreleased section(s):\n'
    printf '.changelog/unreleased/%s\n' "${bad_sections[@]}"
    printf 'Valid sections: [%s]\n' "${valid_sections[*]}"
    ec=1
fi

bad_files="$( find "$ur_dir" -type f -depth 2 -name '*[[:space:]]*' | sed -E 's|^.*/\.changelog/|.changelog/|'  )"
if [[ -n "$bad_files" ]]; then
    printf 'Invalid unreleased filename(s):\n%s\n' "$bad_files"
    ec=1
fi

exit "$ec"
