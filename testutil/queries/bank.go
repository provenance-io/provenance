package queries

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// GetAllBalances executes a query to get all balances for an account, requiring everything to be okay.
func GetAllBalances(t *testing.T, n *network.Network, addr string) sdk.Coins {
	t.Helper()
	rv, ok := AssertGetAllBalances(t, n, addr)
	if !ok {
		t.FailNow()
	}
	return rv
}

// AssertGetAllBalances executes a query to get all balances for an account, asserting that everything is okay.
// The returned bool will be true on success, or false if something goes wrong.
func AssertGetAllBalances(t *testing.T, n *network.Network, addr string) (sdk.Coins, bool) {
	t.Helper()
	url := fmt.Sprintf("/cosmos/bank/v1beta1/balances/%s?limit=10000", addr)
	resp, ok := AssertGetRequest(t, n, url, &banktypes.QueryAllBalancesResponse{})
	if !ok {
		return nil, false
	}
	return resp.Balances, true
}

// GetSpendableBalances executes a query to get spendable balances for an account, requiring everything to be okay.
func GetSpendableBalances(t *testing.T, n *network.Network, addr string) sdk.Coins {
	t.Helper()
	rv, ok := AssertGetSpendableBalances(t, n, addr)
	if !ok {
		t.FailNow()
	}
	return rv
}

// AssertGetSpendableBalances executes a query to get spendable balances for an account, asserting that everything is okay.
// The returned bool will be true on success, or false if something goes wrong.
func AssertGetSpendableBalances(t *testing.T, n *network.Network, addr string) (sdk.Coins, bool) {
	t.Helper()
	url := fmt.Sprintf("/cosmos/bank/v1beta1/spendable_balances/%s?limit=10000", addr)
	resp, ok := AssertGetRequest(t, n, url, &banktypes.QuerySpendableBalancesResponse{})
	if !ok {
		return nil, false
	}
	return resp.Balances, true
}
