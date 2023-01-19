package types

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestNewCreateRootNameProposal(t *testing.T) {
	crnp := NewCreateRootNameProposal("test title", "test description", "root", sdk.AccAddress{}, false)

	require.Equal(t, "test title", crnp.GetTitle())
	require.Equal(t, "test description", crnp.GetDescription())
	require.Equal(t, RouterKey, crnp.ProposalRoute())
	require.Equal(t, ProposalTypeCreateRootName, crnp.ProposalType())
	require.Equal(t, false, crnp.Restricted)
	require.Equal(t, sdk.AccAddress{}.String(), crnp.Owner)
	require.Nil(t, crnp.ValidateBasic())
	require.Equal(t, `Create Root Name Proposal:
  Title:       test title
  Description: test description
  Owner:       
  Name:        root
  Restricted:  false
`, crnp.String())
}

type IntegrationTestSuite struct {
	suite.Suite
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
