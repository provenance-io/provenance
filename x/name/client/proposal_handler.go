package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/provenance-io/provenance/x/name/client/cli"
)

// ProposalHandler is the create root name proposal handler.
var (
	RootNameProposalHandler   = govclient.NewProposalHandler(cli.GetRootNameProposalCmd)
	ModifyNameProposalHandler = govclient.NewProposalHandler(cli.GetModifyNameProposalCmd)
)
