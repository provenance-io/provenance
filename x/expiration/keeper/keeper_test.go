package keeper_test

import (
	"fmt"
	"testing"
	"time"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosauthtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/expiration/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *simapp.App
	ctx         sdk.Context
	queryClient types.QueryClient

	pubKey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubKey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	pubKey3   cryptotypes.PubKey
	user3     string
	user3Addr sdk.AccAddress

	moduleAssetID string
	time          time.Time
	deposit       sdk.Coin
	signers       []string

	validExpiration              types.Expiration
	emptyModuleAssetIdExpiration types.Expiration
	emptyOwnerAddressExpiration  types.Expiration
	minDepositNotMetExpiration   types.Expiration
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{Time: tmtime.Now()})
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.ExpirationKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

	// set up users
	s.pubKey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubKey1.Address())
	s.user1 = s.user1Addr.String()
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, s.user1Addr, sdk.NewCoins(types.DefaultDeposit).Add(types.DefaultDeposit)), "funding account")

	s.pubKey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubKey2.Address())
	s.user2 = s.user2Addr.String()

	s.pubKey3 = secp256k1.GenPrivKey().PubKey()
	s.user3Addr = sdk.AccAddress(s.pubKey3.Address())
	s.user3 = s.user3Addr.String()
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, s.user3Addr, sdk.NewCoins(types.DefaultDeposit)), "funding account")

	// setup up genesis
	var expirationData types.GenesisState
	expirationData.Params = types.DefaultParams()
	s.app.ExpirationKeeper.InitGenesis(s.ctx, &expirationData)

	// set default deposit amount
	types.DefaultDeposit = sdk.NewInt64Coin("nhash", 1000000000)

	// expiration tests
	s.moduleAssetID = "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"
	s.time = s.ctx.BlockTime().AddDate(0, 0, 2)
	s.deposit = types.DefaultDeposit
	s.signers = []string{s.user1}

	s.validExpiration = types.Expiration{
		ModuleAssetId: s.moduleAssetID,
		Owner:         s.user1,
		Time:          s.time,
		Deposit:       s.deposit,
	}
	s.emptyModuleAssetIdExpiration = types.Expiration{
		Owner:   s.user1,
		Time:    s.time,
		Deposit: s.deposit,
	}
	s.emptyOwnerAddressExpiration = types.Expiration{
		ModuleAssetId: s.moduleAssetID,
		Time:          s.time,
		Deposit:       s.deposit,
	}
	s.minDepositNotMetExpiration = types.Expiration{
		ModuleAssetId: s.moduleAssetID,
		Owner:         s.user1,
		Time:          s.time,
		Deposit:       sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1),
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestParams() {
	s.T().Run("param tests", func(t *testing.T) {
		p := s.app.ExpirationKeeper.GetParams(s.ctx)
		assert.NotNil(t, p)
	})
}

func (s *KeeperTestSuite) TestModuleAccount() {
	s.T().Run("module account check", func(t *testing.T) {
		gov := s.app.ExpirationKeeper.GetModuleAccount(s.ctx)
		assert.NotNil(t, gov)
	})
}

