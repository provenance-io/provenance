#!/bin/bash
# This script will git diff go.mod and identify the changes to it,
# outputting the info in a format ready for the changelog.

show_usage () {
    cat << EOF
get-dep-changes.sh: Analyze changes made to go.mod and generate changelog entries.

Usage: get-dep-changes.sh {-p|--pull-request|--pr <num> | -n|--issue-no|--issue <num>}
          [--branch <branch>] [--name <name> [--dir <dir>]]
          [-v|--verbose] [--no-clean] [--force] [-h|--help]

You must provide either a PR number or issue number, but you cannot provide both.

-p|--pull-request|--pr <num>
    Append a PR link to the given <num> to the end of each changelog entry.
-n|--issue-no|--issue <num>
    Append an issue link to the given <num> to the end of each changelog entry.

--name <name>
    The <name> is cleaned then appended to the <num> to create the filename for this change.
    To clean the <name>, it is lowercased, then non-alphanumeric chars are changed to dashes.
    If provided, the changelog entries will be written to
        <repo root>.changelog/unreleased/dependencies/<num>-<name>.md
    If not provided, the changelog entries will be written to stdout.
    If not in a repo, or to put the file in a different directory, use the --dir <dir> option.

--dir <dir>
    Put the changelog entries in the provided <dir>.
    This arg only has meaning if --name is also provided.
    The default is '<repo root>.changelog/unreleased/dependencies'.

--branch <branch>
    Providing this option allows you to compare current changes against a branch other than main.
    By default, <branch> is "main".

-v|--verbose
    Output extra information.

--no-clean
    Do not delete the temporary directory used for processing.

--force
    If the output file already exists, overwrite it instead of outputting an error.

Exit codes:
    0  No errors encountered.
    1  An error was encountered.
    10 There are no changes to go.mod.

EOF

}

while [[ "$#" -gt '0' ]]; do
    case "$1" in
        -h|--help)
            show_usage
            exit 0
            ;;
        --no-clean)
            no_clean='YES'
            ;;
        -v|--verbose)
            verbose='YES'
            ;;
        -b|--branch)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            branch="$2"
            shift
            ;;
        -p|--pull-request|--pr)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            if [[ "$2" =~ [^[:digit:]] ]]; then
                printf 'Invalid %s value: [%s]. Only digits are allowed.\n' "$1" "$2"
                exit 1
            fi
            pr="$2"
            shift
            ;;
        -n|--issue-no|--issue)
            # Using -n and --issue-no to match the `unclog add` flags.
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            if [[ "$2" =~ [^[:digit:]] ]]; then
                printf 'Invalid %s value: [%s]. Only digits are allowed.\n' "$1" "$2"
                exit 1
            fi
            issue="$2"
            shift
            ;;
        --name)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            name="$2"
            shift
            ;;
        --dir)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            out_dir="$2"
            shift
            ;;
        --force)
            force='YES'
            ;;
        *)
            printf 'Unknown argument: %s\n' "$1"
            exit 1
            ;;
    esac
    shift
done

branch="${branch:-main}"

if [[ -n "$pr" && -n "$issue" ]]; then
    printf 'You cannot provide both a pr (%s) and issue (%s) number.\n' "$issue" "$pr"
    exit 1
elif [[ -n "$pr" ]]; then
    link="[PR ${pr}](https://github.com/provenance-io/provenance/pull/${pr})"
    num="$pr"
elif [[ -n "$issue" ]]; then
    link="[#${issue}](https://github.com/provenance-io/provenance/issues/${issue})"
    num="$issue"
else
    printf 'You must provide either a --pr <num> or --issue <num>.\n'
    exit 1
fi
[[ -n "$verbose" ]] && printf 'Link: %s\n' "$link"

