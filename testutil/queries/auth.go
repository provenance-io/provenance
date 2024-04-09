package queries

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetModuleAccountByName executes a query to get a module account by name, requiring everything to be okay.
func GetModuleAccountByName(t *testing.T, val *network.Validator, moduleName string) sdk.ModuleAccountI {
	t.Helper()
	rv, ok := AssertGetModuleAccountByName(t, val, moduleName)
	if !ok {
		t.FailNow()
	}
	return rv
}

// AssertGetModuleAccountByName executes a query to get a module account by name, asserting that everything is okay.
// The returned bool will be true on success, or false if something goes wrong.
func AssertGetModuleAccountByName(t *testing.T, val *network.Validator, moduleName string) (sdk.ModuleAccountI, bool) {
	t.Helper()
	url := fmt.Sprintf("%s/cosmos/auth/v1beta1/module_accounts/%s", val.APIAddress, moduleName)
	resp, ok := AssertGetRequest(t, val, url, &authtypes.QueryModuleAccountByNameResponse{})
	if !ok {
		return nil, false
	}
	if !assert.NotNil(t, resp.Account, "module %q account", moduleName) {
		return nil, false
	}

	var acct sdk.AccountI
	err := val.ClientCtx.Codec.UnpackAny(resp.Account, &acct)
	if !assert.NoError(t, err, "UnpackAny(%#v, %T)", resp.Account, &acct) {
		return nil, false
	}

	rv, ok := acct.(sdk.ModuleAccountI)
	if !assert.True(t, ok, "could not cast %T as a sdk.ModuleAccountI", acct) {
		return nil, false
	}

	return rv, true
}
