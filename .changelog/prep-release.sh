#!/bin/bash
# This script will update the changelog stuff to mark a release.

show_usage () {
    cat << EOF
prep-release.sh: Prepares the changelog for a new release.

Usage: prep-release.sh <version> [--date <date> [--force-date]]
                [--force-version] [-v|--verbose] [--no-clean] [-h|--help]

The <version> must have format vA.B.C or vA.B.C-rcX where A, B, C and X are numbers.

--date <date> is an optional way to define the release date.
    Must have the format YYYY-MM-DD.
    The default is today's date.
--force-date indicates that the provided date is correct.
    By default, this script won't allow dates that are more than 14 days before or after today.
    This flag allows you to bypass that check.

--no-clean causes the temporary directory to remain once the script has exited.

EOF

}

while [[ "$#" -gt '0' ]]; do
    case "$1" in
        -h|--help)
            show_usage
            exit 0
            ;;
        --date)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            date="$2"
            shift
            ;;
        --force-date)
            force_date="$1"
            ;;
        --force-version)
            force_version="$1"
            ;;
        -v|--verbose)
            verbose="$1"
            ;;
        --no-clean)
            no_clean="$1"
            ;;
        *)
            if [[ -n "$version" ]]; then
                printf 'Unknown argument: %s\n' "$1"
            fi
            version="$1"
            ;;
    esac
    shift
done

if [[ -z "$version" ]]; then
    show_usage
    exit 0
fi

if ! command -v unclog > /dev/null 2>&1; then
    # Issue standard command-not-found message.
    unclog
    printf 'See: https://github.com/informalsystems/unclog\n'
    exit 1
fi

repo_root="$( git rev-parse --show-toplevel 2> /dev/null )"
if [[ -z "$repo_root" ]]; then
    where_i_am="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"
    if [[ "$where_i_am" =~ /scripts$ ]]; then
        # If this is in the scripts directory, assume it's {repo_root}/scripts.
        repo_root="$( dirname "$where_i_am" )"
    else
        # Not in a git repo, and who knows where this script is in relation to the root,
        # so let's just hope that our current location is the repo root.
        repo_root='.'
    fi
fi
[[ -n "$verbose" ]] && printf ' Repo root dir: [%s].\n' "$repo_root"

changelog_file="${repo_root}/CHANGELOG.md"
if [[ ! -f "$changelog_file" ]]; then
    printf 'Could not find existing CHANGELOG.md file.\n'
    exit 1
fi
[[ -n "$verbose" ]] && printf 'Changelog file: [%s].\n' "$changelog_file"


# Do some superficial validation on the provided version. We'll do more later though.
if ! grep -Eq '^v[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+(-rc[[:digit:]]+)?$' <<< "$version" 2> /dev/null; then
    printf 'Invalid version format [%s]. Use vA.B.C or vA.B.C-rcX.\n' "$version"
    exit 1
fi
v_major="$( sed -E 's/^v([^.]+)\..*$/\1/' <<< "$version" )"
v_minor="$( sed -E 's/^[^.]+\.//; s/\..*$//' <<< "$version" )"
v_patch="$( sed -E 's/^[^.]+\.[^.]+\.//; s/-rc.*$//;' <<< "$version" )"
v_rc="$( sed -E 's/^[^.]+\.[^.]+\.[^-]+//; s/^.*-rc//;' <<< "$version" )"
if [[ -n "$v_rc"  && "$v_rc" -eq '0' ]]; then
    printf 'Invalid version: [%s]. Release candidate numbering starts at 1.\n' "$version"
    exit 1
fi
v_base="v${v_major}.${v_minor}.${v_patch}"
[[ -n "$verbose" ]] && printf 'Version: [%s] = Major: [%s] . Minor: [%s] . Patch: [%s] (%s) - RC: [%s]\n' \
                                "$version" "$v_major" "$v_minor" "$v_patch" "$v_base" "$v_rc"

# Validate the date (or get it if not provided).
if [[ -z "$date" ]]; then
    date="$( date +'%F' )"
elif [[ ! "$date" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}$ ]]; then
    printf 'Invalid date format [%s]. Use YYYY-MM-DD.\n' "$date"
    exit 1
