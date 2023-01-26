package keeper_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	provenance "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/pioconfig"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app *provenance.App
	ctx sdk.Context
	k   markerkeeper.Keeper

	accountAddr sdk.AccAddress
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.app = provenance.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.k = markerkeeper.NewKeeper(s.app.AppCodec(), s.app.GetKey(markertypes.ModuleName), s.app.GetSubspace(markertypes.ModuleName), s.app.AccountKeeper, s.app.BankKeeper, s.app.AuthzKeeper, s.app.FeeGrantKeeper, s.app.TransferKeeper, s.app.GetKey(banktypes.StoreKey))
	s.accountAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func (s *IntegrationTestSuite) TestMarkerProposals() {
	// Add markers for tests
	validAuthority := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
	server := markerkeeper.NewMsgServerImpl(s.app.MarkerKeeper)
	_, err := server.AddMarkerProposal(s.ctx, markertypes.NewMsgAddMarkerProposal("test1", sdk.NewInt(100), sdk.AccAddress{}, markertypes.StatusActive, markertypes.MarkerType_Coin, []markertypes.AccessGrant{}, true, true, validAuthority))
	require.NoError(s.T(), err)

	_, err = server.AddMarkerProposal(s.ctx, markertypes.NewMsgAddMarkerProposal("testnogov", sdk.NewInt(100), sdk.AccAddress{}, markertypes.StatusActive, markertypes.MarkerType_Coin, []markertypes.AccessGrant{}, true, false, validAuthority))
	require.NoError(s.T(), err)

	_, err = server.AddMarkerProposal(s.ctx, markertypes.NewMsgAddMarkerProposal("pending", sdk.NewInt(100), s.accountAddr, markertypes.StatusFinalized, markertypes.MarkerType_Coin, []markertypes.AccessGrant{}, true, true, validAuthority))
	require.NoError(s.T(), err)

	_, err = server.AddMarkerProposal(s.ctx, markertypes.NewMsgAddMarkerProposal("testrestricted", sdk.NewInt(100), s.accountAddr, markertypes.StatusFinalized, markertypes.MarkerType_RestrictedCoin, []markertypes.AccessGrant{}, true, true, validAuthority))
	require.NoError(s.T(), err)

	testCases := []struct {
		name string
		prop govtypesv1beta1.Content
		err  error
	}{
		// INCREASE SUPPLY PROPOSALS
		{
			"supply increase - valid",
			markertypes.NewSupplyIncreaseProposal("title", "description", sdk.NewCoin("test1", sdk.NewInt(100)), s.accountAddr.String()),
			nil,
		},
		{
			"supply increase - on finalized marker",
			markertypes.NewSupplyIncreaseProposal("title", "description", sdk.NewCoin("pending", sdk.NewInt(100)), s.accountAddr.String()),
			nil,
		},
		{
			"supply increase - marker doesn't exist",
			markertypes.NewSupplyIncreaseProposal("title", "description", sdk.NewCoin("test", sdk.NewInt(100)), s.accountAddr.String()),
			fmt.Errorf("test marker does not exist"),
		},
		{
			"supply increase - no governance allowed",
			markertypes.NewSupplyIncreaseProposal("title", "description", sdk.NewCoin("testnogov", sdk.NewInt(100)), s.accountAddr.String()),
			fmt.Errorf("testnogov marker does not allow governance control"),
		},
		{
			"supply increase - valid no target",
			markertypes.NewSupplyIncreaseProposal("title", "description", sdk.NewCoin("test1", sdk.NewInt(100)), ""),
			nil,
		},

		// DECREASE SUPPLY PROPOSALS
		{
			"supply decrease - valid",
			markertypes.NewSupplyDecreaseProposal("title", "description", sdk.NewCoin("test1", sdk.NewInt(100))),
			nil,
		},
		{
			"supply decrease - no governance allowed",
			markertypes.NewSupplyDecreaseProposal("title", "description", sdk.NewCoin("testnogov", sdk.NewInt(100))),
			fmt.Errorf("testnogov marker does not allow governance control"),
		},
		{
			"supply decrease - marker doesnot exist",
			markertypes.NewSupplyDecreaseProposal("title", "description", sdk.NewCoin("test", sdk.NewInt(100))),
			fmt.Errorf("test marker does not exist"),
		},

		// WITHDRAW PROPOSALS
		{
			"withdraw - valid",
			markertypes.NewWithdrawEscrowProposal("title", "description", "test1", sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(10))), s.accountAddr.String()),
			nil,
		},
		{
			"withdraw - no governance",
			markertypes.NewWithdrawEscrowProposal("title", "description", "testnogov", sdk.NewCoins(sdk.NewCoin("testnogov", sdk.NewInt(1))), ""),
			fmt.Errorf("testnogov marker does not allow governance control"),
		},
		{
			"withdraw - marker doesnot exist",
			markertypes.NewWithdrawEscrowProposal("title", "description", "test", sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(100))), ""),
			fmt.Errorf("test marker does not exist"),
		},
		{
			"withdraw - invalid recpient",
			markertypes.NewWithdrawEscrowProposal("title", "description", "test1", sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(100))), "bad1address"),
			fmt.Errorf("decoding bech32 failed: invalid checksum (expected dpg8tu got ddress)"),
		},

		// STATUS CHANGE PROPOSALS
		{
			"status change - no governance",
			markertypes.NewChangeStatusProposal("title", "description", "testnogov", markertypes.StatusActive),
			fmt.Errorf("testnogov marker does not allow governance control"),
		},
		{
			"status change - marker doesnot exist",
			markertypes.NewChangeStatusProposal("title", "description", "test", markertypes.StatusActive),
			fmt.Errorf("test marker does not exist"),
		},
		{
			"status change - invalid status",
			markertypes.NewChangeStatusProposal("title", "description", "pending", markertypes.StatusUndefined),
			fmt.Errorf("error invalid marker status undefined"),
		},
		{
			"status change - invalid status order",
			markertypes.NewChangeStatusProposal("title", "description", "test1", markertypes.StatusProposed),
			fmt.Errorf("invalid status transition proposed precedes existing status of active"),
		},
		{
			"status change - valid",
			markertypes.NewChangeStatusProposal("title", "description", "pending", markertypes.StatusActive),
			nil,
		},
		{
			"status change - invalid destroy",
			markertypes.NewChangeStatusProposal("title", "description", "pending", markertypes.StatusDestroyed),
			fmt.Errorf("only cancelled markers can be deleted"),
		},
		{
			"status change - valid cancel",
			markertypes.NewChangeStatusProposal("title", "description", "pending", markertypes.StatusCancelled),
			nil,
		},
		{
			"status change - valid destroy",
			markertypes.NewChangeStatusProposal("title", "description", "pending", markertypes.StatusDestroyed),
			nil,
		},

		// ADD ACCESS
		{
			"add access - no governance",
			markertypes.NewSetAdministratorProposal("title", "description", "testnogov", []markertypes.AccessGrant{{Address: s.accountAddr.String(), Permissions: markertypes.AccessListByNames("mint, burn")}}),
			fmt.Errorf("testnogov marker does not allow governance control"),
		},
		{
			"add access - marker doesnot exist",
			markertypes.NewSetAdministratorProposal("title", "description", "test", []markertypes.AccessGrant{{Address: s.accountAddr.String(), Permissions: markertypes.AccessListByNames("mint, burn")}}),
			fmt.Errorf("test marker does not exist"),
		},
		{
			"add access - transfer only on restricted",
			markertypes.NewSetAdministratorProposal("title", "description", "test1", []markertypes.AccessGrant{{Address: s.accountAddr.String(), Permissions: markertypes.AccessListByNames("mint, burn, transfer")}}),
			fmt.Errorf("invalid access privileges granted: ACCESS_TRANSFER is not supported for marker type MARKER_TYPE_COIN"),
		},
		{
			"add access - valid",
			markertypes.NewSetAdministratorProposal("title", "description", "test1", []markertypes.AccessGrant{{Address: s.accountAddr.String(), Permissions: markertypes.AccessListByNames("mint, burn")}}),
			nil,
		},
		{
			"add access - valid restricted",
			markertypes.NewSetAdministratorProposal("title", "description", "testrestricted", []markertypes.AccessGrant{{Address: s.accountAddr.String(), Permissions: markertypes.AccessListByNames("mint, burn, transfer")}}),
			nil,
		},

		// REMOVE ACCESS
		{
			"remove access - no governance",
			markertypes.NewRemoveAdministratorProposal("title", "description", "testnogov", []string{s.accountAddr.String()}),
			fmt.Errorf("testnogov marker does not allow governance control"),
		},
		{
			"remove access - marker doesnot exist",
			markertypes.NewRemoveAdministratorProposal("title", "description", "test", []string{s.accountAddr.String()}),
			fmt.Errorf("test marker does not exist"),
		},
		{
			"remove access - marker doesnot exist",
			markertypes.NewRemoveAdministratorProposal("title", "description", "test1", []string{"bad1address"}),
			fmt.Errorf("decoding bech32 failed: invalid checksum (expected dpg8tu got ddress)"),
		},
		{
			"remove access - valid",
			markertypes.NewRemoveAdministratorProposal("title", "description", "test1", []string{s.accountAddr.String()}),
			nil,
		},

		// SET DENOM METADATA PROPOSALS
		{
			"set denom metadata - bad denom",
			markertypes.NewSetDenomMetadataProposal("title", "description",
				banktypes.Metadata{
					Description: "some denom description",
					Base:        "bad$char",
					Display:     "badchar",
					Name:        "Bad Char",
					Symbol:      "BC",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    "bad$char",
							Exponent: 0,
							Aliases:  nil,
						},
					},
				},
			),
			errors.New("invalid denom: bad$char"),
		},
		{
			"set denom metadata - marker does not exist",
			markertypes.NewSetDenomMetadataProposal("title", "description",
				banktypes.Metadata{
					Description: "another denom description",
					Base:        "doesnotexist",
					Display:     "doesnotexist",
					Name:        "Does Not Exist",
					Symbol:      "DNE",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    "doesnotexist",
							Exponent: 0,
							Aliases:  nil,
						},
					},
				},
			),
			errors.New("doesnotexist marker does not exist"),
		},
		{
			"set denom metadata - no governance",
			markertypes.NewSetDenomMetadataProposal("title", "description",
				banktypes.Metadata{
					Description: "the best denom description",
					Base:        "testnogov",
					Display:     "testnogov",
					Name:        "Test No Governance",
					Symbol:      "TNG",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    "testnogov",
							Exponent: 0,
							Aliases:  []string{"thenextgeneration"},
						},
					},
				},
			),
			errors.New("testnogov marker does not allow governance control"),
		},
		{
			"set denom metadata - valid",
			markertypes.NewSetDenomMetadataProposal("title", "description",
				banktypes.Metadata{
					Description: "the best denom description",
					Base:        "test1",
					Display:     "test1",
					Name:        "Test One",
					Symbol:      "TONE",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    "test1",
							Exponent: 0,
							Aliases:  []string{"tone"},
						},
					},
				},
			),
			nil,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.T().Run(tc.name, func(t *testing.T) {

			var err error
			switch c := tc.prop.(type) {
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
			case *markertypes.SetDenomMetadataProposal:
				err = markerkeeper.HandleSetDenomMetadataProposal(s.ctx, s.k, c)
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
	pioconfig.SetProvenanceConfig("", 0)
	suite.Run(t, new(IntegrationTestSuite))
}
