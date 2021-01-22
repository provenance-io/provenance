package simulation

import (
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/provenance-io/provenance/x/attribute/keeper"
)

// ProposalContents defines the module weighted proposals' contents (none for attribute)
func ProposalContents(k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{}
}
