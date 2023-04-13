package keeper_test

import (
	"sort"
	"testing"
	"time"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/metadata/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *simapp.App
	ctx         sdk.Context
	queryClient types.QueryClient

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	pubkey3   cryptotypes.PubKey
	user3     string
	user3Addr sdk.AccAddress

	pubkey4   cryptotypes.PubKey
	user4     string
	user4Addr sdk.AccAddress

	objectLocator metadatatypes.ObjectStoreLocator
	ownerAddr     sdk.AccAddress
	encryptionKey sdk.AccAddress
	uri           string

	objectLocator1 metadatatypes.ObjectStoreLocator
	ownerAddr1     sdk.AccAddress
	encryptionKey1 sdk.AccAddress
	uri1           string
}

func (s *KeeperTestSuite) SetupTest() {
	pioconfig.SetProvenanceConfig("atom", 0)
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.MetadataKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.pubkey3 = secp256k1.GenPrivKey().PubKey()
	s.user3Addr = sdk.AccAddress(s.pubkey3.Address())
	s.user3 = s.user3Addr.String()

	s.pubkey4 = secp256k1.GenPrivKey().PubKey()
	s.user4Addr = sdk.AccAddress(s.pubkey4.Address())
	s.user4 = s.user4Addr.String()

	// add os locator
	s.ownerAddr = s.user1Addr
	s.uri = "http://foo.com"
	s.encryptionKey = sdk.AccAddress{}
	s.objectLocator = metadatatypes.NewOSLocatorRecord(s.ownerAddr, s.encryptionKey, s.uri)

	s.ownerAddr1 = s.user2Addr
	s.uri1 = "http://bar.com"
	s.encryptionKey1 = sdk.AccAddress(s.pubkey1.Address())
	s.objectLocator1 = metadatatypes.NewOSLocatorRecord(s.ownerAddr1, s.encryptionKey1, s.uri1)
	//set up genesis
	var metadataData metadatatypes.GenesisState
	metadataData.Params = metadatatypes.DefaultParams()
	metadataData.OSLocatorParams = metadatatypes.DefaultOSLocatorParams()
	metadataData.ObjectStoreLocators = append(metadataData.ObjectStoreLocators, s.objectLocator, s.objectLocator1)
	s.app.MetadataKeeper.InitGenesis(s.ctx, &metadataData)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// ownerPartyList returns a party with role OWNER for each address provided.
// This func is used in other keeper test files.
func ownerPartyList(addresses ...string) []types.Party {
	retval := make([]types.Party, len(addresses))
	for i, addr := range addresses {
		retval[i] = types.Party{Address: addr, Role: types.PartyType_PARTY_TYPE_OWNER}
	}
	return retval
}

func (s *KeeperTestSuite) TestParams() {
	s.T().Run("param tests", func(t *testing.T) {
		p := s.app.MetadataKeeper.GetParams(s.ctx)
		assert.NotNil(t, p)

		osp := s.app.MetadataKeeper.GetOSLocatorParams(s.ctx)
		assert.NotNil(t, osp)
		assert.Equal(t, osp.MaxUriLength, s.app.MetadataKeeper.GetMaxURILength(s.ctx))
	})
}

func (s *KeeperTestSuite) TestGetOSLocator() {
	s.Run("get os locator by owner address", func() {
		r, found := s.app.MetadataKeeper.GetOsLocatorRecord(s.ctx, s.user1Addr)
		s.Require().NotEmpty(r)
		s.Require().True(found)
	})
	s.Run("not found by owner address", func() {
		r, found := s.app.MetadataKeeper.GetOsLocatorRecord(s.ctx, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))
		s.Require().Empty(r)
		s.Require().False(found)
	})
}