if [[ -n "$name" ]]; then
    name="$( sed -E 's/[^[:alnum:]]+/-/g; s/^-//; s/-$//;' <<< "$name" | tr '[:upper:]' '[:lower:]' )"
    [[ -n "$verbose" ]] && printf 'Cleaned name: %s\n' "'$name'"
    if [[ -n "$name" ]]; then
        if [[ -z "$out_dir" ]]; then
            repo_root="$( git rev-parse --show-toplevel )" || exit 1
            out_dir="${repo_root}/.changelog/unreleased/dependencies"
        fi
        out_fn="${out_dir}/${num}-${name}.md"
        [[ -n "$verbose" ]] && printf 'Output filename: %s\n' "'$out_fn'"
        if [[ -z "$force" && -e "$out_fn" ]]; then
            printf 'Output file already exists: %s\n' "$out_fn"
            exit 1
        fi
    fi
fi

[[ -n "$verbose" ]] && printf 'Creating temp dir.\n'
temp_dir="$( mktemp -d -t dep-updates )" || exit 1
[[ -n "$verbose" ]] && printf 'Created temp dir: %s\n' "$temp_dir"

clean_exit () {
    local ec
    ec="${1:-0}"
    if [[ -n "$temp_dir" && -d "$temp_dir" ]]; then
        if [[ -z "$no_clean" ]]; then
            [[ -n "$verbose" ]] && printf 'Deleting temp dir: %s\n' "$temp_dir"
            rm -rf "$temp_dir"
            temp_dir=''
        else
            printf 'NOT deleting temp dir: %s\n' "$temp_dir"
        fi
    fi
    exit "$ec"
}

# Usage: <stuff> | clean_diff
# This will reformat the lines from a diff and remove lines we don't care about.
# All result lines will have one of these formats:
#   "<library> <version>"
#   "<library> => <location>"
#   "<library> => <other library> <other version>"
# Note that this will strip out the leading + or -, so if that's important, it's up to you to keep track of.
clean_diff () {
    # Use sed to:
    #   Remove the + or - and leading whitespace.
    #   If it now starts with "require" or "replace", remove that and any whitespace after it.
    #   Remove any line-ending comment.
    #   Remove all trailing whitespace.
    #   Make sure there are spaces around the "=>" in replace lines.
    #   Change all groups of 1 or more whitespace characters to a single space.
    # Then use grep to only keep lines with one of the desired formats.
    sed -E 's/^[-+[:space:]][[:space:]]*//; s/^(require|replace)[[:space:]]*//; s|//.*$||; s/[[:space:]]+$//; s/=>/ => /; s/[[:space:]]+/ /g;' \
        | grep -E '^[^ ]+ (v[^ ]+|=> [^ ]+( v[^ ]+)?)$'
    return 0
}

# Usage: get_replace_str <lib> <filename>
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

# Define all the temp files that we'll be making.
# The numbers in these roughly reflect the step that they're created in.
full_diff="${temp_dir}/1-full.diff"               # The full results of the diff.
minus_lines="${temp_dir}/2-minus-lines.txt"       # Just the subtractions we care about.
plus_lines="${temp_dir}/2-plus-lines.txt"         # Just the additions we care about.
minus_requires="${temp_dir}/3-minus-requires.txt" # Just the removed requirement lines.
plus_requires="${temp_dir}/3-plus-requires.txt"   # Just the added requirement lines.
cur_requires="${temp_dir}/3-cur-requires.txt"     # All current requires (even ones not changing).
minus_replaces="${temp_dir}/3-minus-replaces.txt" # Just the removed replace lines.
plus_replaces="${temp_dir}/3-plus-replaces.txt"   # Just the added replace lines.
cur_replaces="${temp_dir}/3-cur-replaces.txt"     # All current replaces (even ones not changing).
changes="${temp_dir}/4-changes.md"                # All the changelog entries (but without the link).
final="${temp_dir}/5-final.md"                    # The final changelog entry content.

# Get the go.mod diff.
[[ -n "$verbose" ]] && printf 'Creating full diff: %s\n' "$full_diff"
git diff -U0 "$branch" -- go.mod > "$full_diff"
if ! grep -q '.' "$full_diff"; then
    [[ -n "$verbose" ]] && printf 'go.mod does not have any changes.\n'
    # Using the exit code of 10 here to indicate no changes.
    clean_exit 10
fi

