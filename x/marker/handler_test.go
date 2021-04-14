package marker_test

import (
	"fmt"
	"strings"
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/marker"
	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
)

type HandlerTestSuite struct {
	suite.Suite

	app     *app.App
	ctx     sdk.Context
	handler sdk.Handler

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress
}

func (s *HandlerTestSuite) SetupTest() {
	s.app = app.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.handler = marker.NewHandler(s.app.MarkerKeeper)

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

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

func (s HandlerTestSuite) TestMsgAddMarkerRequest() {
	activeStatus := types.NewAddMarkerRequest("hotdog", sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true)
	activeStatus.Status = types.StatusActive

	undefinedStatus := types.NewAddMarkerRequest("hotdog", sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true)
	undefinedStatus.Status = types.StatusUndefined

	cases := []struct {
		name          string
		msg           *types.MsgAddMarkerRequest
		signers       []string
		errorMsg      string
		expectedEvent *types.EventMarkerAdd
	}{
		{
			"should successfully ADD new marker",
			types.NewAddMarkerRequest("hotdog", sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user1},
			"",
			types.NewEventMarkerAdd("hotdog", "100", "proposed", s.user1, types.MarkerType_Coin.String()),
		},
		{
			"should fail to ADD new marker, validate basic failure",
			undefinedStatus,
			[]string{s.user1},
			"invalid marker status: invalid request",
			nil,
		},
		{
			"should fail to ADD new marker, invalid status",
			activeStatus,
			[]string{s.user1},
			"marker can only be created with a Proposed or Finalized status: invalid request",
			nil,
		},
		{
			"should fail to ADD new marker, marker already exists",
			types.NewAddMarkerRequest("hotdog", sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user1},
			fmt.Sprintf("marker address already exists for %s: invalid request", types.MustGetMarkerAddress("hotdog")),
			nil,
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := s.handler(s.ctx, tc.msg)

			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
				em := s.ctx.EventManager()
				if tc.expectedEvent != nil {
					require.Equal(t, 1, len(em.Events().ToABCIEvents()))
					msg1, _ := sdk.ParseTypedEvent(em.Events().ToABCIEvents()[0])
					require.Equal(t, tc.expectedEvent, msg1)
				}

			}
		})
	}
}
