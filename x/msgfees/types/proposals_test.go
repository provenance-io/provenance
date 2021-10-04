package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	yaml "gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddMsgBasedFeesProposalType(t *testing.T) {
	m := NewAddMsgBasedFeesProposal("title", "description", sdk.NewCoin("hotdog", sdk.NewInt(10)), nil, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10))), sdk.OneDec())
	require.NotNil(t, m)

	assert.Equal(t, RouterKey, m.ProposalRoute())
	assert.Equal(t, ProposalTypeAddMsgBasedFees, m.ProposalType())
	err := m.ValidateBasic()
	assert.Equal(t, ErrEmptyMsgType, err)

	assert.NoError(t, m.ValidateBasic())

	out, _ := yaml.Marshal(m)
	assert.Equal(t,
		string(out),
		m.String())
}
