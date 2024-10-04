#!/bin/bash
# This script will add a new entry to the unreleased changes.
# It is similar to unclog, but is hopefully a little easier for us to use.

show_usage () {
    cat << EOF
add-change.sh - Add an entry to the unreleased changelog.

Usage: add-change.sh [<num>] [<id>] [<section>] [<message>]

Each argument can alternatively be provided using flags:
    [-n|--issue|--issue-no|-p|--pr|--pull-request] <num>
    [-i|--id] <id>
    [-s|--section] <section>
    [-m|--message] <message>

[<num>] is the issue or pr number.
    This can only contain digits.
    If one is provided (without a flag), it is treated as a PR number.
    If the current branch name has the format .*/<num>-<id>, this arg can be omitted.
    If the <num> comes from the branch name, it is treated as an issue number.
[<id>] is the filename id to use for this entry.
    This can only have alpha-numeric characters, and dashes.
    If not provided, it will be extracted from the current branch name.
    The branch name can be one of .*/<digits>-<id> or .*/<id> or just <id>.
[<section>] is the section directory name that the entry will go in.
    If not provided, you'll be prompted to select it using fzf.
    If one isn't selected (or fzf isn't available), an error is printed.
    This argument can also be a partial match for a valid section as long as it
    matches only a single valid section value. I.e. you can provide this as "bug"
    and it'll automatically use "bug-fixes". But if it's "dep" it'd match either
    "deprecated" or "dependencies", so you'd be prompted using fzf.
[<message>] is the text to include in the changelog.
    If provided, it MUST have at least one space in it.
    If not provided, and the <section> is "dependencies", the get-dep-changes.sh
    script is used to generate the <message>. If not provided, and the <section>
    is anything else, the new entry will be created with a TODO note in it.
    The link will be added to the <message> automatically.
    Multiple <message>s can be provided to create multiple bullet-points in the
    new entry file.

Differentiation between <id> and <section> (when provided without flags):
    Any un-flagged args that neither have spaces nor are all digits are either
    the <id> and/or <section>. It is an error if there are more than three
    of them provided.

    If two such args are provided, then the order is determined by checking
    each for valid section matches. If the 2nd one is a valid section match (or
    neither are), the order is <id> <section>. Otherwise, it's <section> <id>.
    It is an error if there are two of these but either the <id> or <section>
    is provided using flags.

    If one such arg is provided, it is used to fill in what was not provided.
    I.e. if an <id> is provided using flags, it is the <section>. If a <section>
    is provided using flags, it is the <id>. If neither are provided using
    flags, it is the <section> only if it is a valid section match, otherwise it
    is the <id>. It is an error if there is one of these, but both the
    <id> and <section> were provided using flags.

EOF

}

messages=()

while [[ "$#" -gt '0' ]]; do
    case "$1" in
        -h|--help)
            show_usage
            exit 0
            ;;
        -v|--verbose)
            verbose="$1"
            ;;
        -n|--issue|--issue-no)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            num="$2"
            num_type='issue'
            shift
            ;;
        -p|--pr|--pull-request)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            num="$2"
            num_type='pr'
            shift
            ;;
        -i|--id)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            id="$2"
            shift
            ;;
        -s|--section)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            section="$2"
            shift
            ;;
        -m|--message)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s\n' "$1"
                exit 1
            fi
            messages+=( "$2" )
            shift
            ;;
        *)
            if [[ "$1" =~ [[:space:]] ]]; then
                messages+=( "$1" )
            elif [[ "$1" =~ ^[[:digit:]]+$ ]]; then
                if [[ -n "$num" ]]; then
                    printf 'Unknown argument: [%s]. The <num> was already provided as [%s] using a flag.\n' "$1", "$num"
                    exit 1
                fi
                num="$1"
                num_type='pr'
            elif [[ "$1" =~ ^[-[:alnum:]]+$ ]]; then
                if [[ -z "$id_sect_1" ]]; then
                    id_sect_1="$1"
                elif [[ -z "$id_sect_2" ]]; then
                    id_sect_2="$1"
                else
                    printf 'Unknown argument: [%s]. An <id> and <section> were already provided.\n' "$1"
                    exit 1
                fi
            else
                printf 'Unknown argument: [%s].\n' "$1"
                printf 'The <num> can only contain digits.\n'
                printf 'The <id> and <section> can only contain alphanumeric characters and dashes.\n'
                printf 'The <message> must contain at least one space.\n'
                exit 1
            fi
            ;;
    esac
    shift
done

where_i_am="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"
repo_root="$( git rev-parse --show-toplevel 2> /dev/null )"
if [[ -z "$repo_root" ]]; then
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

