package v042_test

import (
	"fmt"
	"strings"
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

	var attrData types.GenesisState
	attrData.Params.MaxValueLength = 20
	attrData.Attributes = append(attrData.Attributes, types.Attribute{Address: s.user1, AttributeType: types.AttributeType_String, Value: []byte("first"), Name: "example.attribute"})
	attrData.Attributes = append(attrData.Attributes, types.Attribute{Address: s.user1, AttributeType: types.AttributeType_String, Value: []byte("second"), Name: "example.attribute"})
	attrData.Attributes = append(attrData.Attributes, types.Attribute{Address: s.user2, AttributeType: types.AttributeType_String, Value: []byte("third"), Name: "example.attribute"})
	s.attributes = attrData.Attributes
	InitGenesisLegacy(ctx, &attrData, app)
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))
}

// InitGenesisLegacy creates the initial genesis state for the name module. ONLY FOR TESTING.
func InitGenesisNameLegacy(ctx sdk.Context, data *nametypes.GenesisState, app *app.App) {
	for _, attr := range data.Bindings {
		if err := importNameLegacy(ctx, attr, app); err != nil {
			panic(err)
		}
	}
}

// A genesis helper that imports name state without owner checks.ONLY FOR TESTING.
func importNameLegacy(ctx sdk.Context, record nametypes.NameRecord, app *app.App) error {
	key, err := nametypes.GetNameKeyPrefix(record.Name)
	if err != nil {
		return err
	}
	store := ctx.KVStore(app.GetKey(types.ModuleName))
	if store.Has(key) {
		return nametypes.ErrNameAlreadyBound
	}
	if err = record.ValidateBasic(); err != nil {
		return err
	}
	bz, err := types.ModuleCdc.Marshal(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	addr, err := sdk.AccAddressFromBech32(record.Address)
	if err != nil {
		return err
	}
	// Now index by address
	addrPrefix, err := namev042.GetAddressKeyPrefixLegacy(addr)
	if err != nil {
		return err
	}
	indexKey := append(addrPrefix, key...) // [0x04] :: [addr-bytes] :: [name-key-bytes]
	store.Set(indexKey, bz)
	return nil
}

// InitGenesisLegacy creates the initial genesis state for the attribute module. ONLY FOR TESTING.
func InitGenesisLegacy(ctx sdk.Context, data *types.GenesisState, app *app.App) {
	if err := data.ValidateBasic(); err != nil {
		panic(err)
	}
	for _, attr := range data.Attributes {
		if err := importAttributeLegacy(ctx, attr, app); err != nil {
			panic(err)
		}
	}
}

// A genesis helper that imports attribute state without owner checks.ONLY FOR TESTING.
func importAttributeLegacy(ctx sdk.Context, attr types.Attribute, app *app.App) error {
	// Ensure attribute is valid
	if err := attr.ValidateBasic(); err != nil {
		return err
	}
	// Attribute must have a valid, non-empty address to import
	if strings.TrimSpace(attr.Address) == "" {
		return fmt.Errorf("unable to import attribute with empty address")
	}
	acc, err := sdk.AccAddressFromBech32(attr.Address)
	if err != nil {
		return err
	}
	// Ensure name is stored in normalized format.
	if attr.Name, err = app.NameKeeper.Normalize(ctx, attr.Name); err != nil {
		return fmt.Errorf("unable to normalize attribute name \"%s\": %w", attr.Name, err)
	}
	// Store the sanitized account attribute
	bz, err := types.ModuleCdc.Marshal(&attr)
	if err != nil {
		return err
	}
	key := v042.AccountAttributeKeyLegacy(acc, attr)
	store := ctx.KVStore(app.GetKey(types.ModuleName))
	store.Set(key, bz)
	return nil
}

func (s *MigrateTestSuite) TestMigrateTestSuite() {
	v042.MigrateAddressLength(s.ctx, s.app.GetKey("attribute"), types.ModuleCdc)
	store := s.ctx.KVStore(s.app.GetKey(types.ModuleName))
	for _, attr := range s.attributes {
		// Should have removed attribute at legacy key
		acc, _ := sdk.AccAddressFromBech32(attr.Address)
		key := v042.AccountAttributeKeyLegacy(acc, attr)
		result := store.Get(key)
		s.Assert().Nil(result)

		// Should find attribute with updated key
		newAddr := v042.ConvertLegacyAddress(acc)
		key = types.AccountAttributeKey(newAddr, attr)
		result = store.Get(key)
		s.Assert().NotNil(result)
		var resultAttr types.Attribute
		err := types.ModuleCdc.Unmarshal(result, &resultAttr)
		attr.Address = newAddr.String()
		s.Assert().NoError(err)
		s.Assert().Equal(attr, resultAttr)
	}
}

// func (s *MigrateTestSuite) TestConvertLegacyAddress() {
// 	padding := make([]byte, 12)
// 	acc, _ := sdk.AccAddressFromBech32(s.attributes[0].Address)
// 	convertedLegacyAddr := v042.ConvertLegacyAddress(acc)
// 	s.Assert().Equal(32, len(convertedLegacyAddr))
// 	s.Assert().Equal(acc.Bytes()[:20], convertedLegacyAddr.Bytes()[:20])
// 	s.Assert().Equal(padding, convertedLegacyAddr.Bytes()[20:])
// }
