package keeper_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	rewardkeeper "github.com/provenance-io/provenance/x/reward/keeper"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"

	provenance "github.com/provenance-io/provenance/app"
)

type IntegrationTestSuite struct {
	suite.Suite

	app *provenance.App
	ctx sdk.Context
	k   rewardkeeper.Keeper

	accountAddr sdk.AccAddress

	moduleAdd sdk.AccAddress
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.app = provenance.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{Height: 2})
	s.k = rewardkeeper.NewKeeper(s.app.AppCodec(), s.app.GetKey(rewardtypes.ModuleName), s.app.EpochKeeper)
	s.accountAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// s.app.BankKeeper.SendCoinsFromAccountToModule(s.ctx, s.accountAddr, rewardtypes.ModuleName)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) TestRewardProposals() {
	testCases := []struct {
		name string
		prop govtypes.Content
		err  error
	}{
		{
			"add reward - invalid epoch identifier",
			rewardtypes.NewAddRewardProgramProposal("title", "description",
				2,
				s.accountAddr.String(),
				sdk.NewCoin("nhash", sdk.NewInt(10)),
				"night",
				1,
				100,
				rewardtypes.NewEligibilityCriteria("delegation", &rewardtypes.ActionDelegate{}),
			),
			errors.New("invalid epoch identifier: night"),
		},
		{
			"add reward - invalid reward validate basic failure",
			rewardtypes.NewAddRewardProgramProposal("title", "description",
				2,
				s.accountAddr.String(),
				sdk.NewCoin("nhash", sdk.NewInt(0)),
				"day",
				1,
				100,
				rewardtypes.NewEligibilityCriteria("delegation", &rewardtypes.ActionDelegate{}),
			),
			errors.New("reward program requires coins: 0nhash"),
		},
		{
			"add reward - invalid start epoch size",
			rewardtypes.NewAddRewardProgramProposal("title", "description",
				2,
				s.accountAddr.String(),
				sdk.NewCoin("nhash", sdk.NewInt(100)),
				"day",
				0,
				1,
				rewardtypes.NewEligibilityCriteria("delegation", &rewardtypes.ActionDelegate{}),
			),
			errors.New("start epoch 0 cannot be behind current blockheight 2"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.T().Run(tc.name, func(t *testing.T) {
			var err error
			switch c := tc.prop.(type) {
			case *rewardtypes.AddRewardProgramProposal:
				err = rewardkeeper.HandleAddMsgFeeProposal(s.ctx, s.k, c, s.app.InterfaceRegistry())
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
