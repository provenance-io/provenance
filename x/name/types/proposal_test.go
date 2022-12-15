package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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

func TestNewModifyNameProposal(t *testing.T) {
	crnp := NewModifyNameProposal("test title", "test description", "root", sdk.AccAddress{}, false)

	require.Equal(t, "test title", crnp.GetTitle())
	require.Equal(t, "test description", crnp.GetDescription())
	require.Equal(t, RouterKey, crnp.ProposalRoute())
	require.Equal(t, ProposalTypeModifyName, crnp.ProposalType())
	require.Equal(t, false, crnp.Restricted)
	require.Equal(t, sdk.AccAddress{}.String(), crnp.Owner)
	require.Nil(t, crnp.ValidateBasic())
	require.Equal(t, `Modify Name Proposal:
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

func (s *IntegrationTestSuite) TestModifyNameVariations() {
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
		{"valid proposal but empty name", "test title", "test description", "", addr, false, ErrInvalidLengthName},
		{"valid proposal no owner", "test title", "test description", "root", sdk.AccAddress{}, false, nil},
		{"invalid addr", "test title", "test description", "root", sdk.AccAddress(longAddr), false, ErrInvalidAddress},
		{"invalid gov base proposal", "", "test description", "root", addr, false, fmt.Errorf("proposal title cannot be blank: invalid proposal content")},
		{"valid proposal", "test title", "test description", "root", addr, false, nil},
		{"valid proposal with name hierarchy", "test title", "test description", "test.root", addr, false, nil},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			rn := NewModifyNameProposal(tc.title, tc.description, tc.rootname, tc.owner, tc.restricted)
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
