package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	provenance "github.com/provenance-io/provenance/app"
)

type IntegrationTestSuite struct {
	suite.Suite

	app *provenance.App
	ctx sdk.Context
	k   markerkeeper.Keeper

	accountAddr sdk.AccAddress
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.app = provenance.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.k = markerkeeper.NewKeeper(s.app.AppCodec(), s.app.GetKey(markertypes.ModuleName), s.app.GetSubspace(markertypes.ModuleName), s.app.AccountKeeper, s.app.BankKeeper)
	s.accountAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func (s *IntegrationTestSuite) TestMarkerProposals() {

	testCases := []struct {
		name    string
		prop    govtypes.Content
		wantErr bool
		err     error
	}{
		{
			"add marker - valid",
			markertypes.NewAddMarkerProposal("title", "description", "test1", sdk.NewInt(100), sdk.AccAddress{}, markertypes.StatusActive, markertypes.MarkerType_Coin, []markertypes.AccessGrant{}, true, true),
			false,
			nil,
		},
		{
			"add marker - valid no governance",
			markertypes.NewAddMarkerProposal("title", "description", "testnogov", sdk.NewInt(100), sdk.AccAddress{}, markertypes.StatusActive, markertypes.MarkerType_Coin, []markertypes.AccessGrant{}, true, false),
			false,
			nil,
		},
		{
			"add marker - valid finalized",
			markertypes.NewAddMarkerProposal("title", "description", "pending", sdk.NewInt(100), s.accountAddr, markertypes.StatusFinalized, markertypes.MarkerType_Coin, []markertypes.AccessGrant{}, true, true),
			false,
			nil,
		},
		{
			"add marker - already exists",
			markertypes.NewAddMarkerProposal("title", "description", "test1", sdk.NewInt(0), sdk.AccAddress{}, markertypes.StatusActive, markertypes.MarkerType_Coin, []markertypes.AccessGrant{}, true, true),
			true,
			fmt.Errorf("test1 marker already exists"),
		},
		{
			"add marker - invalid status",
			markertypes.NewAddMarkerProposal("title", "description", "test2", sdk.NewInt(100), sdk.AccAddress{}, markertypes.StatusUndefined, markertypes.MarkerType_Coin, []markertypes.AccessGrant{}, true, true),
			true,
			fmt.Errorf("error invalid marker status undefined"),
		},
		{
			"supply increase - valid",
			markertypes.NewSupplyIncreaseProposal("title", "description", sdk.NewCoin("test1", sdk.NewInt(100)), s.accountAddr.String()),
			false,
			nil,
		},
		{
			"supply increase - on finalized marker",
			markertypes.NewSupplyIncreaseProposal("title", "description", sdk.NewCoin("pending", sdk.NewInt(100)), s.accountAddr.String()),
			false,
			nil,
		},
		{
			"supply increase - marker doesn't exist",
			markertypes.NewSupplyIncreaseProposal("title", "description", sdk.NewCoin("test", sdk.NewInt(100)), s.accountAddr.String()),
			true,
			fmt.Errorf("test marker does not exist"),
		},
		{
			"supply increase - no governance allowed",
			markertypes.NewSupplyIncreaseProposal("title", "description", sdk.NewCoin("testnogov", sdk.NewInt(100)), s.accountAddr.String()),
			true,
			fmt.Errorf("testnogov marker does not allow governance control"),
		},
		{
			"supply decrease - valid",
			markertypes.NewSupplyDecreaseProposal("title", "description", sdk.NewCoin("test1", sdk.NewInt(100))),
			false,
			nil,
		},
		{
			"supply increase - valid no target",
			markertypes.NewSupplyIncreaseProposal("title", "description", sdk.NewCoin("test1", sdk.NewInt(100)), ""),
			false,
			nil,
		},
		{
			"withdraw - empty recpient",
			markertypes.NewWithdrawEscrowProposal("title", "description", "test1", sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(100))), ""),
			true,
			fmt.Errorf("empty address string is not allowed"),
		},
		{
			"withdraw - invalid recpient",
			markertypes.NewWithdrawEscrowProposal("title", "description", "test1", sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(100))), "bad1address"),
			true,
			fmt.Errorf("decoding bech32 failed: checksum failed. Expected dpg8tu, got ddress."),
		},
		{
			"withdraw - valid",
			markertypes.NewWithdrawEscrowProposal("title", "description", "test1", sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(100))), s.accountAddr.String()),
			false,
			nil,
		},
		{
			"status change - valid",
			markertypes.NewChangeStatusProposal("title", "description", "pending", markertypes.StatusActive),
			false,
			nil,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {

			var err error
			switch c := tc.prop.(type) {
			case *markertypes.AddMarkerProposal:
				err = markerkeeper.HandleAddMarkerProposal(s.ctx, s.k, c)
			case *markertypes.SupplyIncreaseProposal:
				err = markerkeeper.HandleSupplyIncreaseProposal(s.ctx, s.k, c)
			case *markertypes.SupplyDecreaseProposal:
				err = markerkeeper.HandleSupplyDecreaseProposal(s.ctx, s.k, c)
			case *markertypes.SetAdministratorProposal:
				err = markerkeeper.HandleSetAdministratorProposal(s.ctx, s.k, c)
			case *markertypes.RemoveAdministratorProposal:
				err = markerkeeper.HandleRemoveAdministratorProposal(s.ctx, s.k, c)
			case *markertypes.ChangeStatusProposal:
				err = markerkeeper.HandleChangeStatusProposal(s.ctx, s.k, c)
			case *markertypes.WithdrawEscrowProposal:
				err = markerkeeper.HandleWithdrawEscrowProposal(s.ctx, s.k, c)
			default:
				panic("invalid proposal type")
			}

			if tc.wantErr {
				s.Require().Error(err)
				s.Require().Equal(tc.err.Error(), err.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}

}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
