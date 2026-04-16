package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/attribute/keeper"
	"github.com/provenance-io/provenance/x/attribute/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type MigrationTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	ownerPubKey cryptotypes.PubKey
	ownerAddr   sdk.AccAddress
	ownerBech32 string
}

func TestMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}

func (s *MigrationTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})

	s.ownerPubKey = secp256k1.GenPrivKey().PubKey()
	s.ownerAddr = sdk.AccAddress(s.ownerPubKey.Address())
	s.ownerBech32 = s.ownerAddr.String()

	// Bind the attribute name to the owner via the NameKeeper genesis,
	// same pattern used by KeeperTestSuite.SetupTest.
	var nameData nametypes.GenesisState
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("kyc", s.ownerAddr, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("provenance.kyc", s.ownerAddr, false))
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 3
	nameData.Params.MinSegmentLength = 3
	nameData.Params.MaxSegmentLength = 12
	s.app.NameKeeper.InitGenesis(s.ctx, nameData)

	// Materialize the owner account so SetAttribute's GetAccount check passes.
	s.app.AccountKeeper.SetAccount(
		s.ctx,
		s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.ownerAddr),
	)
}

// state written by the old binary must still be readable by the new one.
func (s *MigrationTestSuite) TestMigrate2to3_IsNoOp_AndPreservesState() {
	// 1) Seed state through the keeper.
	attr := types.Attribute{
		Name:          "provenance.kyc",
		Value:         []byte("verified"),
		AttributeType: types.AttributeType_String,
		Address:       s.ownerBech32,
	}
	s.Require().NoError(
		s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.ownerAddr),
		"SetAttribute must succeed during seeding",
	)

	pre, err := s.app.AttributeKeeper.GetAttributes(s.ctx, s.ownerBech32, attr.Name)
	s.Require().NoError(err)
	s.Require().Len(pre, 1, "expected one attribute before migration")

	// Run Migrate2to3.
	m := keeper.NewMigrator(s.app.AttributeKeeper)
	s.Require().NoError(m.Migrate2to3(s.ctx), "Migrate2to3 must not error")

	post, err := s.app.AttributeKeeper.GetAttributes(s.ctx, s.ownerBech32, attr.Name)
	s.Require().NoError(err)
	s.Require().Len(post, 1, "expected one attribute after migration")
	s.Require().Equal(pre[0], post[0], "attribute contents must be byte-identical before/after migration")
}

// Running Migrate2to3 on an empty store must be safe.
func (s *MigrationTestSuite) TestMigrate2to3_EmptyStore() {
	m := keeper.NewMigrator(s.app.AttributeKeeper)
	s.Require().NoError(m.Migrate2to3(s.ctx), "Migrate2to3 must handle empty state without error")
}
