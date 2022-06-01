#!/usr/bin/env bash
set -ex

#
# Check go.mod for any version updates of tendermint or cosmos-sdk.
# Warn user to update third_party libraries and push again.
#

COLOR_RED=$(TPUT setaf 1)
COLOR_BLUE=$(TPUT setaf 4)
COLOR_OFF=$(TPUT sgr0)

regex='github.com/cosmos/cosmos-sdk|github.com/tendermint/tendermint'
output=$(git diff ..origin/main -- go.mod | grep -E -c $regex)
dir="$( cd "$( dirname "${BASH_SOURCE:-$0}" )/.."; pwd -P )"

# single brackets because of `set -e` option above.
if [ "$output" -gt 0 ]; then
  echo -e "${COLOR_BLUE}Downloading latest third_party proto files for comparison...${COLOR_OFF}"

  # Download third_party proto files int build/ directory for comparison against $dir /third_party
  bash "$dir"/scripts/proto-update-deps.sh build

  echo -e "\n${COLOR_BLUE}Checking Protobuf files for differences...${COLOR_OFF}\n"

  DIFF=$(diff -rq -x '*.yaml' --exclude=google "$dir"/build/third_party "$dir"/third_party || true)
  if [ -n "$DIFF" ]; then
    echo -e "${COLOR_BLUE}\n\nDiff log:\n\n$DIFF${COLOR_OFF}\n\n"

    echo -e "${COLOR_RED}\nFound differences in Protobuf files.\n
    This indicates a version change was detected in one of the following libraries:
      - github.com/cosmos/cosmos-sdk
      - github.com/tendermint/tendermint

    Review the diff log above and update accordingly. Perform the following steps locally.
      1. run: make proto-update-deps
      2. run: make proto-update-check
      3. run: diff -rq -x '*.yaml' --exclude=google build/third_party third_party
      4. commit updates and push\n${COLOR_OFF}" >&2

    exit 1
  fi

  # Version change detected but Protobuf files are already up to date
  exit 0
fi

exit 0
