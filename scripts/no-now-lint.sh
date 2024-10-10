#!/usr/bin/env bash
# This script looks for all uses of a .Now() function.
# If any improper uses are found, they will be printed and the script will exit with code 1.
# If there isn't anything of concern, nothing will be printed, and the script will exit with code 0.
# Providing the -v or --verbose flag (or exporting VERBOSE=1) will make this output middle-step information.

# This exists because use of time.Now() in the processing of a block (or Tx) is an easy way to halt a chain.
# All time-based validation/processing/calculations should be done against the block time.
# Doing them against time.Now() makes the outcome change depending on when it's run.
# That might sound like something desirable, but consider a chain that is starting/catching up by replaying old blocks.
# If using time.Now(), a Tx that was fine when it was run originally might now fail validation.
# If someone decides to be malicious, then could send a Tx with a time field set a few seconds in the future.
# The block creator processes the Tx and includes it in the block.
# Then some or none of the other validators agree as they're checking it a few seconds later.
#
# Known libraries that we know have a .Now() function that we need to worry about:
#  * "time"
#  * "github.com/cometbft/cometbft/types/time"
#  * "github.com/cosmos/cosmos-sdk/telemetry"

# Usage: get_line_count "$var"
get_line_count () {
    # Can't use a herestring here because that automatically adds an ending newline.
    # That added line ending causes any string without an ending newline (even an
    # empty string) to report one more line than there actually is.
    # Not using wc -l for this because it won't count a line if there's no ending newline.
    # Not using echo because it doesn't have a standard ending newline behavior. I.e.
    # some shells will automatically add a newline, some won't, and it's flags can mean
    # different things on different systems. It's just best to avoid it.
    # Examples of commands that don't work like we want in here:
    #   wc -l <<< ""                                  # actual = 1 (want = 0).
    #   wc -l <<< "something"$'\n'                    # actual = 2 (want = 1).
    #   printf '%s' "something" | wc -l               # actual = 0 (want = 1).
    #   printf '%s' "something"$'\n'"else" | wc -l    # actual = 1 (want = 2)
    #   grep -c '^' <<< ""                            # actual = 1 (want = 0).
    #   grep -c '^' <<< "something"$'\n'              # actual = 2 (want = 1).
    # But this command is Goldilocks:
    printf '%s' "$1" | grep -c '^'
}

if [ "$1" == '-v' ] || [ "$1" == '--verbose' ]; then
    VERBOSE='1'
fi

# Find all files that import a "time" or "telemetry" module, and identify what
# they're referenced as in that file. We cast a wide net here in the hopes that
# if call to a .Now() in a new library is added, there's a chance this catches it.
# So we're actually finding an import of any "time" or "telemetry" package.
# If a file has multiple such imports, there'll be a line for each import.
# The vendor directory is ignored since it's not under our control and has lots
# of matches (that hopefully don't cause problems).
#
# First, find all go files of possible interest.
#   Line format: './<file path>'
# Then, grep each go file looking for an import of a "time" or "telemetry" package.
#   Line format: './<file path>:\t<import line>'
# Transform each entry into the format "<file>:<alias>" by:
#  1. Changing ': <alias> "<library>"' or ':import <alias> "<library>"' into ':<alias>'.
#  2. Changing ': "<library>"' or ':import "<library>"' into ':<library's last section>'.
#  3. Removing the leading './' on the filename.
# Then sort them because I'm pedantic like that.
# Output is in the format of "<file>:<alias>", e.g. "app/test_helpers.go:time".
time_imports="$( \
    find . -type f -name '*.go' -not -path '*/vendor/*' -print0 \
    | xargs -0 grep -E '^(import)?[[:space:]]+([[:alnum:]]+[[:space:]]+)?"([^"]+/)?(time|telemetry)"' \
    | sed -E -e 's/:(import)?[[:space:]]+([[:alnum:]]+)[[:space:]]+"([^"]+\/)?(time|telemetry)".*$/:\2/' \
             -e 's/:(import)?[[:space:]]+"([^"]+\/)?(time|telemetry)".*$/:\3/' \
             -e 's/^\.\///' \
    | sort
)"
[ -n "$VERBOSE" ] && printf 'Imports of Interest (%d):\n%s\n\n' "$( get_line_count "$time_imports" )" "$( sed 's/^/  /' <<< "$time_imports" )"

# Find all uses of a .Now() function in a library of interest.
# Loop through each line of time_imports.
# Split the line into the <file> and the <alias>.
# Grep the <file> for <alias>.Now() taking care to not include entries of a different alias that ends the same.
# Then, we want to ignore the line if the only use of <alias>.Now() is actually just in a comment.
# But we still want the whole line in the output if it's a match.
# So we remove any comment from each matching line and re-test it. If it's still a match, include the whole line.
now_uses="$( \
    while IFS= read -r line; do
        file="$( sed 's/:.*$//' <<< "$line" )"
        alias="$( sed 's/^.*://' <<< "$line" )"
        rx="[^[:alnum:]]$alias\.Now\(\)"
        match_lines="$( grep -EHn "$rx" "$file" )"
        if [ -n "$match_lines" ]; then
            while IFS= read -r match; do
                if sed 's/\/\/.*$//;' <<< "$match" | grep -q "$rx" 2>&1; then
                    printf '%s\n' "$match"
                fi
            done <<< "$match_lines"
        fi
    done <<< "$time_imports"
)"
[ -n "$VERBOSE" ] && printf 'All uses of .Now() (%d):\n%s\n\n' "$( get_line_count "$now_uses" )" "$( sed 's/^/  /' <<< "$now_uses" )"

# Ignore known legitimate uses of .Now().
# These are controlled in this script rather than through a nolint directive because:
#  a) Use of time.Now() is very dangerous, and it should be harder to use it than just adding a comment on the line.
#  b) This isn't a full-fledged, actual linter.
filters=()

# Any use in a unit test file is okay (but maybe frowned upon).
#     These are removed here (instead of in the find command) so that they can be included in verbose output above.
filters+=( '^[^:]+_test\.go:' )
# It's okay to use it in the telemetry.MeasureSince and telemetry.ModuleMeasureSince functions.
filters+=( 'telemetry\.(Module)?MeasureSince\(' )
# The x/marker/client/cli/tx.go file has two legitimate uses due to authz and feegrant grant creation.
#     Since that file is not involved in any block processing, we can just ignore the whole file.
filters+=( '^x/marker/client/cli/tx\.go:' )
# The cmd/provenanced/cmd/testnet.go file needs to use it to properly create the genesis file.
#     Since it's setting a variable more specifically named than 'now',
#     we can ignore the specific line, but let it be on any line number.
filters+=( '^cmd/provenanced/cmd/testnet\.go:[[:digit:]]+:[[:space:]]+genTime := [[:alnum:]]+\.Now\(\)$' )

for filter in "${filters[@]}"; do
    now_uses="$( grep -vE "$filter" <<< "$now_uses" )"
    [ -n "$VERBOSE" ] && printf 'After filter %s (%d):\n%s\n\n' "'$filter'" "$( get_line_count "$now_uses" )" "$( sed 's/^/  /' <<< "$now_uses" )"
done

# If there's anything left, it's bad.
if [ -n "$now_uses" ]; then
    printf 'Improper use(s) of .Now() (%d):\n%s\n' "$( get_line_count "$now_uses" )" "$now_uses"
    exit 1
fi
[ -n "$VERBOSE" ] && printf 'No improper uses of .Now().\n'
exit 0
