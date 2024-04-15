package queries

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// GetLastGovProp executes a query to get the most recent governance proposal, requiring everything to be okay.
func GetLastGovProp(t *testing.T, n *network.Network) *govv1.Proposal {
	t.Helper()
	rv, ok := AssertGetLastGovProp(t, n)
	if !ok {
		t.FailNow()
	}
	return rv
}

// AssertGetLastGovProp executes a query to get the most recent governance proposal, asserting that everything is okay.
// The returned bool will be true on success, or false if something goes wrong.
func AssertGetLastGovProp(t *testing.T, n *network.Network) (*govv1.Proposal, bool) {
	t.Helper()
	url := "/cosmos/gov/v1/proposals?limit=1&reverse=true"
	resp, ok := AssertGetRequest(t, n, url, &govv1.QueryProposalsResponse{})
	if !ok {
		return nil, false
	}
	if !assert.NotEmpty(t, resp.Proposals, "returned proposals") {
		return nil, false
	}
	if !assert.NotNil(t, resp.Proposals[0], "most recent proposal") {
		return nil, false
	}
	// Unpack all the proposal messages so that the cachedValue is set in them.
	for i := range resp.Proposals[0].Messages {
		var msg sdk.Msg
		err := n.Validators[0].ClientCtx.Codec.UnpackAny(resp.Proposals[0].Messages[i], &msg)
		if !assert.NoError(t, err, "UnpackAny on Messages[%d]", i) {
			return nil, false
		}
	}
	return resp.Proposals[0], true
}

// GetGovProp executes a query to get the requested governance proposal, requiring everything to be okay.
func GetGovProp(t *testing.T, n *network.Network, propID string) *govv1.Proposal {
	t.Helper()
	rv, ok := AssertGetGovProp(t, n, propID)
	if !ok {
		t.FailNow()
	}
	return rv
}

// AssertGetGovProp executes a query to get the requested governance proposal, asserting that everything is okay.
// The returned bool will be true on success, or false if something goes wrong.
func AssertGetGovProp(t *testing.T, n *network.Network, propID string) (*govv1.Proposal, bool) {
	t.Helper()
	url := fmt.Sprintf("/cosmos/gov/v1/proposals/%s", propID)
	resp, ok := AssertGetRequest(t, n, url, &govv1.QueryProposalResponse{})
	if !ok {
		return nil, false
	}
	if !assert.NotNil(t, resp.Proposal, "governance proposal %d", propID) {
		return nil, false
	}
	// Unpack all the proposal messages so that the cachedValue is set in them.
	for i := range resp.Proposal.Messages {
		var msg sdk.Msg
		err := n.Validators[0].ClientCtx.Codec.UnpackAny(resp.Proposal.Messages[i], &msg)
		if !assert.NoError(t, err, "UnpackAny on Messages[%d]", i) {
			return nil, false
		}
	}
	return resp.Proposal, true
}
