#!/bin/bash
# This script will git diff go.mod and identify the changes to it,
# outputting the info in a format ready for the changelog.

temp_dir="$( mktemp -d -t link-updates )"
[[ -n "$verbose" ]] && printf 'Created temp dir: %s\n' "$temp_dir"
full_diff="${temp_dir}/1-full.diff"               # The full results of the diff.
minus_lines="${temp_dir}/2-minus-lines.txt"       # Just the subtractions we care about.
plus_lines="${temp_dir}/2-plus-lines.txt"         # Just the additions we care about.
minus_requires="${temp_dir}/3-minus-requires.txt" # Just the removed requirement lines.
plus_requires="${temp_dir}/3-plus-requires.txt"   # Just the added requirement lines.
minus_replaces="${temp_dir}/3-minus-replaces.txt" # Just the removed replace lines.
plus_replaces="${temp_dir}/3-plus-replaces.txt"   # Just the added replace lines.

# Usage: <stuff> | clean_diff
# This will reformat the lines from a diff and remove lines we don't care about.
# All result lines will have one of these formats:
#   "<library> <version>"
#   "<library> => <location>"
#   "<library> => <other library> <other version>"
# Note that this will strip out the leading + or -, so if that's important, it's up to you to keep track of.
clean_diff () {
    # Use sed to:
    #   Remove the + or - and any space immediately after it.
    #   If it now starts with "require" or "replace", remove that and any spaces after it.
    #   Remove any line-ending comment.
    #   Remove all trailing whitespace.
    #   Make sure there are spaces around the "=>" in replace lines.
    #   Change all groups of 1 or more whitespace characters to a single space.
    # Then use grep to only keep lines with one of the desired formats.
    sed -E 's/^[-+][[:space:]]*//; s/^(require|replace)[[:space:]]*//; s|//.*$||; s/[[:space:]]+$//; s/=>/ => /; s/[[:space:]]+/ /g;' \
        | grep -E '^[^ ]+ (v[^ ]+|=> [^ ]+( v[^ ]+)?)$'
}

# Usage: get_repl_str <lib> <filename>
# This will look for a replace line in <filename> for the <lib>.
# If found, it'll print a string describing the replacing version.
# If not found, this won't print anything.
get_replace_str () {
    local lib fn repl
    lib="$1"
    fn="$2"
    # Look in the file for a replace line for the library.
    # If found, keep only the stuff after the =>.
    # Using a variable for the grep expression is a bit clunky, so I'm using fixed string here.
    # Technically this will match both "<lib> =>" and "<pre><lib> =>". Matching the second one
    # would be problematic, but it's not accounted for in here. The nature of the names of the
    # libraries makes it highly unlikely to happen.
    repl="$( grep -F "$lib =>" "$fn" | sed -E 's/^.* => //' )"
    if [[ -n "$repl" ]]; then
        if [[ "$repl" =~ ' ' ]]; then
            # $repl has the format "<other library> <other version>"
            # It's provided without quotes so that it gets split on that space and provided
            # to this printf as two separate args to put tics around the <other library>.
            printf '`%s` %s' $repl
        else
            # $repl is a <location>, put tics around it.
            printf '`%s`' "$repl"
        fi
    fi
    return 0
}

# Get the go.mod diff.
git diff -U0 main -- go.mod > "$full_diff"
# Split it into subtractions and additions.
grep -E '^-' "$full_diff" | clean_diff > "$minus_lines"
grep -E '^\+' "$full_diff" | clean_diff > "$plus_lines"
# Split it further into require lines and replace lines.
grep -Ev '=>' "$minus_lines" > "$minus_requires"
grep -E '=>' "$minus_lines" > "$minus_replaces"
grep -Ev '=>' "$plus_lines" > "$plus_requires"
grep -E '=>' "$plus_lines" > "$plus_replaces"
# Identify all libraries that are changing.
libs=( $( sed -E 's/ .*$/' "$plus_lines" "$minus_lines" | sort -u ) )