func (s *KeeperTestSuite) TestAddExpiration() {
	request := types.MsgAddExpirationRequest{}
	cases := []struct {
		name        string
		expiration  types.Expiration
		signers     []string
		msgTypeURL  string
		granter     sdk.AccAddress
		grantee     sdk.AccAddress
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "should fail to validate due to empty module asset id",
			expiration:  s.emptyModuleAssetIdExpiration,
			signers:     s.signers,
			msgTypeURL:  request.MsgTypeURL(),
			granter:     nil,
			grantee:     nil,
			wantErr:     true,
			expectedErr: "invalid module asset id: empty address string is not allowed: invalid address",
		},
		{
			name:        "should fail to validate signers due to empty owner address",
			expiration:  s.emptyOwnerAddressExpiration,
			signers:     s.signers,
			msgTypeURL:  request.MsgTypeURL(),
			granter:     nil,
			grantee:     nil,
			wantErr:     true,
			expectedErr: "invalid owner: empty address string is not allowed: invalid signers",
		},
		{
			name:        "should fail to validate signers due to empty signers",
			expiration:  s.validExpiration,
			signers:     nil,
			msgTypeURL:  request.MsgTypeURL(),
			granter:     nil,
			grantee:     nil,
			wantErr:     true,
			expectedErr: fmt.Sprintf("intended signers [] do not match given signer [%s]: invalid signers", s.validExpiration.Owner),
		},
		{
			name:        "should fail to validate minimum required deposit",
			expiration:  s.minDepositNotMetExpiration,
			signers:     s.signers,
			msgTypeURL:  request.MsgTypeURL(),
			granter:     nil,
			grantee:     nil,
			wantErr:     true,
			expectedErr: fmt.Sprintf("deposit amount %s is less than minimum deposit amount %s: invalid deposit amount", s.minDepositNotMetExpiration.Deposit.Amount, s.deposit.Amount),
		},
		{
			name:        "should fail to validate with authz",
			expiration:  s.validExpiration,
			signers:     []string{s.user2},
			msgTypeURL:  request.MsgTypeURL(),
			granter:     nil,
			grantee:     nil,
			wantErr:     true,
			expectedErr: fmt.Sprintf("intended signers [%s] do not match given signer [%s]: invalid signers", s.user2, s.validExpiration.Owner),
		},
		{
			name:        "should successfully add expiration",
			expiration:  s.validExpiration,
			signers:     s.signers,
			msgTypeURL:  request.MsgTypeURL(),
			granter:     nil,
			grantee:     nil,
			wantErr:     false,
			expectedErr: "",
		},
		{
			name:        "should successfully add expiration with authz",
			expiration:  s.validExpiration,
			signers:     []string{s.user3},
			msgTypeURL:  request.MsgTypeURL(),
			granter:     s.user1Addr, // user1 is the owner in s.expiration
			grantee:     s.user3Addr,
			wantErr:     false,
			expectedErr: "",
		},
	}

	now := s.ctx.BlockHeader().Time
	s.Assert().NotNil(now)

	for _, tc := range cases {
		tc := tc

		s.T().Run(tc.name, func(t *testing.T) {
			createAuth := tc.grantee != nil && tc.granter != nil
			if createAuth {
				a := authz.NewGenericAuthorization(tc.msgTypeURL)
				onehour := now.Add(time.Hour)
				err := s.app.AuthzKeeper.SaveGrant(s.ctx, tc.grantee, tc.granter, a, &onehour)
				s.Assert().NoError(err)
			}

			err := s.app.ExpirationKeeper.ValidateSetExpiration(s.ctx, tc.expiration, tc.signers, tc.msgTypeURL)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr, err.Error(), "%s error", tc.name)
			} else {
				assert.NoError(t, err, "%s unexpected error", tc.name)
				if !tc.expiration.Deposit.IsZero() {
					priv, _, _ := testdata.KeyTestPubAddr()
					addr, _ := sdk.AccAddressFromBech32(tc.expiration.Owner)
					acct := cosmosauthtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
					err := testutil.FundAccount(s.app.BankKeeper, s.ctx, acct.GetAddress(), sdk.NewCoins(tc.expiration.Deposit))
					s.Require().NoError(err, fmt.Sprintf("%s: fund owner account", tc.name))
				}
				if err := s.app.ExpirationKeeper.SetExpiration(s.ctx, tc.expiration); err != nil {
					assert.Fail(t, err.Error())
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestExtendExpiration() {
	request := types.MsgExtendExpirationRequest{}
	cases := []struct {
		name        string
		expiration  types.Expiration
		signers     []string
		msgTypeURL  string
		granter     sdk.AccAddress
		grantee     sdk.AccAddress
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "should fail to validate due to empty module asset id",
			expiration:  s.emptyModuleAssetIdExpiration,
			signers:     s.signers,
			msgTypeURL:  request.MsgTypeURL(),
			granter:     nil,
			grantee:     nil,
			wantErr:     true,
			expectedErr: "invalid module asset id: empty address string is not allowed: invalid address",
		},
		{
			name:        "should fail to validate signers due to empty owner address",
			expiration:  s.emptyOwnerAddressExpiration,
			signers:     s.signers,
			msgTypeURL:  request.MsgTypeURL(),
			granter:     nil,
			grantee:     nil,
			wantErr:     true,
			expectedErr: "invalid owner: empty address string is not allowed: invalid signers",
		},
		{
			name:        "should fail to validate signers due to empty signers",
			expiration:  s.validExpiration,
			signers:     nil,
			msgTypeURL:  request.MsgTypeURL(),
			granter:     nil,
			grantee:     nil,
			wantErr:     true,
			expectedErr: fmt.Sprintf("intended signers [] do not match given signer [%s]: invalid signers", s.validExpiration.Owner),
		},
		{
			name:        "should fail to validate minimum required deposit",
			expiration:  s.minDepositNotMetExpiration,
			signers:     s.signers,
			msgTypeURL:  request.MsgTypeURL(),
			granter:     nil,
			grantee:     nil,
			wantErr:     true,
			expectedErr: fmt.Sprintf("deposit amount %s is less than minimum deposit amount %s: invalid deposit amount", s.minDepositNotMetExpiration.Deposit.Amount, s.deposit.Amount),
		},
		{
			name:        "should fail to validate with authz",
			expiration:  s.validExpiration,
			signers:     []string{s.user2},
			msgTypeURL:  request.MsgTypeURL(),
			granter:     nil,
			grantee:     nil,
			wantErr:     true,
			expectedErr: fmt.Sprintf("intended signers [%s] do not match given signer [%s]: invalid signers", s.user2, s.validExpiration.Owner),
		},
		{
			name:        "should successfully extend expiration",
			expiration:  s.validExpiration,
			signers:     s.signers,
			msgTypeURL:  request.MsgTypeURL(),
			granter:     nil,
			grantee:     nil,
			wantErr:     false,
			expectedErr: "",
		},
		{
			name:        "should successfully extend expiration with authz",
			expiration:  s.validExpiration,
			signers:     []string{s.user3},
			msgTypeURL:  request.MsgTypeURL(),
			granter:     s.user1Addr, // user1 is the owner in s.expiration
			grantee:     s.user3Addr,
			wantErr:     false,
			expectedErr: "",
		},
	}

	now := s.ctx.BlockHeader().Time
	s.Assert().NotNil(now)

	for _, tc := range cases {
		tc := tc

		s.T().Run(tc.name, func(t *testing.T) {
			createAuth := tc.grantee != nil && tc.granter != nil
			if createAuth {
				a := authz.NewGenericAuthorization(tc.msgTypeURL)
				onehour := now.Add(time.Hour)
				err := s.app.AuthzKeeper.SaveGrant(s.ctx, tc.grantee, tc.granter, a, &onehour)
				s.Assert().NoError(err)
			}

			err := s.app.ExpirationKeeper.ValidateSetExpiration(s.ctx, tc.expiration, tc.signers, tc.msgTypeURL)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr, err.Error(), "%s error", tc.name)
			} else {
				assert.NoError(t, err, "%s unexpected error", tc.name)
				if err := s.app.ExpirationKeeper.SetExpiration(s.ctx, tc.expiration); err != nil {
					assert.Fail(t, err.Error())
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestRemoveExpiration() {
	cases := []struct {
		name          string
		moduleAssetId string
		addRemove     bool
		wantErr       bool
		expectedErr   string
	}{
		{
			name:          "should fail to validate due to empty module asset id",
			moduleAssetId: "",
			addRemove:     false,
			wantErr:       true,
			expectedErr:   "empty address string is not allowed: invalid key: failed to invoke expiration",
		},
		{
			name:          "should succeed on no record found",
			moduleAssetId: s.user2,
			addRemove:     false,
			wantErr:       false,
			expectedErr:   "",
		},
		{
			name:          "should successfully remove expiration",
			moduleAssetId: s.validExpiration.ModuleAssetId,
			addRemove:     true,
			wantErr:       false,
			expectedErr:   "",
		},
	}

	now := s.ctx.BlockHeader().Time
	s.Assert().NotNil(now)

	for _, tc := range cases {
		tc := tc

		s.T().Run(tc.name, func(t *testing.T) {
			if tc.addRemove {
				// fund account
				priv, _, _ := testdata.KeyTestPubAddr()
				addr, _ := sdk.AccAddressFromBech32(s.validExpiration.Owner)
				acct := cosmosauthtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
				err := testutil.FundAccount(s.app.BankKeeper, s.ctx, acct.GetAddress(), sdk.NewCoins(s.validExpiration.Deposit))
				s.Require().NoError(err, fmt.Sprintf("%s: fund owner account", tc.name))

				// add expiration
				if err := s.app.ExpirationKeeper.SetExpiration(s.ctx, s.validExpiration); err != nil {
					assert.Fail(t, err.Error())
				}

				// validate expiration exists
				if exp, err := s.app.ExpirationKeeper.GetExpiration(s.ctx, tc.moduleAssetId); exp == nil {
					assert.Fail(t, err.Error(), "%s unexpected error", tc.name)
				}
			}

			refundTo, err := sdk.AccAddressFromBech32(s.validExpiration.Owner)
			assert.NoError(t, err, "%s invalid address", tc.name)
			err = s.app.ExpirationKeeper.RemoveExpiration(s.ctx, tc.moduleAssetId, refundTo)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr, err.Error(), "%s error", tc.name)
			} else {
				assert.NoError(t, err, "%s unexpected error", tc.name)
				if tc.addRemove {
					// validate expiration doesn't exist
					exp, err := s.app.ExpirationKeeper.GetExpiration(s.ctx, tc.moduleAssetId)
					assert.Empty(t, exp, "%s unexpected error", tc.name)
					assert.NoError(t, err, "%s unexpected error", tc.name)
				}
			}
		})
	}
}
