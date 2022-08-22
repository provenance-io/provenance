#!/usr/bin/env bash

#
# Check go.mod for any version updates of third-party proto releated libraries.
# If any such changes are found, get fresh copies of the protos and compare them to what's in the third_party directory.
# If any differences are found, issue a warning about needing to update those dependencies.
#
# The check of go.mod can be skipped by either providing the --force flag when executing this script,
# or by setting the FORCE env var to something (other than an empty string).
#

print_usage () {
    cat << EOF
Usage: scripts/proto-update-check.sh [--force] [branch]

[--force] - Download and compare the protos without checking for version changes.
            A non-empty FORCE environment variable is the same as providing the --force flag.
[branch]  - Optional branch to compare against for version changes. Default is '..origin/main'.
            This can also be provided in the BRANCH environment variable.
            This is ignored if forcing a check.

If this exits with code 0, all is good.

EOF

    return 0
}

# single brackets because the github runners don't have the enhanced (double-bracket) version.
while [ "$#" -gt '0' ]; do
    case "$1" in
        --help)
            print_usage
            exit 2
            ;;
        --force)
            FORCE='1'
            ;;
        *)
            if [ -n "$branch" ]; then
                printf 'Unknown argument: [%s]\n' "$1" >&2
                exit 2
            fi
            branch="$1"
            ;;
    esac
done

set -ex

branch="${branch:-${BRANCH:-..origin/main}}"
repo_root="$( cd "$( dirname "${BASH_SOURCE:-$0}" )/.."; pwd -P )"
update_deps="$repo_root/scripts/proto-update-deps.sh"

if [ -z "$FORCE" ]; then
    # The update-deps script uses go list -m to get library rewrites and version info on several libraries.
    # Search that script to get all of the libraries it checks up on.
    # Start by ignoring all commments, then find all go list -m {lib} entries and extract just the library.
    libs="$( sed 's:#.*$::' "$update_deps" | grep -Eo 'go list -m [^ ]+' | sed 's:.* ::' )"
    # Do our own go list -m to get both the library and possibly what it's rewritten to.
    # We want to check go.mod for changes in any of those.
    # We'll build a regexp to provide to grep to match any libraries of interest.
    libs_regexp=''
    for lib in $libs; do
        # go list -m "$lib" will look like one of these:
        # * '{lib} {version}'
        # * '{lib} {version} => {rw lib} {rw version}'
        # Sed replacements:
        #   s: [^ ]* => :|:  -- Changes [space]{version}[space]=>[space] into a pipe.
        #   s: [^ ]*$::      -- Gets rid of the version at the end of the line (either {version} or {rw version}).
        #   s:\.:\\.:        -- Adds a \ before all periods (to escape them for grep).
        # Result:
        # * '{lib} {version}' becomes '{lib}' (with periods escaped).
        # * '{lib} {version} => {rw lib} {rw version}' becomes '{lib}|{rw lib}' (with periods escaped).
        lib_regexp="$( go list -m "$lib" | sed 's: [^ ]* => :|:;  s: [^ ]*$::;  s:\.:\\.:g;' )"
        if [ -n "$libs_regexp" ]; then
            libs_regexp="$libs_regexp|"
        fi
        libs_regexp="$libs_regexp$lib_regexp"
    done
    # Diff go.mod with main (using ..origin/main here to accomodate github actions) with 0 lines of context.
    # Get rid of the lines that start with an @ since those are more context info and might contain false positives.
    # Remove any comments
    # And look for the libraries from earlier.
    # If grep finds anything it exits with a 0 so the || exit 0 isn't run.
    # If grep does not find anything, it'll exit 1 (or more) and the || exit 0 will run; there's no changes of interest, so nothing to do.
    git diff -U0 "$branch" -- go.mod \
        | grep -v '^@' \
        | sed 's://.*$::' \
        | grep -E "$libs_regexp" \
            || exit 0
fi

# If we get here, either we're forcing the check or there were changes of interest.

printf '\nDownloading latest third_party proto files for comparison...\n'

# Download third_party proto files in the build/ directory for comparison against /third_party
fresh_dir="$repo_root/build/third_party"
existing_dir="$repo_root/third_party"

rm -rf "$fresh_dir"
bash "$update_deps" "$fresh_dir"

printf '\nChecking Protobuf files for differences...\n'

# Diff will exit 0 if there are no changes, and something else if thera are.
# If there aren't changes, we can be done, all is good.
# The || true at the end lets us keep going when there are differecnes (since we're running under set -e).
diff -r --exclude '*.yaml' --exclude=google "$fresh_dir" "$existing_dir" && exit 0 || true

# There were differences, output a message and exit 1.
cat << EOF

Differences were identified in protobuf files between what's in /third_party, and freshly downloaded versions.

Often this can be fixed by running:
    make proto-update-deps
and committing anything updated.

This check can be run locally using:
    make proto-update-check

EOF

exit 1