func (s *KeeperTestSuite) TestAddOSLocator() {
	s.Run("add os locator", func() {
		// create account and check default values
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user3Addr)
		s.Require().NotNil(acc)
		s.Require().Equal(s.user3Addr, acc.GetAddress())
		s.Require().EqualValues(nil, acc.GetPubKey())
		s.Require().EqualValues(0, acc.GetSequence())
		// set and get the new account.
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		acc1 := s.app.AccountKeeper.GetAccount(s.ctx, s.user3Addr)
		s.Require().NotNil(acc1)
		// create os locator with ^^ account
		err := s.app.MetadataKeeper.SetOSLocator(s.ctx, s.user3Addr, sdk.AccAddress{}, "https://bob.com/alice")
		s.Require().Empty(err)
		r, found := s.app.MetadataKeeper.GetOsLocatorRecord(s.ctx, s.user1Addr)
		s.Require().NotEmpty(r)
		s.Require().True(found)
	})

	s.Run("add os locator account does not exist.", func() {
		// create account and check default values
		err := s.app.MetadataKeeper.SetOSLocator(s.ctx, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()), sdk.AccAddress{}, "https://bob.com/alice")
		s.Require().NotEmpty(err)
	})

	s.Run("add os bad uri.", func() {
		pubkey4 := secp256k1.GenPrivKey().PubKey()
		user4Addr := sdk.AccAddress(pubkey4.Address())
		// create account and check default values
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, user4Addr)
		s.Require().NotNil(acc)
		s.Require().Equal(user4Addr, acc.GetAddress())
		s.Require().EqualValues(nil, acc.GetPubKey())
		s.Require().EqualValues(0, acc.GetSequence())
		// set and get the new account.
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		acc1 := s.app.AccountKeeper.GetAccount(s.ctx, user4Addr)
		s.Require().NotNil(acc1)
		// create os locator with ^^ account
		err := s.app.MetadataKeeper.SetOSLocator(s.ctx, user4Addr, s.encryptionKey, "foo.com")
		s.Require().NotEmpty(err)
		r, found := s.app.MetadataKeeper.GetOsLocatorRecord(s.ctx, user4Addr)
		s.Require().Empty(r)
		s.Require().False(found)
	})
}

func (s *KeeperTestSuite) TestModifyOSLocator() {
	s.Run("modify os locator", func() {
		// modify os locator
		err := s.app.MetadataKeeper.ModifyOSLocator(s.ctx, s.user1Addr, s.encryptionKey, "https://bob.com/alice")
		s.Require().Empty(err)
		r, found := s.app.MetadataKeeper.GetOsLocatorRecord(s.ctx, s.user1Addr)
		s.Require().NotEmpty(r)
		s.Require().True(found)
		s.Require().Equal(s.encryptionKey.String(), r.EncryptionKey)
		s.Require().Equal("https://bob.com/alice", r.LocatorUri)
	})
	s.Run("modify os locator invalid uri", func() {
		// modify os locator
		err := s.app.MetadataKeeper.ModifyOSLocator(s.ctx, s.user1Addr, s.encryptionKey, "://bob.com/alice")
		s.Require().NotEmpty(err)
	})

	s.Run("modify os locator invalid uri length", func() {
		// modify os locator
		err := s.app.MetadataKeeper.ModifyOSLocator(s.ctx, s.user1Addr, s.encryptionKey1, "https://www.google.com/search?q=long+url+example&oq=long+uril+&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8")
		s.Require().NotEmpty(err)
		s.Require().Equal("uri length greater than allowed", err.Error())
	})
}

func (s *KeeperTestSuite) TestDeleteOSLocator() {
	s.Run("delete os locator", func() {
		// modify os locator
		err := s.app.MetadataKeeper.RemoveOSLocator(s.ctx, s.user1Addr)
		s.Require().Empty(err)
		r, found := s.app.MetadataKeeper.GetOsLocatorRecord(s.ctx, s.user1Addr)
		s.Require().Empty(r)
		s.Require().False(found)

	})
}

func (s *KeeperTestSuite) TestUnionDistinct() {
	tests := []struct {
		name   string
		inputs [][]string
		output []string
	}{
		{
			"empty in empty out",
			[][]string{},
			[]string{},
		},
		{
			"one set in same set out",
			[][]string{{"a", "b", "c"}},
			[]string{"a", "b", "c"},
		},
		{
			"two dup sets in single entries out",
			[][]string{{"a", "b", "c"}, {"a", "b", "c"}},
			[]string{"a", "b", "c"},
		},
		{
			"unique sets in combined for out",
			[][]string{{"a", "b", "c"}, {"d", "e"}},
			[]string{"a", "b", "c", "d", "e"},
		},
		{
			"empty set filled set in combined for out",
			[][]string{{}, {"a", "b", "c"}},
			[]string{"a", "b", "c"},
		},
		{
			"filled set empty set in combined for out",
			[][]string{{"a", "b", "c"}, {}},
			[]string{"a", "b", "c"},
		},
		{
			"two sets with one common entry in combined correctly for out",
			[][]string{{"a", "b", "c"}, {"d", "a", "e"}},
			[]string{"a", "b", "c", "d", "e"},
		},
		{
			"set with one entry and set with two entries in combined correctly for out",
			[][]string{{"a"}, {"a", "b"}},
			[]string{"a", "b"},
		},
		{
			"set with two entries set with one entry in combined correctly for out",
			[][]string{{"a", "b"}, {"a"}},
			[]string{"a", "b"},
		},
		{
			"set with dups and set with two entries in combined correctly for out",
			[][]string{{"a", "a"}, {"a", "b"}},
			[]string{"a", "b"},
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			output := s.app.MetadataKeeper.UnionDistinct(tc.inputs...)
			sort.Strings(output)
			sort.Strings(tc.output)
			assert.Equal(t, tc.output, output)
		})
	}
}
