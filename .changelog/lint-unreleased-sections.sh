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
for section in $( find "$ur_dir" -type d -depth 1 | sed 's|^.*/||' ); do
    if ! is_valid_section "$section"; then
        printf 'Invalid unreleased section directory: %s\n' "$section"
        ec=1
    fi
done

if [[ "$ec" -ne '0' ]]; then
    printf 'Valid sections: [%s]\n' "${valid_sections[*]}"
fi
exit "$ec"
