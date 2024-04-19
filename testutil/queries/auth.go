package queries

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/testutil"
)

// GetModuleAccountByName executes a query to get a module account by name, requiring everything to be okay.
func GetModuleAccountByName(t *testing.T, n *network.Network, moduleName string) sdk.ModuleAccountI {
	t.Helper()
	rv, ok := AssertGetModuleAccountByName(t, n, moduleName)
	if !ok {
		t.FailNow()
	}
	return rv
}

// AssertGetModuleAccountByName executes a query to get a module account by name, asserting that everything is okay.
// The returned bool will be true on success, or false if something goes wrong.
func AssertGetModuleAccountByName(t *testing.T, n *network.Network, moduleName string) (sdk.ModuleAccountI, bool) {
	t.Helper()
	url := fmt.Sprintf("/cosmos/auth/v1beta1/module_accounts/%s", moduleName)
	resp, ok := AssertGetRequest(t, n, url, &authtypes.QueryModuleAccountByNameResponse{})
	if !ok {
		return nil, false
	}
	if !assert.NotNil(t, resp.Account, "module %q account", moduleName) {
		return nil, false
	}

	var acct sdk.AccountI
	err := n.Validators[0].ClientCtx.Codec.UnpackAny(resp.Account, &acct)
	if !assert.NoError(t, err, "UnpackAny(%#v, %T)", resp.Account, &acct) {
		return nil, false
	}

	rv, ok := acct.(sdk.ModuleAccountI)
	if !assert.True(t, ok, "could not cast %T as a sdk.ModuleAccountI", acct) {
		return nil, false
	}

	return rv, true
}

// GetTxFromResponse gets the TxResponse from the provided ExecTestCLICmd output bytes.
// If the provided output indicates a code other than 0, that is what is returned.
// If the provided output has a code of 0, the tx hash is extracted from it, and queries
// are executed to get the TxResponse for it. Three attempts are made to query for the tx,
// waiting a block between each attempt.
//
// This requires that there are no problems getting the TxResponse.
func GetTxFromResponse(t *testing.T, n *network.Network, txRespBz []byte) sdk.TxResponse {
	t.Helper()
	rv, ok := AssertGetTxFromResponse(t, n, txRespBz)
	if !ok {
		t.FailNow()
	}
	return rv
}

// AssertGetTxFromResponse gets the TxResponse from the provided ExecTestCLICmd output bytes.
// If the provided output indicates a code other than 0, that is what is returned.
// If the provided output has a code of 0, the tx hash is extracted from it, and queries
// are executed to get the TxResponse for it. Three attempts are made to query for the tx,
// waiting a block between each attempt.
//
// This asserts that there are no problems getting the TxResponse, returning true if no assertions failed.
func AssertGetTxFromResponse(t *testing.T, n *network.Network, txRespBz []byte) (sdk.TxResponse, bool) {
	t.Helper()
	if !assert.NotEmpty(t, n.Validators, "Network.Validators") {
		return sdk.TxResponse{}, false
	}
	val := n.Validators[0]

	var origResp sdk.TxResponse
	err := val.ClientCtx.Codec.UnmarshalJSON(txRespBz, &origResp)
	if !assert.NoError(t, err, "UnmarshalJSON(%q, %T) (original tx response)", string(txRespBz), &origResp) {
		return sdk.TxResponse{}, false
	}
	if origResp.Code != 0 || len(origResp.RawLog) > 0 || origResp.Tx != nil {
		return origResp, true
	}
	if !assert.NotEmpty(t, origResp.TxHash, "the tx hash") {
		return sdk.TxResponse{}, false
	}

	cmd := authcli.QueryTxCmd()
	args := []string{origResp.TxHash, "--output", "json"}
	var outBZ []byte
	tries := 3
	for i := 1; i <= tries; i++ {
		out, cmdErr := cli.ExecTestCLICmd(val.ClientCtx, cmd, args)
		outBZ = out.Bytes()
		t.Logf("Tx %s result (try %d of %d):\n%s", origResp.TxHash, i, tries, string(outBZ))
		err = cmdErr
		if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("tx (%s) not found", origResp.TxHash)) {
			break
		}
		if i != tries {
			if !assert.NoError(t, testutil.WaitForNextBlock(n), "WaitForNextBlock after try %d of %d", i, tries) {
				return sdk.TxResponse{}, false
			}
		}
	}
	if !assert.NoError(t, err, "ExecTestCLICmd QueryTxCmd %v", args) {
		return sdk.TxResponse{}, false
	}

	var rv sdk.TxResponse
	err = val.ClientCtx.Codec.UnmarshalJSON(outBZ, &rv)
	if !assert.NoError(t, err, "UnmarshalJSON(%q, %T) (tx query response)", string(outBZ), &rv) {
		return sdk.TxResponse{}, false
	}

	return rv, true
}
