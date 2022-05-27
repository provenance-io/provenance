#!/usr/bin/env bash
set -ex

#
# Check go.mod for any version updates of tendermint or cosmos-sdk.
# Warn user to update third_party libraries and push again.
#

red='\e[0;31m'
lite_blue='\e[1;34m'
yellow='\e[1;33m'
off='\e[0m'

regex='github.com/cosmos/cosmos-sdk|github.com/tendermint/tendermint'
output=$(git diff HEAD:go.mod..origin/main:go.mod | grep -E -c $regex)
dir="$( cd "$( dirname "${BASH_SOURCE:-$0}" )/.."; pwd -P )"

if [[ $output -gt 0 ]]; then
  echo -e "${lite_blue}Downloading latest third_party proto files for comparison...${off}"

  # Download third_party proto files int build/ directory for comparison against $dir /third_party
  . "$dir"/proto-update-deps.sh build

  echo -e "${lite_blue}Checking Protobuf files for differences..."

  check_diff=$(diff -r -q -x '*.yaml' "$dir"/build/third_party "$dir"/third_party)
  if [[ -z "$check_diff" ]]; then
    echo -e "${lite_blue}Diff log:
    ${yellow}${check_diff}${off}"

    echo -e "
    ${red}Found differences in Protobuf files.
    This indicates a version change was detected in one of the following libraries:
      - github.com/cosmos/cosmos-sdk
      - github.com/tendermint/tendermint

    Review the diff log above and update accordingly. Perform the following steps locally.
      1. run: make proto-update-check${off} (repeat until no diffs are found)
      ${red}2. run: make proto-update-deps\n${off} (to update to latest version)
      ${red}3. commit updates and push" >&2

    exit 1
  fi

  # Version change detected but Protobuf files are already up to date
  exit 0
fi

exit 0
