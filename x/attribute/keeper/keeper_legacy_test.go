package keeper_test

import (
	"fmt"
	"strings"
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/attribute/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
	"github.com/stretchr/testify/suite"
)

type KeeperLegacyTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress
}

func TestKeeperLegacyTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperLegacyTestSuite))
}

func (s *KeeperLegacyTestSuite) SetupTest() {
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
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("old.attribute", s.user1Addr, false))
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 3
	nameData.Params.MinSegmentLength = 3
	nameData.Params.MaxSegmentLength = 12

	app.NameKeeper.InitGenesis(ctx, nameData)

	params := app.AttributeKeeper.GetParams(ctx)
	params.MaxValueLength = 10
	app.AttributeKeeper.SetParams(ctx, params)
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	var attributeData types.GenesisState
	attributeData.Attributes = append(attributeData.Attributes, types.Attribute{
		Name:          "old.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	})

	InitGenesisLegacy(s.ctx, &attributeData, s.app)
	s.app.AttributeKeeper.ConvertLegacyAmino(s.ctx)
}

func (s *KeeperLegacyTestSuite) TestSetAttribute() {

	cases := map[string]struct {
		attr      types.Attribute
		accAddr   sdk.AccAddress
		ownerAddr sdk.AccAddress
		wantErr   bool
		errorMsg  string
	}{
		"should successfully add attribute": {
			attr: types.Attribute{
				Name:          "old.attribute",
				Value:         []byte("0123456789"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			accAddr:   s.user1Addr,
			ownerAddr: s.user1Addr,
			wantErr:   false,
			errorMsg:  "",
		},
		"should fail due to validate basic error": {
			attr: types.Attribute{
				Name:          "",
				Value:         []byte("01234567891"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			accAddr:   s.user1Addr,
			ownerAddr: s.user1Addr,
			wantErr:   true,
			errorMsg:  "invalid name: empty",
		},
		"should fail due to attribute length too long": {
			attr: types.Attribute{
				Name:          "name",
				Value:         []byte("01234567891"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			accAddr:   s.user1Addr,
			ownerAddr: s.user1Addr,
			wantErr:   true,
			errorMsg:  "attribute value length of 11 exceeds max length 10",
		},
		"should fail unable to find owner": {
			attr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("0123456789"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			accAddr:   s.user2Addr,
			ownerAddr: s.user2Addr,
			wantErr:   true,
			errorMsg:  fmt.Sprintf("no account found for owner address \"%s\"", s.user2),
		},
		"should fail unable to normalize segment length too short": {
			attr: types.Attribute{
				Name:          "example.cant.normalize.me",
				Value:         []byte("0123456789"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			accAddr:   s.user2Addr,
			ownerAddr: s.user2Addr,
			wantErr:   true,
			errorMsg:  "unable to normalize attribute name \"example.cant.normalize.me\": segment of name is too short",
		},
		"should fail unable to resolve name to user": {
			attr: types.Attribute{
				Name:          "example.not.found",
				Value:         []byte("0123456789"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			accAddr:   s.user1Addr,
			ownerAddr: s.user1Addr,
			wantErr:   true,
			errorMsg:  fmt.Sprintf("\"example.not.found\" does not resolve to address \"%s\"", s.user1),
		},
	}
	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := s.app.AttributeKeeper.SetAttribute(s.ctx, tc.accAddr, tc.attr, tc.ownerAddr)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}

}

func (s *KeeperLegacyTestSuite) TestDeleteAttribute() {

	cases := map[string]struct {
		name      string
		accAddr   sdk.AccAddress
		ownerAddr sdk.AccAddress
		wantErr   bool
		errorMsg  string
	}{
		"should fail to delete, cant resolve name wrong owner": {
			name:      "old.example.attribute",
			accAddr:   s.user1Addr,
			ownerAddr: s.user2Addr,
			wantErr:   true,
			errorMsg:  fmt.Sprintf("no account found for owner address \"%s\"", s.user2Addr),
		},
		"should fail to delete, cant resolve unknown name": {
			name:      "dne",
			accAddr:   s.user1Addr,
			ownerAddr: s.user1Addr,
			wantErr:   true,
			errorMsg:  fmt.Sprintf("\"dne\" does not resolve to address \"%s\"", s.user1Addr),
		},
		"should successfully delete attribute": {
			name:      "old.attribute",
			accAddr:   s.user1Addr,
			ownerAddr: s.user1Addr,
			wantErr:   false,
			errorMsg:  "",
		},
	}
	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := s.app.AttributeKeeper.DeleteAttribute(s.ctx, tc.accAddr, tc.name, nil, tc.ownerAddr)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}

}

func (s *KeeperLegacyTestSuite) TestGetAllAttributes() {

	attributes, err := s.app.AttributeKeeper.GetAllAttributes(s.ctx, s.user1Addr)
	s.NoError(err)
	s.Equal(1, len(attributes))

	attr := types.Attribute{
		Name:          "old.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	}

	attributes, err = s.app.AttributeKeeper.GetAllAttributes(s.ctx, s.user1Addr)
	s.NoError(err)
	s.Equal(attr.Name, attributes[0].Name)
	s.Equal(attr.Address, attributes[0].Address)
	s.Equal(attr.Value, attributes[0].Value)
}

func (s *KeeperLegacyTestSuite) TestGetAttributesByName() {

	attr := types.Attribute{
		Name:          "old.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	}

	_, err := s.app.AttributeKeeper.GetAttributes(s.ctx, s.user1Addr, "blah")
	s.Error(err)
	s.Equal("no address bound to name", err.Error())
	attributes, err := s.app.AttributeKeeper.GetAttributes(s.ctx, s.user1Addr, "old.attribute")
	s.NoError(err)
	s.Equal(1, len(attributes))
	s.Equal(attr.Name, attributes[0].Name)
	s.Equal(attr.Address, attributes[0].Address)
	s.Equal(attr.Value, attributes[0].Value)
}

func (s *KeeperLegacyTestSuite) TestInitGenesisAddingAttributes() {
	var attributeData types.GenesisState
	attributeData.Attributes = append(attributeData.Attributes, types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	})
	s.Assert().NotPanics(func() { s.app.AttributeKeeper.InitGenesis(s.ctx, &attributeData) })

	attributeData.Attributes = append(attributeData.Attributes, types.Attribute{
		Name:          "",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	})

	s.Assert().Panics(func() { s.app.AttributeKeeper.InitGenesis(s.ctx, &attributeData) })
}

func (s *KeeperLegacyTestSuite) TestIterateRecord() {
	s.Run("iterate attribute's", func() {
		attr := types.Attribute{
			Name:          "example.attribute",
			Value:         []byte("0123456789"),
			Address:       s.user1,
			AttributeType: types.AttributeType_String,
		}
		s.app.AttributeKeeper.SetAttribute(s.ctx, s.user1Addr, attr, s.user1Addr)

		records := []types.Attribute{}
		// Callback func that adds records to genesis state.
		appendToRecords := func(record types.Attribute) error {
			records = append(records, record)
			return nil
		}
		// Collect and return genesis state.
		err := s.app.AttributeKeeper.IterateRecords(s.ctx, types.AttributeKeyPrefixAmino, appendToRecords)
		s.Require().NoError(err)
		s.Require().Equal(1, len(records))
	})

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
	bz, err := types.ModuleCdc.MarshalBinaryBare(&attr)
	if err != nil {
		return err
	}
	key := accountAttributeKeyLegacy(acc, attr)
	store := ctx.KVStore(app.GetKey(types.ModuleName))
	store.Set(key, bz)
	return nil
}

// ONLY FOR TESTING.
func accountAttributeKeyLegacy(acc sdk.AccAddress, attr types.Attribute) []byte {
	key := append(types.AttributeKeyPrefixAmino, acc.Bytes()...)
	key = append(key, types.GetNameKeyBytes(attr.Name)...)
	return append(key, attr.Hash()...)
}
