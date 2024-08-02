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

########################################################################################################################
################################################  Setup and Validation  ################################################
########################################################################################################################

printf 'Doing Setup and Validation.\n'

repo_root="$( git rev-parse --show-toplevel 2> /dev/null )"
if [[ -z "$repo_root" ]]; then
    where_i_am="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"
    if [[ "$where_i_am" =~ /.changelog$ || "$where_i_am" =~ /scripts$ ]]; then
        # If this is in the .changelog or scripts directory, assume it's {repo_root}/<dir>.
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
changelog_dir="${repo_root}/.changelog"
if [[ ! -d "$changelog_dir" ]]; then
    printf 'Could not find the .changelog/ dir.\n'
    exit 1
fi
[[ -n "$verbose" ]] && printf ' Changelog dir: [%s].\n' "$changelog_dir"


# Do some superficial validation on the provided version. We'll do more later though.
[[ -n "$verbose" ]] && printf 'Doing initial validation on the version: [%s].\n' "$version"
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

# Make sure the new version directory does not already exist.
new_ver_dir="${changelog_dir}/${version}"
if [[ -d "$new_ver_dir" ]]; then
    # If the new version directory already exists, and you are redoing the changelog for it, you'll need to
    # move the things you want into unreleased and then delete the existing version directory.
    # Nothing is done about it in here because there's no way to know what all should be included, or even
    # if the existing version was just provided by accident.
    printf 'The changelog version directory for [%s] already exists: [%s].\n' "$version" "$new_ver_dir"
    exit 1
fi

# If this is not an rc, there needs to be a summary file.
# If this is an rc and there isn't a summary file, we'll create a default one later (when it's needed).
unreleased_dir="${changelog_dir}/unreleased"
unreleased_sum_file="${unreleased_dir}/summary.md"
[[ -n "$verbose" ]] && printf 'Checking summary file: [%s].\n' "$unreleased_sum_file"
if [[ -f "$unreleased_sum_file" ]] && awk '{ sub(/<!--.*-->/,""); if (in_com) { if (sub(/^.*-->/,"")) { in_com=0; } else { $0=""; }; } else if (sub(/<!--.*$/,"")) { in_com=1; }; if (/./) { all_good=1; exit 0; }; } END { if (!all_good) { exit 1; }; }' "$unreleased_sum_file"; then
    have_summary='YES'
fi
if [[ -z "$v_rc" ]]; then
    if [[ -z "$have_summary" ]]; then
        printf 'A summary is required, but the file either does not exist or does not have any content: [%s].\n' "$unreleased_sum_file"
        exit 1
    fi
fi

# Validate the date (or get it if not provided).
[[ -n "$verbose" ]] && printf 'Getting or validating the date: [%s].\n' "$date"
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

# Usage: clean_exit [<code>]
# Default <code> is 0.
# Since we now have a temp dir to clean up, there should not be any exit statements after this. Use this instead.
# Cleans up the temp dir and exits.
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
# Outputs a message about an invalid version. Then, if not forcing the version, it'll exit this script.
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

# Do some more validation on the new version.
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

