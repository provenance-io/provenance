#!/usr/bin/env bash
# This script looks for all uses of a time.Now() function.
# If any improper uses are found, they will be outputted and the script will exit with code 1.
# If there isn't anything of concern, nothing will be outputted, and the script will exit with code 0.
# Providing the -v or --verbose flag (or exporting VERBOSE=1) will make this output middle-step information.

# This exists because use of time.Now() in the processing of a block (or Tx) is an easy way to halt a chain.
# All time-based validation/processing/calculations should be done against the block time.
# Doing them against time.Now() makes the outcome change depending on when it's run.
# That might sound like something desirable, but consider a chain that is starting/catching up by replaying old blocks.
# If using time.Now(), a Tx that was fine when it was run originally might now fail validation.
# If someone desides to be malicious, then could send a Tx with a time field set a few seconds in the future.
# The block creater processes the Tx and includes it in the block.
# Then some or none of the other validators agree as they're checking it a few seconds later.

if [ "$1" == '-v' ] || [ "$1" == '--verbose' ]; then
    VERBOSE='1'
fi

# Find all files that import the "time" module, and identify what they're referenced as in that file.
# There's at least one other module that has a .Now() function ("github.com/tendermint/tendermint/types/time") that we don't want used.
# So we're actually finding an import of any "time" package.
# If a file has multiple such imports, there'll be a line for each import.
# The vendor directory is ignored since it's not under our control and has lots of matches (that hopefully don't cause problems).
# Output is in the format of "<file>:<alias>", e.g. "app/test_helpers.go:time".
#
# First, find all go files of possible interest.
# Then, grep each go file looking for an import of a "time" package.
# Transform each entry into the format "<file>:<alias>" by:
#  1. Changing ': <alias> "time"' or ':import <alias> "time"' into ':<alias>'.
#  2. Changing ': "time"' or ':import "time"' into ':time'.
#  3. Removing the leading './' on the filename.
# Then sort them because I'm pedantic like that.
time_imports="$( \
    find . -type f -name '*.go' -not -path '*/vendor/*' \
    | xargs grep -E '^(import)?[[:space:]]+([[:alnum:]]+[[:space:]]+)?"([^"]+/)?time"' \
    | sed -E -e 's/:(import)?[[:space:]]+([[:alnum:]]+)[[:space:]]+"([^"]+\/)?time".*$/:\2/' \
             -e 's/:(import)?[[:space:]]+"([^"]+\/)?time".*$/:time/' \
             -e 's/^\.\///' \
    | sort
)"
[ -n "$VERBOSE" ] && printf 'Time imports:\n%s\n\n' "$( sed 's/^/  /' <<< "$time_imports" )"

# Find all uses of a time.Now() function.
# Loop through each line of time imports.
# Split the line into the file and the alias.
# Grep the file for <alias>.Now() taking care to not include entries of a different alias that ends the same.
# Then, we want to ignore the line if the only "use" of <alias>.Now() is actually in a comment.
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
[ -n "$VERBOSE" ] && printf 'All uses of time.Now():\n%s\n\n' "$( sed 's/^/  /' <<< "$now_uses" )"

# Ignore known legitimate uses of time.Now().
# These are controlled in this script rather than through a nolint directive because:
#  a) Use of time.Now() is very dangerous, and it should be harder to use it than just adding a comment on the line.
#  b) This isn't a full-fledged, actual linter.
filters=()

# Any use in a unit test file is okay (but maybe frowned upon).
#     These are removed here (instead of in the find command) so that they can be included in verbose output above.
filters+=( '^[^:]+_test\.go:' )
# It's okay to use it in the telemetry.MeasureSince and telemetry.ModuleMeasureSince functions.
filters+=( 'telemetry\.(Module)?MeasureSince\(' )
# There's a use in the x/reward/simulation/operations.go file that's okay.
#     It's pretty generic though, so rather than ignoring all such lines in the file, only ignore
#     such lines from line 70 to 85 (inclusive). It's on line 78 as of writing this.
filters+=( '^x/reward/simulation/operations\.go:(7[0-9]|8[0-5]):[[:space:]]+now := [[:alnum:]]+\.Now\(\)$' )
# The app/test_helpers.go file also has a legitimate use since it's only for unit tests.
#     It's in the header creation for the BeginBlock.
#     Since it's expected that it might move, and also that additional
#     such uses might be added, allow it to be on any line number.
filters+=( '^app/test_helpers\.go:[[:digit:]]+:.*cmtproto\.Header{' )
# The x/marker/client/cli/tx.go file has two legitimate uses due to authz and feegrant grant creation.
#     Since that file is not involved in any block processing, just ignore the whole file.
filters+=( '^x/marker/client/cli/tx\.go:' )
# The cmd/provenanced/cmd/testnet.go file needs to use it to properly create the genesis file.
#     Since it's setting a variable more specifically named than 'now',
#     we can ignore the specific line, but let it be on any line number.
filters+=( '^cmd/provenanced/cmd/testnet\.go:[[:digit:]]+:[[:space:]]+genTime := [[:alnum:]]+\.Now\(\)$' )

for filters in "${filters[@]}"; do
    now_uses="$( grep -vE "$filters" <<< "$now_uses" )"
    [ -n "$VERBOSE" ] && printf 'After filter %s:\n%s\n\n' "'$filters'" "$( sed 's/^/  /' <<< "$now_uses" )"
done

# If there's anything left, it's bad.
if [ -n "$now_uses" ]; then
    printf 'Improper use(s) of time.Now():\n%s\n' "$now_uses"
    exit 1
fi
[ -n "$VERBOSE" ] && printf 'No improper uses of .Now().\n'
exit 0
