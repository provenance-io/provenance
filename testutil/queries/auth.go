package queries

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
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

// GetTxFromResponse extracts a tx hash from the provided txRespBz and executes a query for it,
// requiring everything to be okay. Since the SDK got rid of block broadcast, we now need to query
// for a tx after submitting it in order to find out what happened.
//
// The provided txRespBz should be the bytes returned from submitting a Tx.
//
// In most cases, you'll have to wait for the next block after submitting your tx, and before calling this.
func GetTxFromResponse(t *testing.T, val *network.Validator, txRespBz []byte) sdk.TxResponse {
	t.Helper()
	rv, ok := AssertGetTxFromResponse(t, val, txRespBz)
	if !ok {
		t.FailNow()
	}
	return rv
}

// AssertGetTxFromResponse extracts a tx hash from the provided txRespBz and executes a query for it,
// asserting that everything is okay. Since the SDK got rid of block broadcast, we now need to query
// for a tx after submitting it in order to find out what happened.
//
// The provided txRespBz should be the bytes returned from submitting a Tx.
//
// In most cases, you'll have to wait for the next block after submitting your tx, and before calling this.
func AssertGetTxFromResponse(t *testing.T, val *network.Validator, txRespBz []byte) (sdk.TxResponse, bool) {
	var origResp sdk.TxResponse
	err := val.ClientCtx.Codec.UnmarshalJSON(txRespBz, &origResp)
	if !assert.NoError(t, err, "UnmarshalJSON(%q, %T) (original tx response)", string(txRespBz), &origResp) {
		return sdk.TxResponse{}, false
	}
	if !assert.NotEmpty(t, origResp.TxHash, "the tx hash") {
		return sdk.TxResponse{}, false
	}

	cmd := authcli.QueryTxCmd()
	args := []string{origResp.TxHash, "--output", "json"}
	out, err := cli.ExecTestCLICmd(val.ClientCtx, cmd, args)
	outBz := out.Bytes()
	t.Logf("Tx %s result:\n%s", origResp.TxHash, string(outBz))
	if !assert.NoError(t, err, "ExecTestCLICmd QueryTxCmd %v", args) {
		return sdk.TxResponse{}, false
	}

	var rv sdk.TxResponse
	err = val.ClientCtx.Codec.UnmarshalJSON(outBz, &rv)
	if !assert.NoError(t, err, "UnmarshalJSON(%q, %T) (tx query response)", string(outBz), &rv) {
		return sdk.TxResponse{}, false
	}

	return rv, true
}