# Usage: combine_rc_dirs
# This is extracted as a function for easier short-circuit control in this process.
# It will move all the entry files from the rc dirs for this version into unreleased.
combine_rc_dirs () {
    local rc_vers v rc_ver v_id rc_ver_dir entries sections s section s_id s_dir e entry e_id v_file u_file

    [[ -n "$verbose" ]] && printf 'Identifying rc version dirs for this version.\n'
    rc_vers=( $( find "$changelog_dir" -type d -depth 1 -name "${version}-rc*" | sed -E 's|^.*/||' ) )
    [[ -n "$verbose" ]] && printf 'Found %d version dirs: [%s].\n' "${#rc_vers[@]}" "${rc_vers[*]}"
    [[ "${#rc_vers[@]}" -eq '0' ]] && return 0

    printf 'Combining Release Candidate dirs back into unreleased.\n'
    v=0
    for rc_ver in "${rc_vers[@]}"; do
        v=$(( v + 1 ))
        v_id="[${v}/${#rc_vers[@]}=${rc_ver}]"
        [[ -n "$verbose" ]] && printf '%s: Identifying entry files.\n' "$v_id"
        rc_ver_dir="${changelog_dir}/${rc_ver}"
        entries=( $( find "$rc_ver_dir" -type f -depth 2 -name '*.md' | grep -Eo '[^/]+/[^/]+$' ) )
        [[ -n "$verbose" ]] && printf '%s: Found %d entry files.\n' "$v_id" "${#entries[@]}"

        if [[ "${#entries[@]}" -gt '0' ]]; then
            [[ -n "$verbose" ]] && printf '%s: Identifying sections.\n' "$v_id"
            sections=( $( printf '%s\n' "${entries[@]}" | sed -E 's|/.*$||' | sort -u ) )

            [[ -n "$verbose" ]] && printf '%s: Making sure %d sections exist in unreleased: [%s].\n' "$v_id" "${#sections[@]}" "${sections[*]}"
            s=0
            for section in "${sections[@]}"; do
                s=$(( s + 1 ))
                s_id="${v_id}[${s}/${#sections[@]}]"
                [[ -n "$verbose" ]] && printf '%s: Making section [%s] in unreleased (if it does not exist yet).\n' "$s_id" "$section"
                s_dir="${unreleased_dir}/${section}"
                if [[ ! -d "$s_dir" ]] && ! mkdir "$s_dir"; then
                    printf '%s: Failed to make section dir: [%s].\n' "$s_id" "$section"
                    return 1
                fi
            done

            [[ -n "$verbose" ]] && printf '%s: Moving %d entry files to unreleased.\n' "$v_id" "${#entries[@]}"
            e=0
            for entry in "${entries[@]}"; do
                e=$(( e + 1 ))
                e_id="${v_id}[${e}/${#entries[@]}]"

                [[ -n "$verbose" ]] && printf '%s: Moving entry to unreleased: [%s].\n' "$e_id" "$entry"
                v_file="${rc_ver_dir}/${entry}"
                u_file="${unreleased_dir}/${entry}"
                if [[ -e "$u_file" ]]; then
                    if diff -q "$v_file" "$u_file" > /dev/null 2>&1; then
                        # The unreleased file already exists and is the same as the version one. Just delete the version one.
                        if ! rm "$v_file"; then
                            printf "%s: Failed to delete version file: [%s].\n" "$e_id" "$v_file"
                            return 1
                        fi
                    else
                        printf '%s: Cannot move entry [%s] into unreleased because it already exists and is different.\n' "$e_id" "$entry"
                        return 1
                    fi
                else
                    if ! mv "$v_file" "$u_file"; then
                        printf '%s: Failed to move entry file [%s] to [%s].\n' "$e_id" "$v_file" "$u_file"
                        return 1
                    fi
                fi
            done
        fi

        [[ -n "$verbose" ]] && printf '%s: Deleting rc dir.\n' "$v_id"
        if [[ -n "$( find "$rc_ver_dir" -type f -not -name '.*' -not -name 'summary.md' )" ]]; then
            printf '%s: Cannot delete non-empty rc directory.\n' "$v_id"
            return 1
        elif ! rm -rf "$rc_ver_dir"; then
            printf '%s: Failed to delete rc directory.\n' "$v_id"
            return 1
        fi
    done

    return 0
}

# If this is not an rc, we want to move all of the "released" rc entries into unreleased so that we can
# get the changelog of all the new stuff in this release.
if [[ -z "$v_rc" ]]; then
    combine_rc_dirs || clean_exit 1
fi

########################################################################################################################
##############################################  Build New Release Notes  ###############################################
########################################################################################################################

printf 'Creating the new Release Notes file.\n'

