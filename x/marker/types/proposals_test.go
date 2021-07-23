package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProposalAddMarker_Format(t *testing.T) {
	m := NewAddMarkerProposal("title", "description", "test", sdk.NewInt(100), sdk.AccAddress{}, StatusProposed, MarkerType_Coin, []AccessGrant{}, true, true)
	require.NotNil(t, m)

	require.Equal(t, RouterKey, m.ProposalRoute())
	require.Equal(t, ProposalTypeAddMarker, m.ProposalType())

	err := m.ValidateBasic()
	require.Error(t, err)
	require.EqualValues(t, fmt.Errorf("marker manager cannot be empty when creating a proposed marker"), err)

	m.Status = StatusUndefined
	require.Error(t, m.ValidateBasic())

	m.Status = StatusActive
	require.NoError(t, m.ValidateBasic())

	m.Status = StatusProposed
	m.Manager = testAddress().String()

	m.Amount.Denom = "123"
	err = m.ValidateBasic()
	require.Error(t, err)
	require.EqualValues(t, "invalid marker denom/total supply: invalid coins", err.Error())
	m.Amount.Denom = "test"

	m.Title = ""
	err = m.ValidateBasic()
	require.Error(t, err)
	require.EqualValues(t, "proposal title cannot be blank: invalid proposal content", err.Error())
	m.Title = "test"

	require.NoError(t, m.ValidateBasic())

	require.Equal(t, `Add Marker Proposal:
  Marker:      test
  Title:       test
  Description: description
  Supply:      100
  Status:      proposed
  Type:        MARKER_TYPE_COIN
`, m.String())
}

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
