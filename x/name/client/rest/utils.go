package rest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

type (
	// CreateRootNameProposalReq defines a create root name proposal request body.
	CreateRootNameProposalReq struct {
		BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`

		Title       string         `json:"title" yaml:"title"`
		Description string         `json:"description" yaml:"description"`
		Owner       sdk.AccAddress `json:"owner" yaml:"owner"`
		Name        string         `json:"name" yaml:"name"`
		Restricted  bool           `json:"restricted" yaml:"restricted"`
		Proposer    sdk.AccAddress `json:"proposer" yaml:"proposer"`
		Deposit     sdk.Coins      `json:"deposit" yaml:"deposit"`
	}
)
