package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/provenance-io/provenance/x/name/client/cli"
	"github.com/provenance-io/provenance/x/name/client/rest"
)

// ProposalHandler is the create root name proposal handler.
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetRootNameProposalCmd, rest.RootNameProposalHandler)
)