# If this is an rc and we don't have a summary, create a default one now.
if [[ -n "$v_rc" && -z "$have_summary" ]]; then
    printf 'This is Provenance Blockchain release candidate `%s`.\n' "$version" > "$unreleased_sum_file"
fi

# Use unclog to generate the beginnings of the new release notes.
unclog_build_file="${temp_dir}/1-unclog-build.md"
[[ -n "$verbose" ]] && printf 'Generating initial changelog entry: [%s].\n' "$unclog_build_file"
cd "${repo_root}" || clean_exit 1
unclog build --unreleased-only > "$unclog_build_file" || clean_exit 1

# Make sure all the PR and issue links have the correct link text.
# Also, if there's period right before and after the link, get rid of the one before.
links_fixed_file="${temp_dir}/2-links-fixed.md"
[[ -n "$verbose" ]] && printf 'Fixing the link text: [%s].\n' "$links_fixed_file"
sed -E -e 's|\[[^[:digit:]]*[[:digit:]]+\](\([^)]+/pull/([[:digit:]]+)\))|[PR \2]\1|g' \
       -e 's|\[[^[:digit:]]*[[:digit:]]+\](\([^)]+/issues/([[:digit:]]+)\))|[#\2]\1|g' \
       -e 's|\.([[:space:]]*\[[^]]+\]\([^)]+\))[[:space:]]*\.|\1.|g' \
       "$unclog_build_file" > "$links_fixed_file" || clean_exit 1
if grep -qE '[^[:space:]]' <<< "$( tail -n 1 "$links_fixed_file" )" > /dev/null 2>&1; then
    # The last line is not a blank line, add one now so that the last section is consistent with the rest.
    printf '\n' >> "$links_fixed_file"
fi

# Split it out into individual sections so that we can more easily re-order them.
cur_file="${temp_dir}/3-section-top.md"
[[ -n "$verbose" ]] && printf 'Splitting into section files.\nNow writing to: [%s].\n' "$cur_file"

# Usage: lower_to_title <words>
# Converts the first letter of each word to upper-case, and standardizes word spacing.
lower_to_title () {
    # Unfortunately, there isn't an easy way to do this that is also portable.
    # Expansions of ${var,,} (convert everything to lower) and ${var^} (convert first char to upper),
    # aren't available everywhere. Also, not all versions of sed allow for the \L and \u directives.
    # Even the toupper and tolower functions in awk aren't always available.
    # The ${var::} expansion seems to be pretty widely available, though, so we'll use that with tr on each word.
    local words
    words=()
    while [[ "$#" -gt '0' ]]; do
        # ${1:0:1} means "in the $1 variable, get a substring starting at char 0 with length 1."
        # ${1:1} means "in the $1 variable, get a substring starting at char 1 (and going to the end of the string)."
        words+=( "$( tr '[:lower:]' '[:upper:]' <<< "${1:0:1}" )${1:1}" )
        shift
    done
    # This also standardizes the spacing before, between, and after the words.
    printf '%s' "${words[*]}"
}

