package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/marker/types"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	startBlockTime time.Time

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress
}

func TestKeeperTestSuite(t *testing.T) {
	pioconfig.SetProvenanceConfig("", 0)
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.startBlockTime = time.Now()
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: s.startBlockTime})

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()
}

func (s *KeeperTestSuite) TestSupplyIncreaseProposal() {
	hotdogMarker := s.createTestMarker("hotdog")
	nonGovernanceMarker := s.createTestMarker("nonGovernanceMarker")
	nonActiveMarker := s.createTestMarker("nonActiveMarker")
	invalidDenom := "invaliddenom"

	nonGovernanceMarker.AllowGovernanceControl = false
	s.app.MarkerKeeper.SetMarker(s.ctx, nonGovernanceMarker)

	nonActiveMarker.SetStatus(types.StatusCancelled)
	s.app.MarkerKeeper.SetMarker(s.ctx, nonActiveMarker)

	tests := []struct {
		name       string
		amount     sdk.Coin
		targetAddr string
		marker     types.MarkerAccountI
		expectErr  string
	}{
		{
			name:   "successful increase",
			amount: sdk.NewCoin(hotdogMarker.Denom, sdkmath.NewInt(1000)),
			marker: hotdogMarker,
		},
		{
			name:       "successful increase with target address",
			amount:     sdk.NewCoin(hotdogMarker.Denom, sdkmath.NewInt(1000)),
			targetAddr: s.user1Addr.String(),
			marker:     hotdogMarker,
		},
		{
			name:       "invalid target address",
			amount:     sdk.NewCoin(hotdogMarker.Denom, sdkmath.NewInt(1000)),
			targetAddr: "invalidaddress",
			marker:     hotdogMarker,
		},
		{
			name:       "marker does not exist",
			amount:     sdk.NewCoin(invalidDenom, sdkmath.NewInt(1000)),
			targetAddr: s.user1Addr.String(),
			expectErr:  "invaliddenom marker does not exist",
		},
		{
			name:       "marker does not allow governance control",
			amount:     sdk.NewCoin(nonGovernanceMarker.GetDenom(), sdkmath.NewInt(1000)),
			targetAddr: s.user1Addr.String(),
			marker:     nonGovernanceMarker,
			expectErr:  "nonGovernanceMarker marker does not allow governance control",
		},
		{
			name:       "marker not in active status",
			amount:     sdk.NewCoin(nonActiveMarker.GetDenom(), sdkmath.NewInt(1000)),
			targetAddr: s.user1Addr.String(),
			marker:     nonActiveMarker,
			expectErr:  "cannot mint coin for a marker that is not in Active status",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.marker != nil {
				s.app.MarkerKeeper.SetMarker(s.ctx, tc.marker)
			}
			err := s.app.MarkerKeeper.HandleSupplyIncreaseProposal(s.ctx, tc.amount, tc.targetAddr)
			if len(tc.expectErr) > 0 {
				assert.Error(s.T(), err, "expected an error in test case: %s", tc.name)
				assert.Contains(s.T(), err.Error(), tc.expectErr, "unexpected error message in test case: %s", tc.name)
			} else {
				assert.NoError(s.T(), err, "did not expect an error in test case: %s", tc.name)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSupplyDecreaseProposal() {
	hotdogMarker := s.createTestMarker("hotdog")
	nonGovernanceMarker := s.createTestMarker("nonGovernanceMarker")
	invalidDenom := "invaliddenom"

	nonGovernanceMarker.AllowGovernanceControl = false
	s.app.MarkerKeeper.SetMarker(s.ctx, nonGovernanceMarker)

	tests := []struct {
		name      string
		amount    sdk.Coin
		marker    types.MarkerAccountI
		expectErr string
	}{
		{
			name:   "successful decrease",
			amount: sdk.NewCoin(hotdogMarker.Denom, sdkmath.NewInt(500)),
			marker: hotdogMarker,
		},
		{
			name:      "marker does not exist",
			amount:    sdk.NewCoin(invalidDenom, sdkmath.NewInt(500)),
			expectErr: "invaliddenom marker does not exist",
		},
		{
			name:      "marker does not allow governance control",
			amount:    sdk.NewCoin(nonGovernanceMarker.GetDenom(), sdkmath.NewInt(500)),
			marker:    nonGovernanceMarker,
			expectErr: "nonGovernanceMarker marker does not allow governance control",
		},
		{
			name:      "decrease amount more than supply",
			amount:    sdk.NewCoin(hotdogMarker.Denom, sdkmath.NewInt(2000)),
			marker:    hotdogMarker,
			expectErr: "cannot reduce marker total supply below zero hotdog, 2000",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.marker != nil {
				s.app.MarkerKeeper.SetMarker(s.ctx, tc.marker)
			}
			err := s.app.MarkerKeeper.HandleSupplyDecreaseProposal(s.ctx, tc.amount)
			if len(tc.expectErr) > 0 {
				assert.Error(s.T(), err, "expected an error in test case: %s", tc.name)
				assert.Contains(s.T(), err.Error(), tc.expectErr, "unexpected error message in test case: %s", tc.name)
			} else {
				assert.NoError(s.T(), err, "did not expect an error in test case: %s", tc.name)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSetAdministratorProposal() {
	hotdogMarker := s.createTestMarker("hotdog")
	nonGovernanceMarker := s.createTestMarker("nonGovernanceMarker")
	invalidDenom := "invaliddenom"

	nonGovernanceMarker.AllowGovernanceControl = false
	s.app.MarkerKeeper.SetMarker(s.ctx, nonGovernanceMarker)

	tests := []struct {
		name         string
		denom        string
		accessGrants []types.AccessGrant
		marker       types.MarkerAccountI
		expectErr    string
	}{
		{
			name:         "successful set administrator",
			denom:        hotdogMarker.Denom,
			accessGrants: []types.AccessGrant{{Address: s.user2Addr.String(), Permissions: types.AccessList{types.Access_Admin}}},
			marker:       hotdogMarker,
		},
		{
			name:         "marker does not exist",
			denom:        invalidDenom,
			accessGrants: []types.AccessGrant{{Address: s.user2Addr.String(), Permissions: types.AccessList{types.Access_Admin}}},
			expectErr:    "invaliddenom marker does not exist",
		},
		{
			name:         "marker does not allow governance control",
			denom:        nonGovernanceMarker.Denom,
			accessGrants: []types.AccessGrant{{Address: s.user2Addr.String(), Permissions: types.AccessList{types.Access_Admin}}},
			marker:       nonGovernanceMarker,
			expectErr:    "nonGovernanceMarker marker does not allow governance control",
		},
		{
			name:         "invalid access grants",
			denom:        hotdogMarker.Denom,
			accessGrants: []types.AccessGrant{{Address: s.user2Addr.String(), Permissions: types.AccessList{types.Access_Admin, types.Access_Admin}}},
			marker:       hotdogMarker,
			expectErr:    "access list contains duplicate entry",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.marker != nil {
				s.app.MarkerKeeper.SetMarker(s.ctx, tc.marker)
			}
			err := s.app.MarkerKeeper.HandleSetAdministratorProposal(s.ctx, tc.denom, tc.accessGrants)
			if len(tc.expectErr) > 0 {
				assert.Error(s.T(), err, "expected an error in test case: %s", tc.name)
				assert.Contains(s.T(), err.Error(), tc.expectErr, "unexpected error message in test case: %s", tc.name)
			} else {
				assert.NoError(s.T(), err, "did not expect an error in test case: %s", tc.name)
			}
		})
	}
}

func (s *KeeperTestSuite) TestRemoveAdministratorProposal() {
	hotdogMarker := s.createTestMarker("hotdog")
	nonGovernanceMarker := s.createTestMarker("nonGovernanceMarker")
	invalidDenom := "invaliddenom"

	nonGovernanceMarker.AllowGovernanceControl = false
	s.app.MarkerKeeper.SetMarker(s.ctx, nonGovernanceMarker)

	tests := []struct {
		name           string
		denom          string
		removedAddress []string
		marker         types.MarkerAccountI
		expectErr      string
	}{
		{
			name:           "successful remove administrator",
			denom:          hotdogMarker.Denom,
			removedAddress: []string{s.user1Addr.String()},
			marker:         hotdogMarker,
		},
		{
			name:           "marker does not exist",
			denom:          invalidDenom,
			removedAddress: []string{s.user1Addr.String()},
			expectErr:      "invaliddenom marker does not exist",
		},
		{
			name:           "marker does not allow governance control",
			denom:          nonGovernanceMarker.Denom,
			removedAddress: []string{s.user1Addr.String()},
			marker:         nonGovernanceMarker,
			expectErr:      "nonGovernanceMarker marker does not allow governance control",
		},
		{
			name:           "invalid address format",
			denom:          hotdogMarker.Denom,
			removedAddress: []string{"invalidaddress"},
			marker:         hotdogMarker,
			expectErr:      "decoding bech32 failed: invalid separator index -1",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.marker != nil {
				s.app.MarkerKeeper.SetMarker(s.ctx, tc.marker)
			}
			err := s.app.MarkerKeeper.HandleRemoveAdministratorProposal(s.ctx, tc.denom, tc.removedAddress)
			if len(tc.expectErr) > 0 {
				assert.Error(s.T(), err, "expected an error in test case: %s", tc.name)
				assert.Contains(s.T(), err.Error(), tc.expectErr, "unexpected error message in test case: %s", tc.name)
			} else {
				assert.NoError(s.T(), err, "did not expect an error in test case: %s", tc.name)
			}
		})
	}
}

func (s *KeeperTestSuite) TestChangeStatusProposal() {
	hotdogMarker := s.createTestMarker("hotdog")
	nonGovernanceMarker := s.createTestMarker("nonGovernanceMarker")
	invalidDenom := "invaliddenom"

	nonGovernanceMarker.AllowGovernanceControl = false
	s.app.MarkerKeeper.SetMarker(s.ctx, nonGovernanceMarker)

	pendingMarker := s.createTestMarker("pendingMarker")
	pendingMarker.SetStatus(types.StatusProposed)
	s.app.MarkerKeeper.SetMarker(s.ctx, pendingMarker)

	cancelledMarker := s.createTestMarker("cancelledMarker")
	cancelledMarker.SetStatus(types.StatusCancelled)
	s.app.MarkerKeeper.SetMarker(s.ctx, cancelledMarker)

	tests := []struct {
		name      string
		denom     string
		status    types.MarkerStatus
		marker    types.MarkerAccountI
		expectErr string
	}{
		{
			name:   "successful status change to active",
			denom:  hotdogMarker.Denom,
			status: types.StatusActive,
			marker: pendingMarker,
		},
		{
			name:   "successful status change to destroyed",
			denom:  cancelledMarker.Denom,
			status: types.StatusDestroyed,
			marker: cancelledMarker,
		},
		{
			name:      "marker does not exist",
			denom:     invalidDenom,
			status:    types.StatusActive,
			expectErr: "invaliddenom marker does not exist",
		},
		{
			name:      "marker does not allow governance control",
			denom:     nonGovernanceMarker.Denom,
			status:    types.StatusActive,
			marker:    nonGovernanceMarker,
			expectErr: "nonGovernanceMarker marker does not allow governance control",
		},
		{
			name:      "invalid marker status undefined",
			denom:     hotdogMarker.Denom,
			status:    types.StatusUndefined,
			marker:    hotdogMarker,
			expectErr: "error invalid marker status undefined",
		},
		{
			name:      "invalid status transition",
			denom:     cancelledMarker.Denom,
			status:    types.StatusProposed,
			marker:    cancelledMarker,
			expectErr: "invalid status transition proposed precedes existing status of cancelled",
		},
		{
			name:      "cannot delete non-cancelled marker",
			denom:     hotdogMarker.Denom,
			status:    types.StatusDestroyed,
			marker:    hotdogMarker,
			expectErr: "only cancelled markers can be deleted",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.marker != nil {
				s.app.MarkerKeeper.SetMarker(s.ctx, tc.marker)
			}
			err := s.app.MarkerKeeper.HandleChangeStatusProposal(s.ctx, tc.denom, tc.status)
			if len(tc.expectErr) > 0 {
				assert.Error(s.T(), err, "expected an error in test case: %s", tc.name)
				assert.Contains(s.T(), err.Error(), tc.expectErr, "unexpected error message in test case: %s", tc.name)
			} else {
				assert.NoError(s.T(), err, "did not expect an error in test case: %s", tc.name)
			}
		})
	}
}

func (s *KeeperTestSuite) TestWithdrawEscrowProposal() {
	hotdogMarker := s.createTestMarker("hotdog")
	nonGovernanceMarker := s.createTestMarker("nonGovernanceMarker")
	invalidDenom := "invaliddenom"

	nonGovernanceMarker.AllowGovernanceControl = false
	s.app.MarkerKeeper.SetMarker(s.ctx, nonGovernanceMarker)

	tests := []struct {
		name          string
		denom         string
		targetAddress string
		amount        sdk.Coins
		marker        types.MarkerAccountI
		expectErr     string
	}{
		{
			name:          "successful withdraw",
			denom:         hotdogMarker.Denom,
			targetAddress: s.user1Addr.String(),
			amount:        sdk.NewCoins(sdk.NewCoin(hotdogMarker.Denom, sdkmath.NewInt(100))),
			marker:        hotdogMarker,
			expectErr:     "",
		},
		{
			name:          "marker does not exist",
			denom:         invalidDenom,
			targetAddress: s.user1Addr.String(),
			expectErr:     "invaliddenom marker does not exist",
			amount:        sdk.NewCoins(sdk.NewCoin(invalidDenom, sdkmath.NewInt(100))),
		},
		{
			name:          "marker does not allow governance control",
			denom:         nonGovernanceMarker.Denom,
			targetAddress: s.user1Addr.String(),
			amount:        sdk.NewCoins(sdk.NewCoin(nonGovernanceMarker.Denom, sdkmath.NewInt(100))),
			marker:        nonGovernanceMarker,
			expectErr:     "nonGovernanceMarker marker does not allow governance control",
		},
		{
			name:          "invalid target address format",
			denom:         hotdogMarker.Denom,
			targetAddress: "invalidaddress",
			amount:        sdk.NewCoins(sdk.NewCoin(hotdogMarker.Denom, sdkmath.NewInt(100))),
			marker:        hotdogMarker,
			expectErr:     "decoding bech32 failed: invalid separator index -1",
		},
		{
			name:          "insufficient funds",
			denom:         hotdogMarker.Denom,
			targetAddress: s.user1Addr.String(),
			amount:        sdk.NewCoins(sdk.NewCoin(hotdogMarker.Denom, sdkmath.NewInt(2000))),
			marker:        hotdogMarker,
			expectErr:     "insufficient funds",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.marker != nil {
				s.app.MarkerKeeper.SetMarker(s.ctx, tc.marker)
			}
			err := s.app.MarkerKeeper.HandleWithdrawEscrowProposal(s.ctx, tc.denom, tc.targetAddress, tc.amount)
			if len(tc.expectErr) > 0 {
				assert.Error(s.T(), err, "expected an error in test case: %s", tc.name)
				assert.Contains(s.T(), err.Error(), tc.expectErr, "unexpected error message in test case: %s", tc.name)
			} else {
				assert.NoError(s.T(), err, "did not expect an error in test case: %s", tc.name)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSetDenomMetadataProposal() {
	hotdogMarker := s.createTestMarker("hotdog")
	nonGovernanceMarker := s.createTestMarker("nonGovernanceMarker")
	invalidDenom := "invaliddenom"

	nonGovernanceMarker.AllowGovernanceControl = false
	s.app.MarkerKeeper.SetMarker(s.ctx, nonGovernanceMarker)

	tests := []struct {
		name      string
		metadata  banktypes.Metadata
		marker    types.MarkerAccountI
		expectErr string
	}{
		{
			name: "successful set denom metadata",
			metadata: banktypes.Metadata{
				Description: "Hotdog token metadata",
				Base:        hotdogMarker.Denom,
				Display:     "Hotdog",
				Name:        "Hotdog Token",
				Symbol:      "HDG",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "hotdog", Exponent: 0},
				},
			},
			marker:    hotdogMarker,
			expectErr: "",
		},
		{
			name: "marker does not exist",
			metadata: banktypes.Metadata{
				Description: "Invalid token metadata",
				Base:        invalidDenom,
				Display:     "Invalid",
				Name:        "Invalid Token",
				Symbol:      "INV",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "invalid", Exponent: 0},
				},
			},
			expectErr: "invaliddenom marker does not exist",
		},
		{
			name: "marker does not allow governance control",
			metadata: banktypes.Metadata{
				Description: "Non-governance token metadata",
				Base:        nonGovernanceMarker.Denom,
				Display:     "NonGov",
				Name:        "Non-Gov Token",
				Symbol:      "NGT",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nonGovernanceMarker", Exponent: 0},
				},
			},
			marker:    nonGovernanceMarker,
			expectErr: "nonGovernanceMarker marker does not allow governance control",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.marker != nil {
				s.app.MarkerKeeper.SetMarker(s.ctx, tc.marker)
			}
			err := s.app.MarkerKeeper.HandleSetDenomMetadataProposal(s.ctx, tc.metadata)
			if len(tc.expectErr) > 0 {
				assert.Error(s.T(), err, "expected an error in test case: %s", tc.name)
				assert.Contains(s.T(), err.Error(), tc.expectErr, "unexpected error message in test case: %s", tc.name)
			} else {
				assert.NoError(s.T(), err, "did not expect an error in test case: %s", tc.name)
			}
		})
	}
}

func (s *KeeperTestSuite) createTestMarker(denom string) *types.MarkerAccount {
	marker := types.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(types.MustGetMarkerAddress(denom)),
		sdk.NewInt64Coin(denom, 1000),
		s.user1Addr,
		[]types.AccessGrant{
			{Address: s.user1Addr.String(), Permissions: types.AccessList{types.Access_Admin, types.Access_Mint}},
		},
		types.StatusProposed,
		types.MarkerType_Coin,
		true,
		true,
		false,
		[]string{},
	)
	s.Require().NoError(s.app.MarkerKeeper.AddSetNetAssetValues(s.ctx, marker, []types.NetAssetValue{types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1)}, types.ModuleName), "Failed to add navs to marker for tests")
	s.Require().NoError(s.app.MarkerKeeper.AddFinalizeAndActivateMarker(s.ctx, marker), "Failed to add marker for tests")
	return marker
}
