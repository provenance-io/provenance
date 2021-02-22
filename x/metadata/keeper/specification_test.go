package keeper_test

import (
	"fmt"
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

func TestScopeSpecKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(ScopeSpecKeeperTestSuite))
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
	s.queryClient = types.NewQueryClient(queryHelper)
}

func containsMetadataAddress(arr []types.MetadataAddress, newVal types.MetadataAddress) bool {
	found := false
	for _, v := range arr {
		if v.Equals(newVal) {
			found = true
			break
		}
	}
	return found
}

func (s *ScopeSpecKeeperTestSuite) TestGetSetDeleteScopeSpecification() {
	newSpec := types.NewScopeSpecification(
		s.specID,
		types.NewDescription(
			"TestGetSetScopeSpecification",
			"A description for a unit test scope specification",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user1Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		[]types.MetadataAddress{types.GroupSpecMetadataAddress(uuid.New())},
	)
	s.NotNil(newSpec, "test setup failure: NewScopeSpecification should not return nil")

	spec1, found1 := s.app.MetadataKeeper.GetScopeSpecification(s.ctx, s.specID)
	s.False(found1, "1: get scope spec should return false before it has been saved")
	s.NotNil(spec1, "1: get scope spec should always return a non-nil scope spec")

	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, *newSpec)

	spec2, found2 := s.app.MetadataKeeper.GetScopeSpecification(s.ctx, s.specID)
	s.True(found2, "get scope spec should return true after it has been saved")
	s.NotNil(spec2, "get scope spec should always return a non-nil scope spec")
	s.Equal(s.specID, spec2.SpecificationId, "2: get scope spec should return a spec containing id provided")

	spec3, found3 := s.app.MetadataKeeper.GetScopeSpecification(s.ctx, types.ScopeSpecMetadataAddress(uuid.New()))
	s.False(found3, "3: get scope spec should return false for an unknown address")
	s.NotNil(spec3, "3: get scope spec should always return a non-nil scope spec")

	s.app.MetadataKeeper.DeleteScopeSpecification(s.ctx, newSpec.SpecificationId)

	spec4, found4 := s.app.MetadataKeeper.GetScopeSpecification(s.ctx, s.specID)
	s.False(found4, "4: get scope spec should return false after it has been deleted")
	s.NotNil(spec4, "4: get scope spec should always return a non-nil scope spec")
}

func (s *ScopeSpecKeeperTestSuite) TestIterateScopeSpecs() {
	size := 10
	scopeSpecs := make([]*types.ScopeSpecification, size)
	for i := 0; i < size; i++ {
		scopeSpecs[i] = types.NewScopeSpecification(
			types.ScopeSpecMetadataAddress(uuid.New()),
			types.NewDescription(
				fmt.Sprintf("TestIterateScopeSpecs[%d]", i),
				fmt.Sprintf("The description for entry [%d] in a unit test scope specification", i),
				fmt.Sprintf("http://%d.test.net", i),
				fmt.Sprintf("http://%d.test.net/ico.png", i),
			),
			[]string{s.user1Addr.String()},
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
			[]types.MetadataAddress{types.GroupSpecMetadataAddress(uuid.New())},
		)
		s.app.MetadataKeeper.SetScopeSpecification(s.ctx, *scopeSpecs[i])
	}

	visitedScopeSpecAddresses := make([]types.MetadataAddress, size)
	count := 0
	err1 := s.app.MetadataKeeper.IterateScopeSpecs(s.ctx, func(spec types.ScopeSpecification) (stop bool) {
		if containsMetadataAddress(visitedScopeSpecAddresses, spec.SpecificationId) {
			s.Fail("function IterateScopeSpecs visited the same scope specification twice: %s", spec.SpecificationId.String())
		}
		visitedScopeSpecAddresses[count] = spec.SpecificationId
		count++
		return false
	})
	s.Nil(err1, "1: function IterateScopeSpecs should not have returned an error")
	s.Equal(size, count, "number of scope specs iterated through")

	shortCount := 0
	err2 := s.app.MetadataKeeper.IterateScopeSpecs(s.ctx, func(spec types.ScopeSpecification) (stop bool) {
		shortCount++
		if shortCount == 5 {
			return true
		}
		return false
	})
	s.Nil(err2, "2: function IterateScopeSpecs should not have returned an error")
	s.Equal(5, shortCount, "function IterateScopeSpecs ignored (stop bool) return value")
}

func (s *ScopeSpecKeeperTestSuite) TestIterateScopeSpecsForAddress() {
	// TODO: implement
}

func (s *ScopeSpecKeeperTestSuite) TestIterateScopeSpecsForContractSpec() {
	// TODO: implement
}

func (s *ScopeSpecKeeperTestSuite) TestValidateScopeSpecUpdate() {
	// TODO: implement
}

func (s *ScopeSpecKeeperTestSuite) TestValidateScopeSpecAllOwnersAreSigners() {
	// TODO: implement
}
