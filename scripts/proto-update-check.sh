#!/usr/bin/env bash
#set -ex

#
# Check go.mod for any version updates of tendermint or cosmos-sdk.
# Warn user to update third_party libraries and push again.
#

red='\e[0;31m'
green='\e[0;32m'
off='\e[0m'

regex='github.com/cosmos/cosmos-sdk|github.com/tendermint/tendermint'
output=$(git diff HEAD:go.mod..origin/main:go.mod | grep -E -c $regex)

if [[ $output -gt 0 ]]; then
  echo -e "
  ${red}A version change was detected in one of the following libraries:
    - github.com/cosmos/cosmos-sdk
    - github.com/tendermint/tendermint

  Run the command below to update the third_party proto files and push again.
    $ make proto-update-deps\n${off}" >&2
  exit 1
else
  echo -e "${green}third_party Protobuf files are up to date.\n${off}"
  exit 0
fi