valid_sections=( $( "${where_i_am}/get-valid-sections.sh" ) )
if [[ "${#valid_sections[@]}" -eq '0' ]]; then
    printf 'Program error: Could not get list of valid sections.\n'
    exit 1
fi

# Usage: validate_section <val>
# If <val> is a valid section, or is a substring of exactly one valid section,
# the valid section name is printed to stdout and this will return with an exit
# code of 0. Otherwise, it'll return exit code of 1 and nothing will be printed.
validate_section () {
    [[ -z "$1" || "$1" =~ [[:space:]] ]] && return 1
    local s opts
    # First look. If there's an exact match, use that.
    # Otherwise, identify all the valid entries that start with the given val.
    opts=()
    for s in "${valid_sections[@]}"; do
        if [[ "$s" == "$1" ]]; then
            printf '%s' "$s"
            return 0
        elif [[ "$s" == "$1"* ]]; then
            opts+=( "$s" )
        fi
    done
    # If there's exactly one option, we're good to go.
    if [[ "${#opts[@]}" -eq '1' ]]; then
        printf '%s' "${opts[*]}"
        return 0
    fi
    # If there were more options, we can't match and can return now.
    [[ "${#opts[@]}" -ne '0' ]] && return 1

    # Second look. Check for it anywhere in the valid entries.
    for s in "${valid_sections[@]}"; do
        if [[ "$s" == *"$1"* ]]; then
            opts+=( "$s" )
        fi
    done
    # If we don't have exactly one match, there's nothing more we can do.
    [[ "${#opts[@]}" -ne '1' ]] && return 1
    printf '%s' "${opts[*]}"
    return 0
}

if [[ -n "$id_sect_2" ]]; then
    # Assume that id_sect_2 is only ever set after id_sect_1, so we've got both here.
    if [[ -n "$id" && -n "$section" ]]; then
        printf 'Unknown arguments: [%s] and [%s]. The id [%s] and section [%s] were provided using flags.\n' "$id_sect_1" "$id_sect_2" "$id" "$section"
        exit 1
    elif [[ -n "$id" ]]; then
        printf 'Unknown arguments: [%s] and [%s]. The id [%s] was provided using flags.\n' "$id_sect_1" "$id_sect_2" "$id"
        exit 1
    elif [[ -n "$section" ]]; then
        printf 'Unknown arguments: [%s] and [%s]. The section [%s] was provided using flags.\n' "$id_sect_1" "$id_sect_2" "$section"
        exit 1
    fi

    # The order is <id> <section> if the 2nd is a valid section (regardless of the first), or if neither are.
    # The order is only ever <section> <id> if the 1st is a valid section, but the 2nd is not.
    if validate_section "$id_sect_2" > /dev/null || ! validate_section "$id_sect_1" > /dev/null; then
        id="$id_sect_1"
        section="$id_sect_2"
    else
        id="$id_sect_2"
        section="$id_sect_1"
    fi
elif [[ -n "$id_sect_1" ]]; then
    if [[ -n "$id" && -n "$section" ]]; then
        printf 'Unknown argument: [%s]. The id [%s] and section [%s] were provided using flags.\n' "$id_sect_1" "$id" "$section"
        exit 1
    fi

    if [[ -z "$id" && -z "$section" ]]; then
        if vs="$( validate_section "$id_sect_1" )"; then
            section="$vs"
        else
            id="$id_sect_1"
        fi
    elif [[ -z "$id" ]]; then
        id="$id_sect_1"
    elif [[ -z "$section" ]]; then
        section="$id_sect_1"
    fi
fi

if [[ -z "$id" || -z "$num" ]]; then
    br="$( git branch --show-current )" || exit 1
    br_sfx="$( sed -E 's|^.*/||' <<< "$br" )"
    if [[ "$br_sfx" =~ ^[[:digit:]]+- ]]; then
        if [[ -z "$num" ]]; then
            num="$( sed -E 's/-.*$//' <<< "$br_sfx" )"
            num_type='issue'
        fi
        br_sfx="$( sed -E 's/^[[:digit:]]+-//' <<< "$br_sfx" )"
    fi
    if [[ -z "$id" ]]; then
        id="$br_sfx"
    fi
fi

if [[ -z "$num" ]]; then
    printf 'No <num> provided, and it could not be determined from the current branch name.\n'
    exit 1
elif [[ "$num" =~ [^[:digit:]] ]]; then
    printf 'Invalid <num>: [%s]. Can only contain digits.\n' "$num"
    exit 1
fi

if [[ -z "$id" ]]; then
    printf 'No <num> provided, and it could not be determined from the current branch name.\n'
    exit 1
