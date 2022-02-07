package v042_test

import (
	"testing"

	cryptotypes "github.com/tendermint/tendermint/crypto"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	v042 "github.com/provenance-io/provenance/x/attribute/legacy/v042"
	"github.com/provenance-io/provenance/x/attribute/types"
	namev042 "github.com/provenance-io/provenance/x/name/legacy/v042"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type MigrateTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	attributes []types.Attribute
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(MigrateTestSuite))
}

func (s *MigrateTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	s.app = app
	s.ctx = ctx

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	var nameData nametypes.GenesisState
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("attribute", s.user1Addr, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("example.attribute", s.user1Addr, false))
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 3
	nameData.Params.MinSegmentLength = 3
	nameData.Params.MaxSegmentLength = 12
	err := InitGenesisNameLegacy(ctx, &nameData, app)
	s.Require().NoError(err)

	var attrData types.GenesisState
	attrData.Params.MaxValueLength = 20
	attrData.Attributes = append(attrData.Attributes, types.Attribute{Address: s.user1, AttributeType: types.AttributeType_String, Value: []byte("first"), Name: "example.attribute"})
	attrData.Attributes = append(attrData.Attributes, types.Attribute{Address: s.user1, AttributeType: types.AttributeType_String, Value: []byte("second"), Name: "example.attribute"})
	attrData.Attributes = append(attrData.Attributes, types.Attribute{Address: s.user2, AttributeType: types.AttributeType_String, Value: []byte("third"), Name: "example.attribute"})
	s.attributes = attrData.Attributes
	err = InitGenesisLegacy(ctx, &attrData, app)
	s.Require().NoError(err)
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))
}

// InitGenesisLegacy creates the initial genesis state for the name module with legacy keys ( < v043)
func InitGenesisNameLegacy(ctx sdk.Context, data *nametypes.GenesisState, app *app.App) error {
	for _, name := range data.Bindings {
		key, err := nametypes.GetNameKeyPrefix(name.Name)
		if err != nil {
			return err
		}
		store := ctx.KVStore(app.GetKey(types.ModuleName))
		bz, err := types.ModuleCdc.Marshal(&name)
		if err != nil {
			return err
		}
		store.Set(key, bz)

		addr, err := sdk.AccAddressFromBech32(name.Address)
		if err != nil {
			return err
		}
		addrPrefix, err := namev042.GetAddressKeyPrefixLegacy(addr)
		if err != nil {
			return err
		}
		indexKey := append(addrPrefix, key...)
		store.Set(indexKey, bz)
	}
	return nil
}

// InitGenesisLegacy creates the initial genesis state for the attribute module with legacy keys ( < v043)
func InitGenesisLegacy(ctx sdk.Context, data *types.GenesisState, app *app.App) error {
	if err := data.ValidateBasic(); err != nil {
		panic(err)
	}
	for _, attr := range data.Attributes {
		acc, err := sdk.AccAddressFromBech32(attr.Address)
		if err != nil {
			return err
		}
		bz, err := types.ModuleCdc.Marshal(&attr)
		if err != nil {
			return err
		}
		key := v042.AccountAttributeKeyLegacy(acc, attr)
		store := ctx.KVStore(app.GetKey(types.ModuleName))
		store.Set(key, bz)
	}
	return nil
}

func (s *MigrateTestSuite) TestMigrateTestSuite() {
	v042.MigrateAddressLength(s.ctx, s.app.GetKey("attribute"))
	store := s.ctx.KVStore(s.app.GetKey(types.ModuleName))
	for _, attr := range s.attributes {
		// Should have removed attribute at legacy key
		acc, _ := sdk.AccAddressFromBech32(attr.Address)
		key := v042.AccountAttributeKeyLegacy(acc, attr)
		result := store.Get(key)
		s.Assert().Nil(result)

		// Should find attribute with updated key
		key = types.AddrAttributeKey(acc, attr)
		result = store.Get(key)
		s.Assert().NotNil(result)
		var resultAttr types.Attribute
		err := types.ModuleCdc.Unmarshal(result, &resultAttr)
		s.Assert().NoError(err)
		s.Assert().Equal(attr, resultAttr)
	}
}
