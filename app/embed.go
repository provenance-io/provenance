//go:build embed_account_data
// +build embed_account_data

package app

import (
	_ "embed"
)

// The testdata/accounts.json file is 35M, so it's not checked into git, and doesn't exist for most.
// If you've got that file in place and want to use it, add the embed_account_data build tag when running stuff in here.

//go:embed testdata/accounts.json
var AccountDataBz []byte

func init() {
	AccountData = AccountDataBz
}