# Split it into subtractions and additions.
[[ -n "$verbose" ]] && printf 'Identifying all subtractions: %s\n' "$minus_lines"
grep -E '^-' "$full_diff" | clean_diff > "$minus_lines"
[[ -n "$verbose" ]] && printf 'Identifying all additions: %s\n' "$plus_lines"
grep -E '^\+' "$full_diff" | clean_diff > "$plus_lines"

# Split it further into require lines and replace lines.
[[ -n "$verbose" ]] && printf 'Identifying subtracted requires: %s\n' "$minus_requires"
grep -Ev '=>' "$minus_lines" > "$minus_requires"
[[ -n "$verbose" ]] && printf 'Identifying subtracted replaces: %s\n' "$minus_replaces"
grep -E '=>' "$minus_lines" > "$minus_replaces"
[[ -n "$verbose" ]] && printf 'Identifying added requires: %s\n' "$plus_requires"
grep -Ev '=>' "$plus_lines" > "$plus_requires"
[[ -n "$verbose" ]] && printf 'Identifying added replaces: %s\n' "$plus_replaces"
grep -E '=>' "$plus_lines" > "$plus_replaces"

# Identify all libraries that are changing.
[[ -n "$verbose" ]] && printf 'Identifying all changed libraries.\n'
libs=( $( sed -E 's/ .*$//' "$plus_lines" "$minus_lines" | sort -u ) )

# Identify all the current replace lines.
# This awk script outputs all lines that are either inside a "replace (" block,
# or start with "replace" (but aren't the beginning of a replace block).
# The clean_diff can be re-used here too to standardize the formatting and get only what we need.
[[ -n "$verbose" ]] && printf 'Identifying all current requires: %s\n' "$cur_requires"
awk '{if (inSec=="1" && /^[[:space:]]*\)[[:space:]]*$/) {inSec="";}; if (inSec=="1" || /^[[:space:]]*require[[:space:]]*[^([:space:]]/) {print $0;}; if (/^[[:space:]]*require[[:space:]]*\(/) {inSec="1";};}' go.mod | clean_diff > "$cur_requires"
[[ -n "$verbose" ]] && printf 'Identifying all current replaces: %s\n' "$cur_replaces"
awk '{if (inSec=="1" && /^[[:space:]]*\)[[:space:]]*$/) {inSec="";}; if (inSec=="1" || /^[[:space:]]*replace[[:space:]]*[^([:space:]]/) {print $0;}; if (/^[[:space:]]*replace[[:space:]]*\(/) {inSec="1";};}' go.mod | clean_diff > "$cur_replaces"