elif [[ "$id" =~ [^-[:alnum:]] ]]; then
    printf 'Invalid <id>: [%s]. Can only contain alphanumeric characters and dashes.\n' "$id"
    exit 1
fi

if vs="$( validate_section "$section" )"; then
    section="$vs"
elif command -v fzf > /dev/null 2>&1; then
    # Use fzf to prompt for selection of the desired section.
    # Make the height be the number of entries + 1 for the prompt line. That will make it so all options are
    # visible, and it makes fzf show up below the command prompt (as opposed to taking over the whole screen).
    # Use --layout reverse-list so that the top item is selected first instead of the bottom, and so
    # the options are in the original order (instead of reversed).
    vs="$(
        printf '%s\n' "${valid_sections[@]}" \
            | fzf --cycle --no-multi --no-info --layout reverse-list --query "$section" \
                  --height "$(( ${#valid_sections[@]} + 1 ))" --prompt 'Select a section:'
    )"
    if [[ -n "$vs" ]]; then
        section="$vs"
    fi
fi

if ! validate_section "$section" > /dev/null; then
    printf 'Invalid <section>: [%s].\n' "$section"
    printf 'Valid Sections:\n'
    printf '  %s\n' "${valid_sections[@]}"
    exit 1
fi

if [[ -n "$verbose" ]]; then
    printf '    <num>: [%s].\n' "$num"
    printf '     <id>: [%s].\n' "$id"
    printf '<section>: [%s].\n' "$section"
    if [[ "${#messages[@]}" -gt '0' ]]; then
        printf '<message>: [%s].\n' "${messages[@]}"
    else
        printf '<message>: not provided.\n'
    fi
fi

if [[ "$num_type" == 'issue' ]]; then
    link="[#${num}](https://github.com/provenance-io/provenance/issues/${num})"
    num_type_flag='--issue-no'
elif [[ "$num_type" == 'pr' ]]; then
    link="[PR ${num}](https://github.com/provenance-io/provenance/pull/${num})"
    num_type_flag='--pull-request'
else
    printf 'Program error: Unknown num_type: [%s]. Should be either [issue] or [pr].\n' "$num_type"
    exit 1
fi

if [[ "${#messages[@]}" -eq '0' ]]; then
    if [[ "$section" == "dependencies" ]]; then
        [[ -n "$verbose" ]] && printf 'Using get-dep-changes.sh for new entry.\n'
        "${where_i_am}/get-valid-sections.sh" "$num_type_flag" "$num" --id "$id" --force
        exit $?
    fi
    messages+=( "TODO: Write me." )
fi

[[ -n "$verbose" ]] && printf 'Creating temp file.\n'
temp_file="$( mktemp -t add-change.XXXX )" || exit 1
[[ -n "$verbose" ]] && printf 'Created temp file: %s\n' "$temp_file"

# Usage: clean_exit [<code>]
# Default <code> is 0.
# Deletes the temp file if it exists, then exits.
clean_exit () {
    local ec
    ec="${1:-0}"
    if [[ -n "$temp_file" && -f "$temp_file" ]]; then
        rm -rf "$temp_file" > /dev/null 2>&1
        temp_file=''
    fi
    exit "$ec"
}

[[ -n "$verbose" ]] && printf 'Creating entry content.\n'
for message in "${messages[@]}"; do
    # For the first line, make sure it starts with a "* " and add the link to the end, removing any
    # period from before the link. For all other lines, add two spaces to the beginning of the line.
    awk -v link="$link" '{if(NR==1){sub(/^[[:space:]]*[-*]?[[:space:]]*/,"* "); sub(/[[:space:]]*(\.)?[[:space:]]*$/," " link "."); print;}else{print "  " $0;};}' <<< "$message" > "$temp_file"
done

section_dir="${repo_root}/.changelog/unreleased/${section}"
if [[ ! -d "$section_dir" ]]; then
    [[ -n "$verbose" ]] && printf 'Creating section dir: [%s].\n' "$section_dir"
    mkdir "$section_dir" || clean_exit 1
fi
filename="${section_dir}/${num}-${id}.md"
[[ -n "$verbose" ]] && printf 'Moving temp file [%s] to [%s].\n' "$temp_file" "$filename"
mv "$temp_file" "$filename" || clean_exit 1
temp_file=''

# filename will have the full path to the file, but we want to output the path relative to the repo root.
# This uses parameter expansion to get all characters of filename starting at n+1 and going to the end,
# where n is the number of characters in repo_root. If such parameter expansion isn't available, we
# suppress any error messages and the filename will just have to have everything.
if shorter="${filename:${#repo_root}+1}" 2> /dev/null; then
    filename="$shorter"
fi
printf 'Wrote entry to: %s\n' "${filename}"

clean_exit 0
