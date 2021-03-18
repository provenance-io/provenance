package keeper_test

import (
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
)

type MsgServerTestSuite struct {
	suite.Suite
	cfg     testnet.Config
	testnet *testnet.Network

	app         *app.App
	ctx         sdk.Context
	queryClient types.QueryClient
	msgServer   types.MsgServer

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress
}

func (s *MsgServerTestSuite) SetupTest() {
	app := simapp.Setup(false)
	s.app = app
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	s.ctx = ctx

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.MetadataKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	s.queryClient = queryClient
	msgServer := keeper.NewMsgServerImpl(app.MetadataKeeper)
	s.msgServer = msgServer

	// TODO: Add a msgServer client to properly pass context
	// for now use the handler_tests.go to test msg_server

}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

// TODO: MemorializeContract tests
// TODO: ChangeOwnership tests
// TODO: AddScope tests
// TODO: DeleteScope tests
// TODO: AddRecordGroup tests
// TODO: AddRecord tests
// TODO: AddScopeSpecification tests
// TODO: DeleteScopeSpecification tests
// TODO: AddContractSpecification tests
// TODO: DeleteContractSpecification tests
// TODO: AddRecordSpecification tests
// TODO: DeleteRecordSpecification tests