# Figure out (and output) the changelog entry for each lib being changed.
[[ -n "$verbose" ]] && printf 'Identifying changelog entries for %d libraries: %s\n' "${#libs[@]}" "$changes"
i=0
for lib in "${libs[@]}"; do
    i=$(( i + 1 ))
    [[ -n "$verbose" ]] && printf '[%d/%d]: Processing "%s".\n' "$i" "${#libs[@]}" "$lib"

    # These will either be empty or have the format "`<other lib>` <other version" or "`<location>`".
    new_repl="$( get_replace_str "$lib" "$plus_replaces" )"
    was_repl="$( get_replace_str "$lib" "$minus_replaces" )"
    cur_repl="$( get_replace_str "$lib" "$cur_replaces" )"
    [[ -n "$verbose" ]] && printf '[%d/%d]:   %s="%s"  %s="%s"  %s="%s"\n' "$i" "${#libs[@]}" 'new_repl' "$new_repl" 'was_repl' "$was_repl" 'cur_repl' "$cur_repl"


    # These will be either empty or be "<version>".
    new_req="$( grep -F "$lib v" "$plus_requires" | sed -E 's/^.* //' )"
    was_req="$( grep -F "$lib v" "$minus_requires" | sed -E 's/^.* //' )"
    cur_req="$( grep -F "$lib v" "$cur_requires" | sed -E 's/^.* //' )"
    [[ -n "$verbose" ]] && printf '[%d/%d]:   %s="%s"  %s="%s"  %s="%s"\n' "$i" "${#libs[@]}" 'new_req' "$new_req" 'was_req' "$was_req" 'cur_req' "$cur_req"

    # If there weren't changes to require lines, but there is a require line
    # for the library, we want a warning added to the end of the changelog entry
    # since that change would be largely inconsequential.
    warning=''
    if [[ -z "$new_repl" && -z "$was_repl" && -n "$cur_repl" ]]; then
        warning=" but is still replaced by $cur_repl"
    fi

    # Pick the strings to use for the old and new versions.
    # If there was a change to a replace line, use that over a changed require line.
    new="${new_repl:-$new_req}"
    was="${was_repl:-$was_req}"

    # Edge case: A replace line was removed, and the main entry didn't change.
    if [[ -z "$new" && -n "$was_repl" && -z "$was_req" && -n "$cur_req" ]]; then
        # We want to report that we are now on the currently required version.
        new="$cur_req"
        [[ -n "$verbose" ]] && printf '[%d/%d]:   Using currently required version as new.\n' "$i" "${#libs[@]}"
    fi

    # Edge case: A replace line was added, and the main entry didn't change.
    if [[ -z "$was" && -n "$new_repl" && -z "$new_req" && -n "$cur_req" ]]; then
        # We want to report that we were on the the currently required version.
        was="$cur_req"
        [[ -n "$verbose" ]] && printf '[%d/%d]:   Using currently required version as was.\n' "$i" "${#libs[@]}"
    fi

    [[ -n "$verbose" ]] && printf '[%d/%d]:   %s="%s"  %s="%s"  %s="%s"\n' "$i" "${#libs[@]}" 'new' "$new" 'was' "$was" 'warning' "$warning"

    # Now generate the changelog line for this library.
    if [[ -n "$new" && -n "$was" ]]; then
        if [[ "$new" != "$was" ]]; then
            [[ -n "$verbose" ]] && printf '[%d/%d]: Creating bump line.\n' "$i" "${#libs[@]}"
            printf '* `%s` bumped to %s (from %s)%s\n' "$lib" "$new" "$was" "$warning" >> "$changes"
        else
            [[ -n "$verbose" ]] && printf '[%d/%d]: No change to report.\n' "$i" "${#libs[@]}"
        fi
    elif [[ -n "$new" ]]; then
        [[ -n "$verbose" ]] && printf '[%d/%d]: Creating add line.\n' "$i" "${#libs[@]}"
        printf '* `%s` added at %s%s\n' "$lib" "$new" "$warning" >> "$changes"
    elif [[ -n "$was" ]]; then
        [[ -n "$verbose" ]] && printf '[%d/%d]: Creating remove line.\n' "$i" "${#libs[@]}"
        printf '* `%s` removed at %s%s\n' "$lib" "$was" "$warning" >> "$changes"
    else
        # It shouldn't be possible to see this, but it's here just in case things go wonky.
        [[ -n "$verbose" ]] && printf '[%d/%d]: Creating unknown line.\n' "$i" "${#libs[@]}"
        printf '* `%s` TODO: Could not identify dependency change details, fix me.\n' "$lib" >> "$changes"
    fi
done

# Append the link to each line.
[[ -n "$verbose" ]] && printf 'Appending link (%s) to each entry: %s\n' "$link" "$final"
awk -v link="$link" '{print $0 " " link ".";}' "$changes" > "$final"

# Either put the file in place or output the content.
if [[ -n "$out_fn" ]]; then
    out_dir="$( dirname "$out_fn" )"
    if [[ -n "$out_dir" && "$out_dir" != '.' ]]; then
        [[ -n "$verbose" ]] && printf 'Making dir (if it does not exist yet): %s\n' "$out_dir"
        mkdir -p "$out_dir" || failed="YES"
    fi
    if [[ -z "$failed" ]]; then
        [[ -n "$verbose" ]] && printf 'Copying final file from %s to %s\n' "$final" "$out_fn"
        if cp "$final" "$out_fn"; then
            copied='YES'
            printf 'Dependency changelog entry created: %s\n' "$out_fn"
        fi
    fi
fi

if [[ -z "$copied" ]]; then
    cat "$final"
fi

clean_exit 0