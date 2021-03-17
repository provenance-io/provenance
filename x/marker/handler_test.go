package marker_test

import (
	"strings"
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/marker"
	"github.com/provenance-io/provenance/x/marker/keeper"
)

func TestInvalidMsg(t *testing.T) {
	k := keeper.Keeper{}
	h := marker.NewHandler(k)

	res, err := h(sdk.NewContext(nil, tmproto.Header{}, false, nil), testdata.NewTestMsg())
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, strings.Contains(err.Error(), "unknown message type: Test message"))
}

func TestInvalidProposal(t *testing.T) {
	k := keeper.Keeper{}
	h := marker.NewProposalHandler(k)

	err := h(sdk.NewContext(nil, tmproto.Header{}, false, nil), govtypes.NewTextProposal("Test", "description"))
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "unrecognized marker proposal content type: *types.TextProposal"))
}
