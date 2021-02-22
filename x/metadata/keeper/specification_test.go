package keeper

import (
	"testing"

	"github.com/google/uuid"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/x/metadata/types"
)

type ScopeSpecKeeperTestSuite struct {
	suite.Suite

	app         *app.App
	ctx         sdk.Context
	queryClient types.QueryClient

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	specUUID uuid.UUID
	specID   types.MetadataAddress
}

func (s *ScopeSpecKeeperTestSuite) SetupTest() {
	testApp := simapp.Setup(false)
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.specUUID = uuid.New()
	s.specID = types.ScopeSpecMetadataAddress(s.specUUID)

	s.app = testApp
	s.ctx = ctx

	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, testApp.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, testApp.MetadataKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	s.queryClient = queryClient
}

func (s *ScopeSpecKeeperTestSuite) TestIterateScopeSpecs() {
	// TODO: implement
}

func (s *ScopeSpecKeeperTestSuite) TestIterateScopeSpecsForAddress() {
	// TODO: implement
}

func (s *ScopeSpecKeeperTestSuite) TestIterateScopeSpecsForContractSpec() {
	// TODO: implement
}

func (s *ScopeSpecKeeperTestSuite) TestGetScopeSpecification() {
	// TODO: implement
}

func (s *ScopeSpecKeeperTestSuite) TestSetScopeSpecification() {
	// TODO: implement
}

func (s *ScopeSpecKeeperTestSuite) TestDeleteScopeSpecification() {
	// TODO: implement
}

func (s *ScopeSpecKeeperTestSuite) TestValidateScopeSpecUpdate() {
	// TODO: implement
}

func (s *ScopeSpecKeeperTestSuite) TestValidateScopeSpecAllOwnersAreSigners() {
	// TODO: implement
}

func TestScopeSpecKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(ScopeSpecKeeperTestSuite))
}
