#!/bin/bash
# This script will create the changelog entry for a dependabot PR.
# It's designed to be called by a github action kicked off because of a dependabot PR.

show_usage () {
    cat << EOF
dependabot-changelog.sh will create the changelog entries for a dependabot PR.

Usage: ./dependabot-changelog.sh --pr <num> --title <title> --head-branch <branch> --target-branch <branch>

--pr <num>
    Identifies the PR number for use in the links as well as the changelog filename.
--title <title>
    Identifies the title of the PR. It is used when there are no changes to go.mod.
    Expected format: "Bump <library> to <new version> from <old version>"
--head-branch <branch>
    Identifies the name of the branch with the change that we want merged into the target branch.
    For dependabot changes, it will have the format "dependabot/<type>/<library>-<new version>".
    The filename containing the new entries is derived from this.
--target-branch <branch>
    Identifies the branch that this PR is going into. It will almost always be "main".

EOF

}

while [[ "$#" -gt '0' ]]; do
    case "$1" in
        --help)
            printf 'Usage: ./dependabot-changelog.sh --pr <num> --title <title> --head-branch <branch> --target-branch <branch>\n'
            exit 0
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
        -t|--title)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            title="$2"
            shift
            ;;
        --head-branch)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            head_branch="$2"
            shift
            ;;
        --target-branch)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            target_branch="$2"
            shift
            ;;
        -v|--verbose)
            verbose="$1"
            ;;
        *)
            printf 'Unknown argument: [%s].\n' "$1"
            exit 1
            ;;
    esac
    shift
done

if [[ -z "$head_branch" ]]; then
    head_branch="$( git branch --show-current )"
fi

if [[ -z "$head_branch" || "$head_branch" == 'HEAD' ]]; then
    printf 'Could not determine the head branch and no --head-branch <branch> provided.\n'
    exit 1
fi
[[ -n "$verbose" ]] && printf '    Head Branch: "%s"\n' "$head_branch"

if [[ -z "$target_branch" ]]; then
    printf 'No --target-branch <branch> provided.\n'
    exit 1
fi
[[ -n "$verbose" ]] && printf '  Target Branch: "%s"\n' "$head_branch"


if [[ -z "$pr" ]]; then
    printf 'No --pr <num> provided.\n'
    exit 1
fi
[[ -n "$verbose" ]] && printf '             PR: "%s"\n' "$pr"

if [[ -z "$title" ]]; then
    printf 'No --title <title> provided.\n'
    exit 1
fi
[[ -n "$verbose" ]] && printf '          Title: "%s"\n' "$title"

# Dependabot branch names look like this: "dependabot/github_actions/bufbuild/buf-setup-action-1.34.0"
# The "github_actions" part can also be "go_modules" (and probably other things too).
# For the filename, we'll omit the "dependabot/<lib type>/" part and use just what's left.
branch_fn="$( sed -E 's|^[^/]+/[^/]+/||; s|/|-|g;' <<< "$head_branch" )"
[[ -n "$verbose" ]] && printf 'Branch Filename: "%s"\n' "$branch_fn"

# This script requires another script that must be in the same directory.
# To consistently find them, we'll need to know the absolute path to the dir with this script.
where_i_am="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"
[[ -n "$verbose" ]] && printf '     Where I Am: "%s"\n' "$where_i_am"

[[ -n "$verbose" ]] && printf 'Looking for go.mod dependency changes.\n'
# Run the script to create the entry from the changes in go.mod.
# The $verbose variable is purposely not quoted so that it doesn't count as an arg if it's empty.
"$where_i_am/get-dep-changes.sh" --pr "$pr" --name "$branch_fn" $verbose --force --target-branch "$target_branch"
ec=$?
[[ -n "$verbose" ]] && printf 'Exit code from get-dep-changes.sh: %d\n' "$ec"

# That script exits with 0 when there are go.mod changes and the new file was created.
# If there were go.mod changes, we're all done here.
# I don't think I've ever seen a dependabot PR that bumps both a go module and something else.
if [[ "$ec" -eq '0' ]]; then
    [[ -n "$verbose" ]] && printf 'Changes identified through go.mod. Done.\n'
    exit 0
fi

