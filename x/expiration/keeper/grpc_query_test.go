package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	types2 "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/expiration/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GrpcQueryTestSuite struct {
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
}

func (s *GrpcQueryTestSuite) SetupTest() {
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
	s.moduleAssetID = "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"
	s.time = s.ctx.BlockTime().AddDate(0, 0, 2)
	s.deposit = types.DefaultDeposit
	s.signers = []string{s.user1}
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(GrpcQueryTestSuite))
}

func anyMsg(owner string) types2.Any {
	scopeID := metadatatypes.ScopeMetadataAddress(uuid.New())
	msg := &metadatatypes.MsgDeleteScopeRequest{
		ScopeId: scopeID,
		Signers: []string{owner},
	}
	anyMsg, err := types2.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}
	return *anyMsg
}

func (s *GrpcQueryTestSuite) TestQueryExpiration() {
	moduleAssetID := s.moduleAssetID

	s.T().Run("add expiration for querying", func(t *testing.T) {
		expiration := *types.NewExpiration(moduleAssetID, s.user1, s.time, s.deposit, anyMsg(s.user1))
		assert.NoError(t, expiration.ValidateBasic(), "ValidateBasic: %s", "NewExpiration")
		err := s.app.ExpirationKeeper.SetExpiration(s.ctx, expiration)
		assert.NoError(t, err, "SetExpiration: %s", "NewExpiration")
	})

	s.T().Run("query expiration", func(t *testing.T) {
		req := types.QueryExpirationRequest{ModuleAssetId: s.moduleAssetID}
		res, err := s.queryClient.Expiration(context.Background(), &req)
		assert.NoError(t, err, "query: %s", "error")
		assert.NotNil(t, res, "query: %s", "response")
		assert.Equal(t, moduleAssetID, res.Expiration.ModuleAssetId, "query: %s", "expiration")
	})
}

func (s *GrpcQueryTestSuite) TestQueryAllExpirations() {
	expectedAll := 2

	s.T().Run("add expirations for querying", func(t *testing.T) {
		expiration1 := *types.NewExpiration(s.moduleAssetID, s.user1, s.time, s.deposit, anyMsg(s.user1))
		assert.NoError(t, expiration1.ValidateBasic(), "ValidateBasic: %s", "NewExpiration")
		err := s.app.ExpirationKeeper.SetExpiration(s.ctx, expiration1)
		assert.NoError(t, err, "SetExpiration: %s", "NewExpiration")

		expiration2 := *types.NewExpiration(s.user2, s.user3, s.time, s.deposit, anyMsg(s.user3))
		assert.NoError(t, expiration2.ValidateBasic(), "ValidateBasic: %s", "NewExpiration")
		err = s.app.ExpirationKeeper.SetExpiration(s.ctx, expiration2)
		assert.NoError(t, err, "SetExpiration: %s", "NewExpiration")
	})

	s.T().Run("query all expirations", func(t *testing.T) {
		req := types.QueryAllExpirationsRequest{}
		res, err := s.queryClient.AllExpirations(context.Background(), &req)
		assert.NoError(t, err, "query all: %s", "error")
		assert.NotNil(t, res, "query all: %s", "response")
		assert.Equal(t, expectedAll, len(res.Expirations), "query all: %s", "expirations")
	})
}

func (s *GrpcQueryTestSuite) TestQueryAllExpirationsByOwner() {
	moduleAssetID1 := s.user1
	moduleAssetID2 := s.user2
	moduleAssetID3 := s.user3

	sameOwner := s.user1
	diffOwner := s.user3

	expectedAll := 3
	expectedByOwner := 2
	expectedExpired := 0

	s.T().Run("add expirations for querying", func(t *testing.T) {
		expiration1 := *types.NewExpiration(moduleAssetID1, sameOwner, s.time, s.deposit, anyMsg(sameOwner))
		assert.NoError(t, expiration1.ValidateBasic(), "ValidateBasic: %s", "NewExpiration")
		err := s.app.ExpirationKeeper.SetExpiration(s.ctx, expiration1)
		assert.NoError(t, err, "SetExpiration: %s", "NewExpiration")

		expiration2 := *types.NewExpiration(moduleAssetID2, sameOwner, s.time, s.deposit, anyMsg(sameOwner))
		assert.NoError(t, expiration2.ValidateBasic(), "ValidateBasic: %s", "NewExpiration")
		err = s.app.ExpirationKeeper.SetExpiration(s.ctx, expiration2)
		assert.NoError(t, err, "SetExpiration: %s", "NewExpiration")

		expiration3 := *types.NewExpiration(moduleAssetID3, diffOwner, s.time, s.deposit, anyMsg(diffOwner))
		assert.NoError(t, expiration3.ValidateBasic(), "ValidateBasic: %s", "NewExpiration")
		err = s.app.ExpirationKeeper.SetExpiration(s.ctx, expiration3)
		assert.NoError(t, err, "SetExpiration: %s", "NewExpiration")
	})

	s.T().Run("query all expirations", func(t *testing.T) {
		req := types.QueryAllExpirationsRequest{}
		res, err := s.queryClient.AllExpirations(context.Background(), &req)
		assert.NoError(t, err, "query by owner: %s", "error")
		assert.NotNil(t, res, "query by owner: %s", "response")
		assert.Equal(t, expectedAll, len(res.Expirations), "query by owner: %s", "expirations")
	})

	// Query expirations by owner
	s.T().Run("query all expirations by owner", func(t *testing.T) {
		req := types.QueryAllExpirationsByOwnerRequest{Owner: sameOwner}
		res, err := s.queryClient.AllExpirationsByOwner(context.Background(), &req)
		assert.NoError(t, err, "query by owner: %s", "error")
		assert.NotNil(t, res, "query by owner: %s", "response")
		assert.Equal(t, expectedByOwner, len(res.Expirations), "query by owner: %s", "expirations")
	})

	// Query expired expirations
	s.T().Run("query all expired expirations", func(t *testing.T) {
		req := types.QueryAllExpiredExpirationsRequest{}
		res, err := s.queryClient.AllExpiredExpirations(context.Background(), &req)
		assert.NoError(t, err, "query expired: %s", "error")
		assert.NotNil(t, res, "query expired: %s", "response")
		assert.Equal(t, expectedExpired, len(res.Expirations), "query expired: %s", "expirations")
	})
}