else
    # The GNU version of `date --help` exits with code 0.
    # The BSD/OSX version of `date --help` exits with code 1.
    # We use that to identify which version of the date command we have so we can use the correct args.
    if date --help > /dev/null 2>&1; then
        gnu_date='YES'
        # GNU version. If the provided date is not a valid date (e.g. month 13 or day 31 in a 30 day month),
        # this command will exit with code 1 and output some stuff to stderr (which we nullify).
        # If it's an actual date, it will exit with code 0.
        if ! date -d "$date" +'%F' > /dev/null 2>&1; then
            printf 'Invalid date: %s\n' "$date"
            exit 1
        fi
    else
        # BSD/OSX version. This one is so good that providing a date of '2024-06-31' is just treated as valid, but `2024-07-01`.
        # But something like '2024-13-01' or '2024-07-32' exit with code 1 and print only to stderr.
        # So for this one, we have to compare that result back to the date we have.
        if [[ "$date" != "$( date -j -f '%F' "$date" +'%F' 2> /dev/null )" ]]; then
            printf 'Invalid date: %s\n' "$date"
            exit 1
        fi
    fi

    # Make sure the date is within the previous or next 14 days.
    # This is mostly to make it harder to accidentally use the wrong year or month.
    if [[ -n "$gnu_date" ]]; then
        date_s="$( date -d "$date" +'%s' 2> /dev/null )"
        # GNU date will read the format 'YYYY-MM-DD' as having 00:00:00 for the time, which is what we want.
        cur_date_s="$( date -d "$( date +'%F' )" +'%s' )"
    else
        date_s="$( date -j -f '%F' "$date" +'%s' )"
        # BSD/OSX date will read the format 'YYYY-MM-DD' as having the current time, but we want the epoch at
        # the start of the day so that we're only paying attention to whole days.
        cur_date_s="$( date -j -f '%F %T' "$( date +'%F' ) 00:00:00" +'%s' )"
    fi
    date_diff_s=$(( date_s - cur_date_s ))
    date_diff_s=${date_diff_s#-} # remove the leading minus if there.
    # 60s/m * 60m/h * 24h/d * 14d = 1209600s
    if [[ "$date_diff_s" -gt '1209600' ]]; then
        if [[ -z "$force_date" ]]; then
            printf 'The date %s is too far away. If it is correct, rerun this command and include the --force-date flag.\n' "$date"
            exit 1
        else
            printf 'Warning: [%s] is more than 14 days away from today.\n' "$date"
        fi
    fi
fi
[[ -n "$verbose" ]] && printf 'Date: [%s].\n' "$date"

# Create a temp directory to store some processing files.
[[ -n "$verbose" ]] && printf 'Creating temp dir.\n'
temp_dir="$( mktemp -d -t prep-release.XXXX )" || exit 1
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

# Usage: handle_invalid_version <msg> <args>
handle_invalid_version () {
    local msg
    msg="$1"
    shift
    if [[ -z "$force_version" ]]; then
        printf 'Invalid version: [%s]. '"$msg"'\n' "$version" "$@"
        clean_exit 1
    else
        printf 'Warning: Version: [%s]: '"$msg"'\n' "$version" "$@"
    fi
    return 0
}

# Create a file with a list of all the current versions.
versions_file="${temp_dir}/versions.txt"
[[ -n "$verbose" ]] && printf 'Creating versions file: %s.\n' "$versions_file"
grep -oE '^## \[v[^]]+' "$changelog_file" | sed -E 's/^## \[//' > "$versions_file"

# Do some more validation on the version.
[[ -n "$verbose" ]] && printf 'Validating new version against existing ones.\n'

if grep -qFx "$version" "$versions_file" 2> /dev/null; then
    handle_invalid_version 'Version already exists.'
fi
if [[ -n "$v_rc" ]] && grep -qFx "$v_base" "$versions_file" 2> /dev/null; then
    handle_invalid_version 'Cannot create a release candidate for a version that already exists: [%s].' "$v_base"
fi

# Get the most recent non-rc version so that we can ensure we're using the right next version.
prev_ver="$( { cat "$versions_file"; printf '%s\n' "$v_base"; } | grep -v -e '-rc' | sort --version-sort --reverse | grep -A 1 -Fx "$v_base" | tail -n 1 )"

if [[ -n "$v_rc" && "$v_rc" -ge '2' ]]; then
    # If the new version is a release candidate of 2 or more, also ensure the previous rc exists.
    prev_ver_rc="$( { cat "$versions_file"; printf '%s\n' "$version"; } | sort --version-sort --reverse | grep -A 1 -Fx "$version" | tail -n 1 )"
    if [[ "${v_base}-rc$(( v_rc - 1 ))" != "$prev_ver_rc" ]]; then
        handle_invalid_version 'Release candidate is not sequential. Previous version: [%s].' "$prev_ver_rc"
    fi
fi

if [[ "$v_patch" -ne '0' ]]; then
    if [[ "v${v_major}.${v_minor}.$(( v_patch - 1 ))" != "$prev_ver" ]]; then
        handle_invalid_version 'Patch number is not sequential. Previous version: [%s].' "$prev_ver"
    fi
elif [[ "$v_minor" -ne '0' ]]; then
    if [[ "v${v_major}.$(( v_minor - 1 ))." != "$( sed -E 's/[^.]+$/' <<< "$prev_ver" )" ]]; then
        handle_invalid_version 'Minor number is not sequential. Previous version: [%s].' "$prev_ver"
    fi
else
    if [[ "v${v_major}." != "$( sed -E 's/[^.]+\.[^.]+*$//' <<< "$prev_ver" )" ]]; then
        handle_invalid_version 'Major number is not sequential. Previous version: [%s].' "$prev_ver"
    fi
fi

if [[ -n "$verbose" ]]; then
    printf '            New Version: [%s].\n' "$version"
    printf 'Previous Non-RC Version: [%s].\n' "$prev_ver"
    [[ -n "$prev_ver_rc" ]] && printf '    Previous RC Version: [%s].\n' "$prev_ver_rc"
fi


clean_exit 0
