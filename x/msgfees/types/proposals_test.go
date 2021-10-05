package types

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddMsgBasedFeesProposalType(t *testing.T) {
	m := NewAddMsgBasedFeesProposal("title", "description", sdk.NewCoin("hotdog", sdk.NewInt(10)), nil, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10))), sdk.OneDec())
	require.NotNil(t, m)

	assert.Equal(t, RouterKey, m.ProposalRoute())
	assert.Equal(t, ProposalTypeAddMsgBasedFees, m.ProposalType())
	err := m.ValidateBasic()
	assert.ErrorIs(t, err, ErrEmptyMsgType)

	m.Msg = metadatatypes.MsgWriteRecordRequest{}.Type()

	assert.NoError(t, m.ValidateBasic())

	assert.Equal(t,
		fmt.Sprintf(`Add Msg Based Fees Proposal:
	Title:       %s
	Description: %s
	Amount:      %s
	Msg:         %s
	MinFee:      %s
	FeeRate:     %s
`, m.Title, m.Description, m.Amount, m.Msg, m.MinFee, m.FeeRate),
		m.String())
}
