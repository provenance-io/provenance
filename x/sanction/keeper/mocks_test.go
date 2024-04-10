package keeper_test

import (
	"context"

	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/x/sanction"
)

// Define a Mock Gov Keeper that records calls to GetProposal and allows
// definition of what it returns.

type MockGovKeeper struct {
	GetProposalCalls   []uint64
	GetProposalReturns map[uint64]govv1.Proposal
}

var _ sanction.GovKeeper = &MockGovKeeper{}

func NewMockGovKeeper() *MockGovKeeper {
	return &MockGovKeeper{
		GetProposalCalls:   nil,
		GetProposalReturns: make(map[uint64]govv1.Proposal),
	}
}

func (k *MockGovKeeper) GetProposal(_ context.Context, proposalID uint64) *govv1.Proposal {
	k.GetProposalCalls = append(k.GetProposalCalls, proposalID)
	prop, ok := k.GetProposalReturns[proposalID]
	if !ok {
		return nil
	}
	return &prop
}