# That script exits with 10 to indicate there were no go.mod changes.
# All other exit codes are an error that requires attention.
if [[ "$ec" -ne '10' ]]; then
    printf 'An error was encountered.\n'
    exit "$ec"
fi

[[ -n "$verbose" ]] && printf 'Creating changelog entry from PR title.\n'

# Okay. There weren't any go.mod changes. It's a bump to something else (e.g. a
# github action helper). Create an entry ourselves, based on the title, which
# should look something like this: "Bump <library> from <old version> to <new version>".
# First, though, standardize the spacing so the rest of the regex stuff is cleaner.
title="$( sed -E 's/[[:space:]]+/ /; s/^ //; s/ $//;' <<< "$title" )"
[[ -n "$verbose" ]] && printf 'Clean Title: "%s"\n' "$title"
if ! grep -Eqi '^Bump [^ ]+ from [^ ]+ to [^ ]+$' <<< "$title"; then
    printf 'Unknown title format: %s\n' "$title"
    exit 1
fi

lib="$( sed -E 's/^Bump //; s/ from.*$//;' <<< "$title" )"
[[ -n "$verbose" ]] && printf '    Library: "%s"\n' "$lib"
if [[ -z "$lib" || "$lib" == "$title" || "$lib" =~ ' ' ]]; then
    printf 'Could not extract library from title: %s\n' "$title"
    exit 1
fi

old_ver="$( sed -E 's/^.*from //; s/ to.*$//' <<< "$title" )"
[[ -n "$verbose" ]] && printf '    Old Ver: "%s"\n' "$old_ver"
if [[ -z "$old_ver" || "$old_ver" == "$title" || "$old_ver" =~ ' ' ]]; then
    printf 'Could not extract old version from title: %s\n' "$title"
    exit 1
fi

new_ver="$( sed -E 's/^.*to //' <<< "$title" )"
[[ -n "$verbose" ]] && printf '    New Ver: "%s"\n' "$new_ver"
if [[ -z "$new_ver" || "$new_ver" == "$title" || "$new_ver" =~ ' ' ]]; then
    printf 'Could not extract new version from title: %s\n' "$title"
    exit 1
fi

link="[PR ${pr}](https://github.com/provenance-io/provenance/pull/${pr})"
[[ -n "$verbose" ]] && printf '       Link: "%s"\n' "$link"

repo_root="$( git rev-parse --show-toplevel 2> /dev/null )"
if [[ -z "$repo_root" ]]; then
    if [[ "$where_i_am" =~ /scripts$ ]]; then
        # If this is in the scripts directory, assume it's {repo_root}/scripts.
        repo_root="$( dirname "$where_i_am" )"
    else
        # Not in a git repo, and who knows where this script is in relation to the root,
        # so let's just hope that our current location is the repo root.
        repo_root='.'
    fi
    # Since we're not exactly sure we have the right repo_root here, we want to make sure the .changelog
    # dir already exists. If not, we'd probably be trying to create the new file in the wrong place.
    if [[ ! -d "${repo_root}/.changelog" ]]; then
        printf 'Could not identify target directory.\n'
        exit 1
    fi
fi
[[ -n "$verbose" ]] && printf '  Repo Root: "%s"\n' "$repo_root"

out_dir="${repo_root}/.changelog/unreleased/dependencies"
[[ -n "$verbose" ]] && printf ' Output Dir: "%s"\n' "$out_dir"
if ! mkdir -p "$out_dir"; then
    printf 'Could not create directory: %s\n' "$out_dir"
    exit 1
fi

name="$( sed -E 's/[^[:alnum:]]+/-/g; s/^-//; s/-$//;' <<< "$branch_fn" | tr '[:upper:]' '[:lower:]' )"
[[ -n "$verbose" ]] && printf '       Name: "%s"\n' "$name"
out_fn="${out_dir}/${pr}-${name}.md"
[[ -n "$verbose" ]] && printf 'Output File: "%s"\n' "$out_fn"

printf '* `%s` bumped to %s (from %s) %s.\n' "$lib" "$new_ver" "$old_ver" "$link" > "$out_fn" || exit 1

printf 'Dependabot changelog entry created: %s\n' "$out_fn"
exit 0
