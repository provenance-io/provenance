package metadata_test

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
	"github.com/provenance-io/provenance/x/metadata"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
	v039metadata "github.com/provenance-io/provenance/x/metadata/types/p8e"
)

type HandlerTestSuite struct {
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
	handler   sdk.Handler
}

func (s *HandlerTestSuite) SetupTest() {
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
	handler := metadata.NewHandler(app.MetadataKeeper)
	s.handler = handler
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (s HandlerTestSuite) TestAddContractSpecMsg() {
	validInputSpec := v039metadata.DefinitionSpec{
		Name: "perform_input_checks",
		ResourceLocation: &v039metadata.Location{Classname: "io.provenance.loan.LoanProtos$PartiesList",
			Ref: &v039metadata.ProvenanceReference{Hash: "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="},
		},
		Type: 1,
	}
	validOutputSpec := v039metadata.OutputSpec{Spec: &v039metadata.DefinitionSpec{
		Name: "additional_parties",
		ResourceLocation: &v039metadata.Location{
			Classname: "io.provenance.loan.LoanProtos$PartiesList",
			Ref: &v039metadata.ProvenanceReference{
				Hash: "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw==",
			},
		},
		Type: 1,
	},
	}
	validDefinition := v039metadata.DefinitionSpec{
		Name: "ExampleContract",
		ResourceLocation: &v039metadata.Location{Classname: "io.provenance.contracts.ExampleContract",
			Ref: &v039metadata.ProvenanceReference{Hash: "E36eeTUk8GYXGXjIbZTm4s/Dw3G1e42SinH1195t4ekgcXXPhfIpfQaEJ21PTzKhdv6JjhzQJ2kAJXK+TRXmeQ=="},
		},
		Type: 1,
	}
	validContractSpec := v039metadata.ContractSpec{ConsiderationSpecs: []*v039metadata.ConsiderationSpec{
		{FuncName: "additionalParties",
			InputSpecs:       []*v039metadata.DefinitionSpec{&validInputSpec},
			OutputSpec:       &validOutputSpec,
			ResponsibleParty: 1,
		},
	},
		Definition:      &validDefinition,
		InputSpecs:      []*v039metadata.DefinitionSpec{&validInputSpec},
		PartiesInvolved: []v039metadata.PartyType{v039metadata.PartyType_PARTY_TYPE_AFFILIATE},
	}
	_, err := s.handler(s.ctx, &types.MsgAddP8EContractSpecRequest{Contractspec: validContractSpec, Signers: []string{s.user1Addr.String()}})
	s.NoError(err)
}
