package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestParameterChangeProposal(t *testing.T) {
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

func (s *IntegrationTestSuite) TestParamChangeVariations() {
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	longAddr := make([]byte, 256)
	testCases := []struct {
		name        string
		title       string
		description string
		rootname    string
		owner       sdk.AccAddress
		restricted  bool
		valError    error
	}{
		{"valid proposal no owner", "test title", "test description", "root", sdk.AccAddress{}, false, nil},
		{"valid proposal", "test title", "test description", "root", addr, false, nil},
		{"invalid name", "test title", "test description", "sub.root", addr, false, ErrNameContainsSegments},
		{"invalid empty name", "test title", "test description", "", addr, false, ErrInvalidLengthName},
		{"invalid addr", "test title", "test description", "root", sdk.AccAddress(longAddr), false, ErrInvalidAddress},
		{"invalid gov base proposal", "", "test description", "root", addr, false, fmt.Errorf("proposal title cannot be blank: invalid proposal content")},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			rn := NewCreateRootNameProposal(tc.title, tc.description, tc.rootname, tc.owner, tc.restricted)
			// in order to evaluate wrapped errors we need to convert to string form for basic evaluation
			if tc.valError != nil {
				err := rn.ValidateBasic()
				s.Require().Error(err)
				if err != nil {
					s.Require().Equal(tc.valError.Error(), err.Error())
				}
			} else {
				s.Require().NoError(rn.ValidateBasic())
			}
		})

	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
