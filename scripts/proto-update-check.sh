#!/usr/bin/env bash
set -ex

#
# Check go.mod for any version updates of tendermint or cosmos-sdk.
# Warn user to update third_party libraries and push again.
#

if [ "$1" == '--force' ]; then
    output='1'
else
    regex='github.com/cosmos/cosmos-sdk|github.com/tendermint/tendermint'
    output=$(git diff ..origin/main -- go.mod | grep -E -c "$regex" || true)
fi

# single brackets because the github runners don't have the enhanced (double-bracket) version.
if [ "$output" -gt "0" ]; then
  printf 'Downloading latest third_party proto files for comparison...\n'

  # Download third_party proto files in the build/ directory for comparison against $dir /third_party
  dir="$( cd "$( dirname "${BASH_SOURCE:-$0}" )/.."; pwd -P )"
  bash "$dir/scripts/proto-update-deps.sh" build

  printf '\nChecking Protobuf files for differences...\n'

  DIFF=$( diff -rq -x '*.yaml' --exclude=google "$dir/build/third_party" "$dir/third_party" || true )
  if [ -n "$DIFF" ]; then
    cat << EOF >&2

Diff log:

$DIFF

Found differences in Protobuf files.
This indicates a version change was detected in one of the following libraries:
  - github.com/cosmos/cosmos-sdk
  - github.com/tendermint/tendermint

Review the diff log above and update accordingly. Perform the following steps locally.
  1. run: make proto-update-deps
  2. rerun: make proto-update-check  and make sure it passes.
  3. commit updates and push

EOF
    exit 1
  fi

  # Version change detected but Protobuf files are already up to date
  exit 0
fi

exit 0
