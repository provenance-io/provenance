package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/marker/types"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
	s.app = simapp.Setup(s.T())
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
	hotdogMarker := types.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(types.MustGetMarkerAddress("hotdog")),
		sdk.NewInt64Coin("hotdog", 1000),
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
	s.Require().NoError(s.app.MarkerKeeper.AddSetNetAssetValues(s.ctx, hotdogMarker, []types.NetAssetValue{types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1)}, types.ModuleName), "Failed to add navs to 'hotdog' marker for tests")
	s.Require().NoError(s.app.MarkerKeeper.AddFinalizeAndActivateMarker(s.ctx, hotdogMarker), "Failed to add 'hotdog' marker for tests")
	tests := []struct {
		name       string
		amount     sdk.Coin
		targetAddr sdk.AccAddress
		expectErr  bool
	}{
		{
			name:       "successful increase",
			amount:     sdk.NewCoin(hotdogMarker.Denom, sdkmath.NewInt(1000)),
			targetAddr: s.user1Addr,
			expectErr:  false,
		},
		{
			name:       "invalid coin denomination",
			amount:     sdk.NewCoin("invalidcoin", sdkmath.NewInt(1000)),
			targetAddr: s.user1Addr,
			expectErr:  true,
		},
		{
			name:       "zero amount increase",
			amount:     sdk.NewCoin(hotdogMarker.Denom, sdkmath.ZeroInt()),
			targetAddr: s.user1Addr,
			expectErr:  true,
		},
		{
			name:       "negative amount",
			amount:     sdk.Coin{Denom: hotdogMarker.Denom, Amount: sdkmath.NewInt(-1000)},
			targetAddr: s.user1Addr,
			expectErr:  true,
		},
		{
			name:       "unauthorized user",
			amount:     sdk.NewCoin(hotdogMarker.Denom, sdkmath.NewInt(1000)),
			targetAddr: s.user2Addr,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.app.MarkerKeeper.HandleSupplyIncreaseProposal(s.ctx, tc.amount, tc.targetAddr.String())
			if tc.expectErr {
				assert.Error(s.T(), err, "expected an error in test case: %s", tc.name)
			} else {
				assert.NoError(s.T(), err, "did not expect an error in test case: %s", tc.name)
			}
		})
	}
}
