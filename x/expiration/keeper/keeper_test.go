package keeper_test

import (
	"fmt"
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/expiration/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"

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

	moduleAssetId string
	owner         string
	blockHeight   int64
	deposit       sdk.Coin
	signers       []string

	validExpiration              types.Expiration
	emptyModuleAssetIdExpiration types.Expiration
	emptyOwnerAddressExpiration  types.Expiration
	minDepositNotMetExpiration   types.Expiration
}

func (s *KeeperTestSuite) SetupTest() {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 0)

	s.app = simapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{Time: tmtime.Now()})
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.ExpirationKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

	// set up users
	s.pubKey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubKey1.Address())
	s.user1 = s.user1Addr.String()
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	s.pubKey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubKey2.Address())
	s.user2 = s.user2Addr.String()

	s.pubKey3 = secp256k1.GenPrivKey().PubKey()
	s.user3Addr = sdk.AccAddress(s.pubKey3.Address())
	s.user3 = s.user3Addr.String()

	// setup up genesis
	var expirationData types.GenesisState
	expirationData.Params = types.DefaultParams()
	s.app.ExpirationKeeper.InitGenesis(s.ctx, &expirationData)

	// expiration tests
	s.moduleAssetId = "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"
	s.owner = s.user1
	s.blockHeight = s.ctx.BlockHeight() + 1
	s.deposit = types.DefaultDeposit
	s.signers = []string{s.user1}

	s.validExpiration = types.Expiration{
		ModuleAssetId: s.moduleAssetId,
		Owner:         s.owner,
		BlockHeight:   s.blockHeight,
		Deposit:       s.deposit,
	}
	s.emptyModuleAssetIdExpiration = types.Expiration{
		Owner:       s.owner,
		BlockHeight: s.blockHeight,
		Deposit:     s.deposit,
	}
	s.emptyOwnerAddressExpiration = types.Expiration{
		ModuleAssetId: s.moduleAssetId,
		BlockHeight:   s.blockHeight,
		Deposit:       s.deposit,
	}
	s.minDepositNotMetExpiration = types.Expiration{
		ModuleAssetId: s.moduleAssetId,
		Owner:         s.owner,
		BlockHeight:   s.blockHeight,
		Deposit:       sdk.NewInt64Coin(simapp.DefaultFeeDenom, 1),
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// TODO fix  >>>>>  panic: UnmarshalJSON expects a pointer
//func (s *KeeperTestSuite) TestParams() {
//	s.T().Run("param tests", func(t *testing.T) {
//		p := s.app.ExpirationKeeper.GetParams(s.ctx)
//		assert.NotNil(t, p)
//	})
//}

func (s *KeeperTestSuite) TestAddExpiration() {
	request := types.MsgAddExpirationRequest{}
	cases := []struct {
		name        string
		expiration  types.Expiration
		signers     []string
		msgTypeURL  string
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "should successfully add expiration",
			expiration:  s.validExpiration,
			signers:     s.signers,
			msgTypeURL:  request.MsgTypeURL(),
			wantErr:     false,
			expectedErr: "",
		},
		{
			name:        "should fail due to empty module asset id",
			expiration:  s.emptyModuleAssetIdExpiration,
			signers:     s.signers,
			msgTypeURL:  request.MsgTypeURL(),
			wantErr:     true,
			expectedErr: "invalid module asset id: empty address string is not allowed: invalid address",
		},
		{
			name:        "should fail to validate signers due to empty owner address",
			expiration:  s.emptyOwnerAddressExpiration,
			signers:     s.signers,
			msgTypeURL:  request.MsgTypeURL(),
			wantErr:     true,
			expectedErr: "invalid owner: empty address string is not allowed: invalid signers",
		},
		{
			name:        "should fail to validate signers due to empty signers",
			expiration:  s.validExpiration,
			signers:     nil,
			msgTypeURL:  request.MsgTypeURL(),
			wantErr:     true,
			expectedErr: fmt.Sprintf("intended signers [] do not match given signer [%s]: invalid signers", s.validExpiration.Owner),
		},
		{
			name:        "should fail to validate minimum required deposit",
			expiration:  s.minDepositNotMetExpiration,
			signers:     s.signers,
			msgTypeURL:  request.MsgTypeURL(),
			wantErr:     true,
			expectedErr: fmt.Sprintf("deposit amount %s is less than minimum deposit amount %s: invalid deposit amount", s.minDepositNotMetExpiration.Deposit.Amount, s.deposit.Amount),
		},
	}

	for _, tc := range cases {
		tc := tc

		s.T().Run(tc.name, func(t *testing.T) {
			err := s.app.ExpirationKeeper.ValidateSetExpiration(s.ctx, tc.expiration, tc.signers, tc.msgTypeURL)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
				if err := s.app.ExpirationKeeper.SetExpiration(s.ctx, tc.expiration); err != nil {
					assert.Fail(t, err.Error())
				}
			}
		})
	}
}
