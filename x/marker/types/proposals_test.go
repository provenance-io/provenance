package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProposalTypeIncreaseSupply_Format(t *testing.T) {
	m := NewSupplyIncreaseProposal("title", "description", sdk.NewCoin("test", sdk.NewInt(100)), "")
	require.NotNil(t, m)

	require.Equal(t, RouterKey, m.ProposalRoute())
	require.Equal(t, ProposalTypeIncreaseSupply, m.ProposalType())

	err := m.ValidateBasic()
	require.NoError(t, err)

	require.Equal(t, `MarkerAccount Token Supply Increase Proposal:
  Marker:      test
  Title:       title
  Description: description
  Amount to Increase: 100
`, m.String())
}

func TestProposalTypeDecreaseSupply_Format(t *testing.T) {
	m := NewSupplyDecreaseProposal("title", "description", sdk.NewCoin("test", sdk.NewInt(100)))
	require.NotNil(t, m)

	require.Equal(t, RouterKey, m.ProposalRoute())
	require.Equal(t, ProposalTypeDecreaseSupply, m.ProposalType())

	err := m.ValidateBasic()
	require.NoError(t, err)
	require.Equal(t, `MarkerAccount Token Supply Decrease Proposal:
  Marker:      test
  Title:       title
  Description: description
  Amount to Decrease: 100
`, m.String())
}

func TestProposalTypeSetAdministrator_Format(t *testing.T) {
	addr := testAddress()
	m := NewSetAdministratorProposal("title", "description", "test", []AccessGrant{*NewAccessGrant(addr, AccessListByNames("mint"))})
	require.NotNil(t, m)

	require.Equal(t, RouterKey, m.ProposalRoute())
	require.Equal(t, ProposalTypeSetAdministrator, m.ProposalType())

	err := m.ValidateBasic()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`MarkerAccount Set Administrator Proposal:
  Marker:      test
  Title:       title
  Description: description
  Administrator Access Grant: [AccessGrant: %s [mint]]
`, addr), m.String())
}

func TestProposalTypeRemoveAdministrator_Format(t *testing.T) {
	addr := testAddress()
	m := NewRemoveAdministratorProposal("title", "description", "test", []string{addr.String()})
	require.NotNil(t, m)

	require.Equal(t, RouterKey, m.ProposalRoute())
	require.Equal(t, ProposalTypeRemoveAdministrator, m.ProposalType())

	err := m.ValidateBasic()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`MarkerAccount Remove Administrator Proposal:
  Marker:      test
  Title:       title
  Description: description
  Administrators To Remove: [%s]
`, addr), m.String())
}

func TestProposalTypeChangeStatus_Format(t *testing.T) {
	m := NewChangeStatusProposal("title", "description", "test", StatusProposed)
	require.NotNil(t, m)

	require.Equal(t, RouterKey, m.ProposalRoute())
	require.Equal(t, ProposalTypeChangeStatus, m.ProposalType())

	err := m.ValidateBasic()
	require.NoError(t, err)
	require.Equal(t, `MarkerAccount Change Status Proposal:
  Marker:      test
  Title:       title
  Description: description
  Change Status To: proposed
`, m.String())
}

func TestProposalTypeWithdrawEscrow_Format(t *testing.T) {
	addr := testAddress()
	m := NewWithdrawEscrowProposal("title", "description", "test", sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(100))), addr.String())
	require.NotNil(t, m)

	require.Equal(t, RouterKey, m.ProposalRoute())
	require.Equal(t, ProposalTypeWithdrawEscrow, m.ProposalType())

	err := m.ValidateBasic()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`MarkerAccount Withdraw Escrow Proposal:
  Marker:      test
  Title:       title
  Description: description
  Withdraw 100test and transfer to %s
`, addr), m.String())
}

func TestProposalTypeSetDenomMetadataProposal_Format(t *testing.T) {
	m := NewSetDenomMetadataProposal("sdmdptitle", "sdmdpdescription",
		banktypes.Metadata{
			Description: "test md description",
			Base:        "testmd",
			Display:     "testmdd",
			Name:        "Test Metadata",
			Symbol:      "TMD",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    "testmd",
					Exponent: 0,
					Aliases:  []string{"atestmd", "btestmd"},
				},
				{
					Denom:    "testmdd",
					Exponent: 5,
					Aliases:  []string{"ctestmd", "dtestmd"},
				},
			},
		},
	)
	require.NotNil(t, m)

	assert.Equal(t, RouterKey, m.ProposalRoute())
	assert.Equal(t, ProposalTypeSetDenomMetadata, m.ProposalType())

	err := m.ValidateBasic()
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(`Set Denom Metadata Proposal:
  Marker:      testmd
  Title:       sdmdptitle
  Description: sdmdpdescription
  Metadata:    %s
`, m.Metadata.String()), m.String())
}
