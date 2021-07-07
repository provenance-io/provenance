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

	v042 "github.com/provenance-io/provenance/x/metadata/legacy/v042"
	"github.com/provenance-io/provenance/x/metadata/types"
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

	osLocators []types.ObjectStoreLocator
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

	osLocators := []types.ObjectStoreLocator{
		types.NewOSLocatorRecord(s.user1Addr, "http://migration.test.user1.com"),
		types.NewOSLocatorRecord(s.user2Addr, "http://migration.test.user2.com"),
	}
	s.osLocators = osLocators
	var metadataData types.GenesisState
	metadataData.ObjectStoreLocators = append(metadataData.ObjectStoreLocators, osLocators...)
	err := InitGenesisLegacy(ctx, &metadataData, app)
	s.Require().NoError(err)
}

// InitGenesisLegacy sets up the key store with legacy key format (< v042)
func InitGenesisLegacy(ctx sdk.Context, data *types.GenesisState, app *app.App) error {
	for _, locator := range data.ObjectStoreLocators {
		accAddr, _ := sdk.AccAddressFromBech32(locator.Owner)
		key := v042.GetOSLocatorKeyLegacy(accAddr)

		bz, err := types.ModuleCdc.Marshal(&locator)
		if err != nil {
			return err
		}

		store := ctx.KVStore(app.GetKey(types.ModuleName))
		store.Set(key, bz)
	}
	return nil
}

func (s *MigrateTestSuite) TestMigrateTestSuite() {
	err := v042.MigrateOSLocatorKeys(s.ctx, s.app.GetKey("metadata"), types.ModuleCdc)
	s.Assert().NoError(err)
	store := s.ctx.KVStore(s.app.GetKey(types.ModuleName))
	for _, locator := range s.osLocators {
		// Should have removed object store locator at legacy key
		acc, _ := sdk.AccAddressFromBech32(locator.Owner)
		key := v042.GetOSLocatorKeyLegacy(acc)
		result := store.Get(key)
		s.Assert().Nil(result)

		// Should find object store locator from updated key
		key = types.GetOSLocatorKey(acc)
		result = store.Get(key)
		s.Assert().NotNil(result)
		var resultOSLocator types.ObjectStoreLocator
		err = types.ModuleCdc.Unmarshal(result, &resultOSLocator)
		s.Assert().NoError(err)
		s.Assert().Equal(locator, resultOSLocator)
	}
}
