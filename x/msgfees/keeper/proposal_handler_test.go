package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	msgfeeskeeper "github.com/provenance-io/provenance/x/msgfees/keeper"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"

	provenance "github.com/provenance-io/provenance/app"
)

type IntegrationTestSuite struct {
	suite.Suite

	app *provenance.App
	ctx sdk.Context
	k   msgfeeskeeper.Keeper

	accountAddr sdk.AccAddress
}

func (s *IntegrationTestSuite) SetupSuite() {
	// s.app = provenance.Setup(false)
	// s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	// s.k = msgfeeskeeper.NewKeeper(s.app.AppCodec(), s.app.GetKey(msgfeestypes.ModuleName), s.app.GetSubspace(msgfeestypes.ModuleName), "")
	// s.accountAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func (s *IntegrationTestSuite) TestMarkerProposals() {

	testCases := []struct {
		name string
		prop govtypes.Content
		err  error
	}{
		{
			"add msgfees - valid",
			msgfeestypes.NewAddMsgBasedFeesProposal("title", "description", sdk.NewCoin("hotdog", sdk.NewInt(10)), nil, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10))), sdk.OneDec()),
			nil,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.T().Run(tc.name, func(t *testing.T) {

			var err error
			switch c := tc.prop.(type) {
			case *msgfeestypes.AddMsgBasedFeesProposal:
				err = msgfeeskeeper.HandleAddMsgBasedFeesProposal(s.ctx, s.k, c)
			case *msgfeestypes.UpdateMsgBasedFeesProposal:
				err = msgfeeskeeper.HandleUpdateMsgBasedFeesProposal(s.ctx, s.k, c)
			case *msgfeestypes.RemoveMsgBasedFeesProposal:
				err = msgfeeskeeper.HandleRemoveMsgBasedFeesProposal(s.ctx, s.k, c)
			default:
				panic("invalid proposal type")
			}

			if tc.err != nil {
				require.Error(t, err)
				require.Equal(t, tc.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}

}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
