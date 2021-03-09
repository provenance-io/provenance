package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestProposalAddMarker_Format(t *testing.T) {
	m := NewAddMarkerProposal("title", "description", "test", sdk.NewInt(100), sdk.AccAddress{}, StatusProposed, MarkerType_Coin)
	require.NotNil(t, m)

	require.Equal(t, RouterKey, m.ProposalRoute())
	require.Equal(t, ProposalTypeAddMarker, m.ProposalType())

	err := m.ValidateBasic()
	require.Error(t, err)
	require.EqualValues(t, fmt.Errorf("marker manage cannot be empty when creating a proposed marker"), err)

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
