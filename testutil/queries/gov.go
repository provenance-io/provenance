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
func GetLastGovProp(t *testing.T, val *network.Validator) *govv1.Proposal {
	t.Helper()
	rv, ok := AssertGetLastGovProp(t, val)
	if !ok {
		t.FailNow()
	}
	return rv
}

// AssertGetLastGovProp executes a query to get the most recent governance proposal, asserting that everything is okay.
// The returned bool will be true on success, or false if something goes wrong.
func AssertGetLastGovProp(t *testing.T, val *network.Validator) (*govv1.Proposal, bool) {
	t.Helper()
	url := fmt.Sprintf("%s/cosmos/gov/v1/proposals?limit=1&reverse=true", val.APIAddress)
	resp, ok := AssertGetRequest(t, val, url, &govv1.QueryProposalsResponse{})
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
		err := val.ClientCtx.Codec.UnpackAny(resp.Proposals[0].Messages[i], &msg)
		if !assert.NoError(t, err, "UnpackAny on Messages[%d]", i) {
			return nil, false
		}
	}
	return resp.Proposals[0], true
}

// GetGovProp executes a query to get the requested governance proposal, requiring everything to be okay.
func GetGovProp(t *testing.T, val *network.Validator, propID string) *govv1.Proposal {
	t.Helper()
	rv, ok := AssertGetGovProp(t, val, propID)
	if !ok {
		t.FailNow()
	}
	return rv
}

// AssertGetGovProp executes a query to get the requested governance proposal, asserting that everything is okay.
// The returned bool will be true on success, or false if something goes wrong.
func AssertGetGovProp(t *testing.T, val *network.Validator, propID string) (*govv1.Proposal, bool) {
	t.Helper()
	url := fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s", val.APIAddress, propID)
	resp, ok := AssertGetRequest(t, val, url, &govv1.QueryProposalResponse{})
	if !ok {
		return nil, false
	}
	if !assert.NotNil(t, resp.Proposal, "governance proposal %d", propID) {
		return nil, false
	}
	// Unpack all the proposal messages so that the cachedValue is set in them.
	for i := range resp.Proposal.Messages {
		var msg sdk.Msg
		err := val.ClientCtx.Codec.UnpackAny(resp.Proposal.Messages[i], &msg)
		if !assert.NoError(t, err, "UnpackAny on Messages[%d]", i) {
			return nil, false
		}
	}
	return resp.Proposal, true
}
