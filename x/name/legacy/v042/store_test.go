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

	v042 "github.com/provenance-io/provenance/x/name/legacy/v042"
	"github.com/provenance-io/provenance/x/name/types"
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

	names []types.NameRecord
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
	s.user2 = s.user1Addr.String()

	s.names = []types.NameRecord{
		types.NewNameRecord("attribute", s.user1Addr, false),
		types.NewNameRecord("example.attribute", s.user1Addr, false),
		types.NewNameRecord("attribute2", s.user1Addr, false),
		types.NewNameRecord("example.attribute2", s.user1Addr, false),
	}

	var nameData types.GenesisState
	nameData.Bindings = append(nameData.Bindings, s.names...)
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 3
	nameData.Params.MinSegmentLength = 3
	nameData.Params.MaxSegmentLength = 12

	err := InitGenesisLegacy(ctx, &nameData, app)
	s.Require().NoError(err)
}

// InitGenesisLegacy creates the initial genesis state for the name module with legacy keys ( < v043)
func InitGenesisLegacy(ctx sdk.Context, data *types.GenesisState, app *app.App) error {
	for _, attr := range data.Bindings {
		key, err := types.GetNameKeyPrefix(attr.Name)
		if err != nil {
			return err
		}
		store := ctx.KVStore(app.GetKey(types.ModuleName))
		if err = attr.ValidateBasic(); err != nil {
			return err
		}
		bz, err := types.ModuleCdc.Marshal(&attr)
		if err != nil {
			return err
		}
		store.Set(key, bz)
		addr, err := sdk.AccAddressFromBech32(attr.Address)
		if err != nil {
			return err
		}

		addrPrefix, err := v042.GetAddressKeyPrefixLegacy(addr)
		if err != nil {
			return err
		}
		indexKey := append(addrPrefix, key...)
		store.Set(indexKey, bz)
	}
	return nil
}

func (s *MigrateTestSuite) TestMigrateTestSuite() {
	v042.MigrateAddresses(s.ctx, s.app.GetKey("name"))
	store := s.ctx.KVStore(s.app.GetKey(types.ModuleName))
	for _, name := range s.names {
		// Should have removed attribute at legacy key
		acc, _ := sdk.AccAddressFromBech32(name.Address)
		key, err := v042.GetAddressKeyPrefixLegacy(acc)
		s.Assert().NoError(err)
		nameKey, err := types.GetNameKeyPrefix(name.Name)
		s.Assert().NoError(err)
		result := store.Get(append(key, nameKey...))
		s.Assert().Nil(result)

		// Should find attribute with updated key
		key, err = types.GetAddressKeyPrefix(acc)
		s.Assert().NoError(err)
		newKey := append(key, nameKey...)
		result = store.Get(newKey)
		s.Assert().NotNil(result)
		var resultRecord types.NameRecord
		err = types.ModuleCdc.Unmarshal(result, &resultRecord)
		s.Assert().NoError(err)
		s.Assert().Equal(name, resultRecord, "address key record should equal new record")
	}
}
