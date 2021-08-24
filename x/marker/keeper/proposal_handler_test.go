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
		// ADD MARKER PROPOSALS
		{
			"add marker - valid",
			markertypes.NewAddMarkerProposal("title", "description", "test1", sdk.NewInt(100), sdk.AccAddress{}, markertypes.StatusActive, markertypes.MarkerType_Coin, []markertypes.AccessGrant{}, true, true),
			false,
			nil,
		},
		{
			"add marker - valid",
			markertypes.NewAddMarkerProposal("title", "description", "testrestricted", sdk.NewInt(100), sdk.AccAddress{}, markertypes.StatusActive, markertypes.MarkerType_RestrictedCoin, []markertypes.AccessGrant{}, true, true),
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

		// INCREASE SUPPLY PROPOSALS
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
			"supply increase - valid no target",
			markertypes.NewSupplyIncreaseProposal("title", "description", sdk.NewCoin("test1", sdk.NewInt(100)), ""),
			false,
			nil,
		},

		// DECREASE SUPPLY PROPOSALS
		{
			"supply decrease - valid",
			markertypes.NewSupplyDecreaseProposal("title", "description", sdk.NewCoin("test1", sdk.NewInt(100))),
			false,
			nil,
		},
		{
			"supply decrease - no governance allowed",
			markertypes.NewSupplyDecreaseProposal("title", "description", sdk.NewCoin("testnogov", sdk.NewInt(100))),
			true,
			fmt.Errorf("testnogov marker does not allow governance control"),
		},
		{
			"supply decrease - marker doesnot exist",
			markertypes.NewSupplyDecreaseProposal("title", "description", sdk.NewCoin("test", sdk.NewInt(100))),
			true,
			fmt.Errorf("test marker does not exist"),
		},

		// WITHDRAW PROPOSALS
		{
			"withdraw - valid",
			markertypes.NewWithdrawEscrowProposal("title", "description", "test1", sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(10))), s.accountAddr.String()),
			false,
			nil,
		},
		{
			"withdraw - no governance",
			markertypes.NewWithdrawEscrowProposal("title", "description", "testnogov", sdk.NewCoins(sdk.NewCoin("testnogov", sdk.NewInt(1))), ""),
			true,
			fmt.Errorf("testnogov marker does not allow governance control"),
		},
		{
			"withdraw - marker doesnot exist",
			markertypes.NewWithdrawEscrowProposal("title", "description", "test", sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(100))), ""),
			true,
			fmt.Errorf("test marker does not exist"),
		},
		{
			"withdraw - invalid recpient",
			markertypes.NewWithdrawEscrowProposal("title", "description", "test1", sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(100))), "bad1address"),
			true,
			fmt.Errorf("decoding bech32 failed: checksum failed. Expected dpg8tu, got ddress."),
		},

		// STATUS CHANGE PROPOSALS
		{
			"status change - no governance",
			markertypes.NewChangeStatusProposal("title", "description", "testnogov", markertypes.StatusActive),
			true,
			fmt.Errorf("testnogov marker does not allow governance control"),
		},
		{
			"status change - marker doesnot exist",
			markertypes.NewChangeStatusProposal("title", "description", "test", markertypes.StatusActive),
			true,
			fmt.Errorf("test marker does not exist"),
		},
		{
			"status change - invalid status",
			markertypes.NewChangeStatusProposal("title", "description", "pending", markertypes.StatusUndefined),
			true,
			fmt.Errorf("error invalid marker status undefined"),
		},
		{
			"status change - invalid status order",
			markertypes.NewChangeStatusProposal("title", "description", "test1", markertypes.StatusProposed),
			true,
			fmt.Errorf("invalid status transition proposed precedes existing status of active"),
		},
		{
			"status change - valid",
			markertypes.NewChangeStatusProposal("title", "description", "pending", markertypes.StatusActive),
			false,
			nil,
		},
		{
			"status change - invalid destroy",
			markertypes.NewChangeStatusProposal("title", "description", "pending", markertypes.StatusDestroyed),
			true,
			fmt.Errorf("only cancelled markers can be deleted"),
		},
		{
			"status change - valid cancel",
			markertypes.NewChangeStatusProposal("title", "description", "pending", markertypes.StatusCancelled),
			false,
			nil,
		},
		{
			"status change - valid destroy",
			markertypes.NewChangeStatusProposal("title", "description", "pending", markertypes.StatusDestroyed),
			false,
			nil,
		},

		// ADD ACCESS
		{
			"add access - no governance",
			markertypes.NewSetAdministratorProposal("title", "description", "testnogov", []markertypes.AccessGrant{{Address: s.accountAddr.String(), Permissions: markertypes.AccessListByNames("mint, burn")}}),
			true,
			fmt.Errorf("testnogov marker does not allow governance control"),
		},
		{
			"add access - marker doesnot exist",
			markertypes.NewSetAdministratorProposal("title", "description", "test", []markertypes.AccessGrant{{Address: s.accountAddr.String(), Permissions: markertypes.AccessListByNames("mint, burn")}}),
			true,
			fmt.Errorf("test marker does not exist"),
		},
		{
			"add access - valid",
			markertypes.NewSetAdministratorProposal("title", "description", "test1", []markertypes.AccessGrant{{Address: s.accountAddr.String(), Permissions: markertypes.AccessListByNames("mint, burn")}}),
			false,
			nil,
		},
		{
			"add access - valid restricted",
			markertypes.NewSetAdministratorProposal("title", "description", "testrestricted", []markertypes.AccessGrant{{Address: s.accountAddr.String(), Permissions: markertypes.AccessListByNames("mint, burn, transfer")}}),
			false,
			nil,
		},

		// REMOVE ACCESS
		{
			"remove access - no governance",
			markertypes.NewRemoveAdministratorProposal("title", "description", "testnogov", []string{s.accountAddr.String()}),
			true,
			fmt.Errorf("testnogov marker does not allow governance control"),
		},
		{
			"remove access - marker doesnot exist",
			markertypes.NewRemoveAdministratorProposal("title", "description", "test", []string{s.accountAddr.String()}),
			true,
			fmt.Errorf("test marker does not exist"),
		},
		{
			"remove access - marker doesnot exist",
			markertypes.NewRemoveAdministratorProposal("title", "description", "test1", []string{"bad1address"}),
			true,
			fmt.Errorf("decoding bech32 failed: checksum failed. Expected dpg8tu, got ddress."),
		},
		{
			"remove access - valid",
			markertypes.NewRemoveAdministratorProposal("title", "description", "test1", []string{s.accountAddr.String()}),
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