while IFS="" read -r line || [[ -n "$line" ]]; do
    if [[ "$line" =~ ^##[[:space:]]+Unreleased[[:space:]]*$ ]]; then
        printf '## [%s](https://github.com/provenance-io/provenance/releases/tag/%s) %s\n' "$version" "$version" "$date" >> "$cur_file"
    elif [[ "$line" =~ ^###[[:space:]] ]]; then
        [[ -n "$verbose" ]] && printf 'Found new section line: [%s].\n' "$line"
        section="$( sed -E 's/^###[[:space:]]+//; s/[[:space:]]+$//; s/[^[:alnum:]]+/-/g;' <<< "$line" | tr '[:upper:]' '[:lower:]' )"
        cur_file="${temp_dir}/3-section-${section}.md"
        [[ -n "$verbose" ]] && printf 'Now writing to: [%s].\n' "$cur_file"
        # I'm providing the section unquoted here so that the shell splits it into words for us.
        printf '### %s\n' "$( lower_to_title $( tr '-' ' ' <<< "$section" ) )" >> "$cur_file"
    else
        printf '%s\n' "$line" >> "$cur_file"
    fi
done < "$links_fixed_file"

# Sort the entries of the dependencies section.
# They have the format "* `<library>` <action> <version> ..." where <action> is one of "added at" "bumped to" or "removed at".
# So, if we just sort them using version sort, it'll end up sorting them by library and version, which a handy way to view them.
dep_file="${temp_dir}/3-section-dependencies.md"
if [[ -f "$dep_file" ]]; then
    [[ -n "$verbose" ]] && printf 'Sorting the dependency entries: [%s].\n' "$dep_file"
    orig_dep_file="${dep_file}.orig"
    mv "$dep_file" "$orig_dep_file"
    head -n 2 "$orig_dep_file" > "$dep_file"
    grep -E '^[[:space:]]*[-*]' "$orig_dep_file" | sort --version-sort >> "$dep_file"
    printf '\n' >> "$dep_file"
fi

new_rl_file="${temp_dir}/4-release-notes.md"
[[ -n "$verbose" ]] && printf 'Re-combining sections in the desired order: [%s].\n' "$new_rl_file"
# Usage: include_sections <section 1> <section 2> ...
# Appends the provided sections to the new_rl_file ("temp release notes file"),
# and marks each section as included.
include_sections () {
    local s section s_id s_file
    s=0
    for section in "$@"; do
        s=$(( s + 1 ))
        s_id="[${s}/${#}=${section}]"
        s_file="${temp_dir}/3-section-${section}.md"
        if [[ -f "$s_file" ]]; then
            [[ -n "$verbose" ]] && printf '%s: Including [%s].\n' "$s_id" "$s_file"
            if ! cat "$s_file" >> "$new_rl_file"; then
                printf '%s: Could not append [%s] to [%s].\n' "$s_id" "$s_file" "$new_rl_file"
                clean_exit 1
            fi
            if ! mv "$s_file" "$s_file.included"; then
                printf '%s: Could not mark file as included: [%s].\n' "$s_id" "$s_file"
                clean_exit 1
            fi
        else
            [[ -n "$verbose" ]] && printf '%s: No section file to include: [%s].\n' "$s_id" "$s_file"
        fi
    done
}

section_order=()
section_order+=( top )
section_order+=( $( awk '{ if (in_com) { if (/^"/) { sub(/^"/,""); sub(/".*$/,""); sub(/^[[:space:]]+/,""); sub(/[[:space:]]$/,""); gsub(/[[:space:]]+/,"-"); print $0; } else if (/-->/) { exit 0; } }; if (/<!--/) { in_com=1; }; }' "$changelog_file" | tr '[:upper:]' '[:lower:]' ) )
[[ -n "$verbose" ]] && printf 'Order (%d): [%s].\n' "${#section_order[@]}" "${section_order[*]}"
include_sections "${section_order[@]}"

other_sections=()
other_sections+=( $( find "$temp_dir" -type f -name '3-section-*.md' | sed -E 's|^.*/3-section-||; s/\.md$//;' | sort ) )
if [[ "${#other_sections[@]}" -ne '0' ]]; then
    [[ -n "$verbose" ]] && printf 'Including other sections (%d): [%s].\n' "${#other_sections[@]}" "${other_sections[*]}"
    include_sections "${other_sections[@]}"
fi

[[ -n "$verbose" ]] && printf 'Appending diff links: [%s].\n' "$new_rl_file"
printf '### Full Commit History\n\n' >> "$new_rl_file"
[[ -n "$prev_ver_rc" ]] && printf '* https://github.com/provenance-io/provenance/compare/%s...%s\n' "$prev_ver_rc" "$version" >> "$new_rl_file"
printf '* https://github.com/provenance-io/provenance/compare/%s...%s\n\n' "$prev_ver" "$version" >> "$new_rl_file"

# Usage: clean_versions <input file>
#    or: <stuff> clean_versions
clean_versions () {
    awk -v version="$version" '{ if (/^##[[:space:]]/) { if (index($0,version "-rc") || index($0,"[" version "]")) { keep=0; } else { keep=1; }; }; if (keep) { print $0; }; }' "$0"
}

# If this is an rc and there's an existing release notes, append those to the end, removing any existing section for this version.
release_notes_file="${repo_root}/RELEASE_NOTES.md"
if [[ -n "$v_rc" && -f "$release_notes_file" ]]; then
    [[ -n "$verbose" ]] && printf 'Including existing release notes: [%s].\n' "$release_notes_file"
    clean_versions "$release_notes_file" >> "$new_rl_file"
fi

[[ -n "$verbose" ]] && printf 'Putting release notes into place: cp "%s" "%s".\n' "$new_rl_file" "$release_notes_file"
cp "$new_rl_file" "$release_notes_file" || clean_exit 1

########################################################################################################################
###############################################  Building New Changelog  ###############################################
########################################################################################################################

printf 'Creating the new Changelog file.\n'

new_cl_file="${temp_dir}/5-new-changelog.md"
[[ -n "$verbose" ]] && printf 'Putting top of changelog in temp file: [%s].\n' "$new_cl_file"
# Get the top of the changelog file up to the first version header that isn't unreleased.
awk '{if (/^##[[:space:]]/ && $0 !~ /[Uu][Nn][Rr][Ee][Ll][Ee][Aa][Ss][Ee][Dd]/) { exit 0; }; print $0; }' "$changelog_file" > "$new_cl_file"

[[ -n "$verbose" ]] && printf 'Appending the new release notes (without the blurb): [%s].\n' "$new_cl_file"
# Include the new stuff, but remove any blurb.
awk '{ if (/^##[[:space:]]/) { in_top=1; print $0; print ""; } else if (/^###[[:space:]]/) { in_top=0; }; if (!in_top) { print $0; }; }' "$new_rl_file" >> "$new_cl_file"
# Add a divider.
printf -- '---\n\n' >> "$new_cl_file"

[[ -n "$verbose" ]] && printf 'Appending the rest of the changelog file: [%s].\n' "$new_cl_file"
# Get the rest of the changelog file, but remove any existing entry for this version and also any rcs for
# this version (e.g. if the version is v1.2.3, remove v1.2.3-rc1 etc.). If this version is an rc, it shouldn't
# remove the other rcs, though (e.g. if the version is v1.2.3-rc2, we still want v1.2.3-rc1 in there).
awk '{if (!in_bot && /^##[[:space:]]/ && $0 !~ /[Uu][Nn][Rr][Ee][Ll][Ee][Aa][Ss][Ee][Dd]/) { in_bot=1; }; if (in_bot) { print $0; }; }' "$changelog_file" | clean_versions >> "$new_cl_file"

[[ -n "$verbose" ]] && printf 'Putting changelog into place: [%s].\n' "$new_cl_file"
cp "$new_cl_file" "$changelog_file" || clean_exit 1

########################################################################################################################
##################################################  Finalize Release  ##################################################
########################################################################################################################

printf 'Finishing the release.\n'

# Note: I'm not using unlcog release here because they assume you're going to do that before unclog build,
# and will try to open the editor to change the summary. But we don't want that since we've already done
# stuff with the summary and it shouldn't change now.

[[ -n "$verbose" ]] && printf 'Moving unreleased entries to new version dir: [%s].\n' "$new_ver_dir"
mv "$unreleased_dir" "$new_ver_dir" || clean_exit 1
rm "$new_ver_dir/.gitkeep" > /dev/null 2>&1
[[ -n "$verbose" ]] && printf 'Creating new unreleased dir: [%s].\n' "$unreleased_dir"
mkdir -p "$unreleased_dir" || clean_exit 1
touch "$unreleased_dir/.gitkeep"

printf 'Done.\n'
clean_exit 0
