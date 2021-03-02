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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/x/metadata/types"
)

type SpecKeeperTestSuite struct {
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

	scopeSpecUUID uuid.UUID
	scopeSpecID   types.MetadataAddress

	contractSpecUUID1 uuid.UUID
	contractSpecID1   types.MetadataAddress
	contractSpecID2   types.MetadataAddress
}

func TestSpecKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(SpecKeeperTestSuite))
}

func (s *SpecKeeperTestSuite) SetupTest() {
	testApp := simapp.Setup(false)
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.scopeSpecUUID = uuid.New()
	s.scopeSpecID = types.ScopeSpecMetadataAddress(s.scopeSpecUUID)

	s.contractSpecUUID1 = uuid.New()
	s.contractSpecID1 = types.ContractSpecMetadataAddress(s.contractSpecUUID1)
	s.contractSpecID2 = types.ContractSpecMetadataAddress(uuid.New())

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

func areEquivalentSetsOfMetaAddresses(arr1 []types.MetadataAddress, arr2 []types.MetadataAddress) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for _, v2 := range arr2 {
		if ! containsMetadataAddress(arr1, v2) {
			return false
		}
	}
	return true
}

func (s *SpecKeeperTestSuite) TestGetSetDeleteContractSpecification() {
	newSpec := types.NewContractSpecification(
		s.contractSpecID1,
		types.NewDescription(
			"TestGetSetDeleteContractSpecification",
			"A description for a unit test contract specification",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user1Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		types.NewContractSpecificationSourceHash("somehash"),
		"someclass",
		[]types.MetadataAddress{},
	)
	require.NotNil(s.T(), newSpec, "test setup failure: NewContractSpecification should not return nil")

	spec1, found1 := s.app.MetadataKeeper.GetContractSpecification(s.ctx, s.contractSpecID1)
	s.False(found1, "1: get contract spec should return false before it has been saved")
	s.NotNil(spec1, "1: get contract spec should always return a non-nil scope spec")

	s.app.MetadataKeeper.SetContractSpecification(s.ctx, *newSpec)

	spec2, found2 := s.app.MetadataKeeper.GetContractSpecification(s.ctx, s.contractSpecID1)
	s.True(found2, "get contract spec should return true after it has been saved")
	s.NotNil(spec2, "get contract spec should always return a non-nil scope spec")
	s.Equal(s.contractSpecID1, spec2.SpecificationId, "2: get contract spec should return a spec containing id provided")

	spec3, found3 := s.app.MetadataKeeper.GetContractSpecification(s.ctx, types.ContractSpecMetadataAddress(uuid.New()))
	s.False(found3, "3: get contract spec should return false for an unknown address")
	s.NotNil(spec3, "3: get contract spec should always return a non-nil scope spec")

	s.app.MetadataKeeper.RemoveContractSpecification(s.ctx, newSpec.SpecificationId)

	spec4, found4 := s.app.MetadataKeeper.GetContractSpecification(s.ctx, s.contractSpecID1)
	s.False(found4, "4: get contract spec should return false after it has been deleted")
	s.NotNil(spec4, "4: get contract spec should always return a non-nil scope spec")
}

func (s *SpecKeeperTestSuite) TestIterateContractSpecs() {
	size := 10
	specs := make([]*types.ContractSpecification, size)
	for i := 0; i < size; i++ {
		specs[i] = types.NewContractSpecification(
			types.ScopeSpecMetadataAddress(uuid.New()),
			types.NewDescription(
				fmt.Sprintf("TestIterateContractSpecs[%d]", i),
				fmt.Sprintf("The description for entry [%d] in a unit test contract specification", i),
				fmt.Sprintf("http://%d.test.net", i),
				fmt.Sprintf("http://%d.test.net/ico.png", i),
			),
			[]string{s.user1Addr.String()},
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
			types.NewContractSpecificationSourceHash(fmt.Sprintf("somehash%d", i)),
			fmt.Sprintf("someclass_%d", i),
			[]types.MetadataAddress{},
		)
		s.app.MetadataKeeper.SetContractSpecification(s.ctx, *specs[i])
	}

	visitedContractSpecIDs := make([]types.MetadataAddress, size)
	count := 0
	err1 := s.app.MetadataKeeper.IterateContractSpecs(s.ctx, func(spec types.ContractSpecification) (stop bool) {
		if containsMetadataAddress(visitedContractSpecIDs, spec.SpecificationId) {
			s.Fail("function IterateContractSpecs visited the same scope specification twice: %s", spec.SpecificationId.String())
		}
		visitedContractSpecIDs[count] = spec.SpecificationId
		count++
		return false
	})
	s.Nil(err1, "1: function IterateContractSpecs should not have returned an error")
	s.Equal(size, count, "number of contract specs iterated through")

	shortCount := 0
	err2 := s.app.MetadataKeeper.IterateContractSpecs(s.ctx, func(spec types.ContractSpecification) (stop bool) {
		shortCount++
		if shortCount == 5 {
			return true
		}
		return false
	})
	s.Nil(err2, "2: function IterateContractSpecs should not have returned an error")
	s.Equal(5, shortCount, "function IterateContractSpecs ignored (stop bool) return value")
}

func (s *SpecKeeperTestSuite) TestIterateContractSpecsForOwner() {
	// Create 5 contract specs. Two owned by user1, Two owned by user2, and One owned by both user1 and user2.
	specs := make([]*types.ContractSpecification, 5)
	user1SpecIDs := make([]types.MetadataAddress, 3)
	user2SpecIDs := make([]types.MetadataAddress, 3)
	specs[0] = types.NewContractSpecification(
		types.ContractSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestIterateContractSpecsForOwner[0]",
			"A description for a unit test contract specification - owner: user1",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user1Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		types.NewContractSpecificationSourceHash("somehash0"),
		"someclass_0",
		[]types.MetadataAddress{},
	)
	user1SpecIDs[0] = specs[0].SpecificationId
	specs[1] = types.NewContractSpecification(
		types.ContractSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestIterateContractSpecsForOwner[1]",
			"A description for a unit test contract specification - owner: user2",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user2Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		types.NewContractSpecificationSourceHash("somehash1"),
		"someclass_1",
		[]types.MetadataAddress{},
	)
	user2SpecIDs[0] = specs[1].SpecificationId
	specs[2] = types.NewContractSpecification(
		types.ContractSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestIterateContractSpecsForOwner[2]",
			"A description for a unit test contract specification - owner: user1",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user1Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		types.NewContractSpecificationSourceHash("somehash2"),
		"someclass_2",
		[]types.MetadataAddress{},
	)
	user1SpecIDs[1] = specs[2].SpecificationId
	specs[3] = types.NewContractSpecification(
		types.ContractSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestIterateContractSpecsForOwner[3]",
			"A description for a unit test contract specification - owner: user2",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user2Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		types.NewContractSpecificationSourceHash("somehash3"),
		"someclass_3",
		[]types.MetadataAddress{},
	)
	user2SpecIDs[1] = specs[3].SpecificationId
	specs[4] = types.NewContractSpecification(
		types.ContractSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestIterateContractSpecsForOwner[4]",
			"A description for a unit test contract specification - owners: user1, user2",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user1Addr.String(), s.user2Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		types.NewContractSpecificationSourceHash("somehash4"),
		"someclass_4",
		[]types.MetadataAddress{},
	)
	user1SpecIDs[2] = specs[4].SpecificationId
	user2SpecIDs[2] = specs[4].SpecificationId

	for _, spec := range specs {
		s.app.MetadataKeeper.SetContractSpecification(s.ctx, *spec)
	}

	// Make sure all user1 scope specs are iterated over
	user1SpecIDsIterated := []types.MetadataAddress{}
	errUser1 := s.app.MetadataKeeper.IterateContractSpecsForOwner(s.ctx, s.user1Addr, func(specID types.MetadataAddress) (stop bool) {
		user1SpecIDsIterated = append(user1SpecIDsIterated, specID)
		return false
	})
	s.Nil(errUser1, "user1: should not have returned an error")
	s.Equal(3, len(user1SpecIDsIterated), "user1: iteration count")
	s.True(areEquivalentSetsOfMetaAddresses(user1SpecIDs, user1SpecIDsIterated),
		"user1: iterated over unexpected contract specs:\n  expected: %v\n    actual: %v\n",
		user1SpecIDs, user1SpecIDsIterated)

	// Make sure all user2 scope specs are iterated over
	user2SpecIDsIterated := []types.MetadataAddress{}
	errUser2 := s.app.MetadataKeeper.IterateContractSpecsForOwner(s.ctx, s.user2Addr, func(specID types.MetadataAddress) (stop bool) {
		user2SpecIDsIterated = append(user2SpecIDsIterated, specID)
		return false
	})
	s.Nil(errUser2, "user2: should not have returned an error")
	s.Equal(3, len(user2SpecIDsIterated), "user2: iteration count")
	s.True(areEquivalentSetsOfMetaAddresses(user2SpecIDs, user2SpecIDsIterated),
		"user2: iterated over unexpected contract specs:\n  expected: %v\n    actual: %v\n",
		user2SpecIDs, user2SpecIDsIterated)

	// Make sure an unknown user address results in zero iterations.
	user3Addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	user3Count := 0
	errUser3 := s.app.MetadataKeeper.IterateContractSpecsForOwner(s.ctx, user3Addr, func(specID types.MetadataAddress) (stop bool) {
		user3Count++
		return false
	})
	s.Nil(errUser3, "user3: should not have returned an error")
	s.Equal(0, user3Count, "user3: iteration count")

	// Make sure the stop bool is being recognized.
	countStop := 0
	errStop := s.app.MetadataKeeper.IterateContractSpecsForOwner(s.ctx, s.user1Addr, func(specID types.MetadataAddress) (stop bool) {
		countStop++
		if countStop == 2 {
			return true
		}
		return false
	})
	s.Nil(errStop, "stop bool: should not have returned an error")
	s.Equal(2, countStop, "stop bool: iteration count")
}

func (s *SpecKeeperTestSuite) TestValidateContractSpecUpdate() {
	recordSpecIDe := s.contractSpecID1.GetRecordSpecAddress("exists")
	recordSpecIDm := s.contractSpecID1.GetRecordSpecAddress("missing")

	// Trick the store into thinking that recordSpecIDe exists.
	metadataStoreKey := s.app.GetKey(types.StoreKey)
	store := s.ctx.KVStore(metadataStoreKey)
	store.Set(recordSpecIDe, []byte{0x01})

	otherContractSpecID := types.ContractSpecMetadataAddress(uuid.New())
	tests := []struct {
		name        string
		existing    *types.ContractSpecification
		proposed    *types.ContractSpecification
		signers     []string
		want        string
	}{
		{
			"existing specificationID does not match proposed specificationID causes error",
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			types.NewContractSpecification(
				otherContractSpecID,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			[]string{s.user1Addr.String()},
			fmt.Sprintf("cannot update contract spec identifier. expected %s, got %s",
				s.contractSpecID1, otherContractSpecID),
		},
		{
			"proposed basic validation causes error",
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			[]string{s.user1Addr.String()},
			"invalid owner addresses count (expected > 0 got: 0)",
		},
		{
			"basic validation not done on existing",
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			[]string{s.user1Addr.String()},
			"",
		},
		{
			"changing owner, only signed by new owner - error",
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user2Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			[]string{s.user2Addr.String()},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1Addr.String()),
		},
		{
			"adding owner, only existing owner needs to sign",
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String(), s.user2Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			[]string{s.user1Addr.String()},
			"",
		},
		{
			"adding owner, both signed - ok too",
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String(), s.user2Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			"",
		},
		{
			"new entry, no signers required",
			nil,
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			[]string{},
			"",
		},
		{
			"RecordSpecIds - proposed index 0 does not exist - fail",
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{recordSpecIDm},
			),
			[]string{s.user1Addr.String()},
			fmt.Sprintf("no record spec exists with id %s", recordSpecIDm),
		},
		{
			"RecordSpecIds - existing index 0 does not exist - ok",
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{recordSpecIDm},
			),
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{},
			),
			[]string{s.user1Addr.String()},
			"",
		},
		{
			"RecordSpecIds - proposed index 0 does not exist - fail",
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{recordSpecIDe, recordSpecIDe},
			),
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{recordSpecIDe, recordSpecIDe, recordSpecIDm},
			),
			[]string{s.user1Addr.String()},
			fmt.Sprintf("no record spec exists with id %s", recordSpecIDm),
		},
		{
			"RecordSpecIds - existing index 2 does not exist - ok",
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{recordSpecIDe, recordSpecIDe, recordSpecIDm},
			),
			types.NewContractSpecification(
				s.contractSpecID1,
				types.NewDescription(
					"TestValidateContractSpecUpdate",
					"A description for a unit test contract specification",
					"http://test.net",
					"http://test.net/ico.png",
				),
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				types.NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]types.MetadataAddress{recordSpecIDe, recordSpecIDe},
			),
			[]string{s.user1Addr.String()},
			"",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateContractSpecUpdate(s.ctx, tt.existing, *tt.proposed, tt.signers)
			if err != nil {
				require.Equal(t, tt.want, err.Error(), "ScopeSpec Keeper ValidateContractSpecUpdate error")
			} else if len(tt.want) > 0 {
				t.Errorf("ScopeSpec Keeper ValidateContractSpecUpdate error = nil, expected: %s", tt.want)
			}
		})
	}

	// I'm not really sure what all gets shared between unit tests.
	// So just to be on the safe side...
	store.Delete(recordSpecIDe)
}

func (s *SpecKeeperTestSuite) TestGetSetDeleteScopeSpecification() {
	newSpec := types.NewScopeSpecification(
		s.scopeSpecID,
		types.NewDescription(
			"TestGetSetScopeSpecification",
			"A description for a unit test scope specification",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user1Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		[]types.MetadataAddress{s.contractSpecID1},
	)
	require.NotNil(s.T(), newSpec, "test setup failure: NewScopeSpecification should not return nil")

	spec1, found1 := s.app.MetadataKeeper.GetScopeSpecification(s.ctx, s.scopeSpecID)
	s.False(found1, "1: get scope spec should return false before it has been saved")
	s.NotNil(spec1, "1: get scope spec should always return a non-nil scope spec")

	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, *newSpec)

	spec2, found2 := s.app.MetadataKeeper.GetScopeSpecification(s.ctx, s.scopeSpecID)
	s.True(found2, "get scope spec should return true after it has been saved")
	s.NotNil(spec2, "get scope spec should always return a non-nil scope spec")
	s.Equal(s.scopeSpecID, spec2.SpecificationId, "2: get scope spec should return a spec containing id provided")

	spec3, found3 := s.app.MetadataKeeper.GetScopeSpecification(s.ctx, types.ScopeSpecMetadataAddress(uuid.New()))
	s.False(found3, "3: get scope spec should return false for an unknown address")
	s.NotNil(spec3, "3: get scope spec should always return a non-nil scope spec")

	s.app.MetadataKeeper.RemoveScopeSpecification(s.ctx, newSpec.SpecificationId)

	spec4, found4 := s.app.MetadataKeeper.GetScopeSpecification(s.ctx, s.scopeSpecID)
	s.False(found4, "4: get scope spec should return false after it has been deleted")
	s.NotNil(spec4, "4: get scope spec should always return a non-nil scope spec")
}

func (s *SpecKeeperTestSuite) TestIterateScopeSpecs() {
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
			[]types.MetadataAddress{s.contractSpecID1},
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

func (s *SpecKeeperTestSuite) TestIterateScopeSpecsForOwner() {
	// Create 5 scope specs. Two owned by user1, Two owned by user2, and One owned by both user1 and user2.
	scopeSpecs := make([]*types.ScopeSpecification, 5)
	user1ScopeSpecIDs := make([]types.MetadataAddress, 3)
	user2ScopeSpecIDs := make([]types.MetadataAddress, 3)
	scopeSpecs[0] = types.NewScopeSpecification(
		types.ScopeSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestIterateScopeSpecsForOwner[0]",
			"A description for a unit test scope specification - owner: user1",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user1Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		[]types.MetadataAddress{s.contractSpecID1},
	)
	user1ScopeSpecIDs[0] = scopeSpecs[0].SpecificationId
	scopeSpecs[1] = types.NewScopeSpecification(
		types.ScopeSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestIterateScopeSpecsForOwner[1]",
			"A description for a unit test scope specification - owner: user2",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user2Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		[]types.MetadataAddress{s.contractSpecID1},
	)
	user2ScopeSpecIDs[0] = scopeSpecs[1].SpecificationId
	scopeSpecs[2] = types.NewScopeSpecification(
		types.ScopeSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestIterateScopeSpecsForOwner[2]",
			"A description for a unit test scope specification - owner: user1",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user1Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		[]types.MetadataAddress{s.contractSpecID2},
	)
	user1ScopeSpecIDs[1] = scopeSpecs[2].SpecificationId
	scopeSpecs[3] = types.NewScopeSpecification(
		types.ScopeSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestIterateScopeSpecsForOwner[3]",
			"A description for a unit test scope specification - owner: user2",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user2Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		[]types.MetadataAddress{s.contractSpecID2},
	)
	user2ScopeSpecIDs[1] = scopeSpecs[3].SpecificationId
	scopeSpecs[4] = types.NewScopeSpecification(
		types.ScopeSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestIterateScopeSpecsForOwner[4]",
			"A description for a unit test scope specification - owners: user1, user2",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user1Addr.String(), s.user2Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		[]types.MetadataAddress{s.contractSpecID1, s.contractSpecID2},
	)
	user1ScopeSpecIDs[2] = scopeSpecs[4].SpecificationId
	user2ScopeSpecIDs[2] = scopeSpecs[4].SpecificationId

	for _, spec := range scopeSpecs {
		s.app.MetadataKeeper.SetScopeSpecification(s.ctx, *spec)
	}

	// Make sure all user1 scope specs are iterated over
	user1ScopeSpecIDsIterated := []types.MetadataAddress{}
	errUser1 := s.app.MetadataKeeper.IterateScopeSpecsForOwner(s.ctx, s.user1Addr, func(specID types.MetadataAddress) (stop bool) {
		user1ScopeSpecIDsIterated = append(user1ScopeSpecIDsIterated, specID)
		return false
	})
	s.Nil(errUser1, "user1: should not have returned an error")
	s.Equal(3, len(user1ScopeSpecIDsIterated), "user1: iteration count")
	s.True(areEquivalentSetsOfMetaAddresses(user1ScopeSpecIDs, user1ScopeSpecIDsIterated),
		"user1: iterated over unexpected scope specs:\n  expected: %v\n    actual: %v\n",
		user1ScopeSpecIDs, user1ScopeSpecIDsIterated)

	// Make sure all user2 scope specs are iterated over
	user2ScopeSpecIDsIterated := []types.MetadataAddress{}
	errUser2 := s.app.MetadataKeeper.IterateScopeSpecsForOwner(s.ctx, s.user2Addr, func(specID types.MetadataAddress) (stop bool) {
		user2ScopeSpecIDsIterated = append(user2ScopeSpecIDsIterated, specID)
		return false
	})
	s.Nil(errUser2, "user2: should not have returned an error")
	s.Equal(3, len(user2ScopeSpecIDsIterated), "user2: iteration count")
	s.True(areEquivalentSetsOfMetaAddresses(user2ScopeSpecIDs, user2ScopeSpecIDsIterated),
		"user2: iterated over unexpected scope specs:\n  expected: %v\n    actual: %v\n",
		user2ScopeSpecIDs, user2ScopeSpecIDsIterated)

	// Make sure an unknown user address results in zero iterations.
	user3Addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	user3Count := 0
	errUser3 := s.app.MetadataKeeper.IterateScopeSpecsForOwner(s.ctx, user3Addr, func(specID types.MetadataAddress) (stop bool) {
		user3Count++
		return false
	})
	s.Nil(errUser3, "user3: should not have returned an error")
	s.Equal(0, user3Count, "user3: iteration count")

	// Make sure the stop bool is being recognized.
	countStop := 0
	errStop := s.app.MetadataKeeper.IterateScopeSpecsForOwner(s.ctx, s.user1Addr, func(specID types.MetadataAddress) (stop bool) {
		countStop++
		if countStop == 2 {
			return true
		}
		return false
	})
	s.Nil(errStop, "stop bool: should not have returned an error")
	s.Equal(2, countStop, "stop bool: iteration count")
}

func (s *SpecKeeperTestSuite) TestIterateScopeSpecsForContractSpec() {
	// Create 5 scope specs. Two with just contract spec 1, two with just contract spec 2, and one with both contract specs 1 and 2.
	scopeSpecs := make([]*types.ScopeSpecification, 5)
	contractSpec1ScopeSpecIDs := make([]types.MetadataAddress, 3)
	contractSpec2ScopeSpecIDs := make([]types.MetadataAddress, 3)
	scopeSpecs[0] = types.NewScopeSpecification(
		types.ScopeSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestGetSetScopeSpecification[0]",
			"A description for a unit test scope specification - owner: user1",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user1Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		[]types.MetadataAddress{s.contractSpecID1},
	)
	contractSpec1ScopeSpecIDs[0] = scopeSpecs[0].SpecificationId
	scopeSpecs[1] = types.NewScopeSpecification(
		types.ScopeSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestGetSetScopeSpecification[1]",
			"A description for a unit test scope specification - owner: user2",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user2Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		[]types.MetadataAddress{s.contractSpecID1},
	)
	contractSpec1ScopeSpecIDs[1] = scopeSpecs[1].SpecificationId
	scopeSpecs[2] = types.NewScopeSpecification(
		types.ScopeSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestGetSetScopeSpecification[2]",
			"A description for a unit test scope specification - owner: user1",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user1Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		[]types.MetadataAddress{s.contractSpecID2},
	)
	contractSpec2ScopeSpecIDs[0] = scopeSpecs[2].SpecificationId
	scopeSpecs[3] = types.NewScopeSpecification(
		types.ScopeSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestGetSetScopeSpecification[3]",
			"A description for a unit test scope specification - owner: user2",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user2Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		[]types.MetadataAddress{s.contractSpecID2},
	)
	contractSpec2ScopeSpecIDs[1] = scopeSpecs[3].SpecificationId
	scopeSpecs[4] = types.NewScopeSpecification(
		types.ScopeSpecMetadataAddress(uuid.New()),
		types.NewDescription(
			"TestGetSetScopeSpecification[4]",
			"A description for a unit test scope specification - owners: user1, user2",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user1Addr.String(), s.user2Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		[]types.MetadataAddress{s.contractSpecID1, s.contractSpecID2},
	)
	contractSpec1ScopeSpecIDs[2] = scopeSpecs[4].SpecificationId
	contractSpec2ScopeSpecIDs[2] = scopeSpecs[4].SpecificationId

	for _, spec := range scopeSpecs {
		s.app.MetadataKeeper.SetScopeSpecification(s.ctx, *spec)
	}

	// Make sure all contract spec 1 scope specs are iterated over
	contractSpec1ScopeSpecIDsIterated := []types.MetadataAddress{}
	errContractSpec1 := s.app.MetadataKeeper.IterateScopeSpecsForContractSpec(s.ctx, s.contractSpecID1, func(specID types.MetadataAddress) (stop bool) {
		contractSpec1ScopeSpecIDsIterated = append(contractSpec1ScopeSpecIDsIterated, specID)
		return false
	})
	s.Nil(errContractSpec1, "contract spec 1: should not have returned an error")
	s.Equal(3, len(contractSpec1ScopeSpecIDsIterated), "contract spec 1: iteration count")
	s.True(areEquivalentSetsOfMetaAddresses(contractSpec1ScopeSpecIDs, contractSpec1ScopeSpecIDsIterated),
		"contract spec 1: iterated over unexpected scope specs:\n  expected: %v\n    actual: %v\n",
		contractSpec1ScopeSpecIDs, contractSpec1ScopeSpecIDsIterated)

	// Make sure all contract spec 2 scope specs are iterated over
	contractSpec2ScopeSpecIDsIterated := []types.MetadataAddress{}
	errContractSpec2 := s.app.MetadataKeeper.IterateScopeSpecsForContractSpec(s.ctx, s.contractSpecID2, func(specID types.MetadataAddress) (stop bool) {
		contractSpec2ScopeSpecIDsIterated = append(contractSpec2ScopeSpecIDsIterated, specID)
		return false
	})
	s.Nil(errContractSpec2, "contract spec 2: should not have returned an error")
	s.Equal(3, len(contractSpec2ScopeSpecIDsIterated), "contract spec 2: iteration count")
	s.True(areEquivalentSetsOfMetaAddresses(contractSpec2ScopeSpecIDs, contractSpec2ScopeSpecIDsIterated),
		"contract spec 2: iterated over unexpected scope specs:\n  expected: %v\n    actual: %v\n",
		contractSpec2ScopeSpecIDs, contractSpec2ScopeSpecIDsIterated)

	// Make sure an unknown contract spec results in zero iterations.
	contractSpecID3 := types.ContractSpecMetadataAddress(uuid.New())
	contractSpec3Count := 0
	errContractSpec3 := s.app.MetadataKeeper.IterateScopeSpecsForContractSpec(s.ctx, contractSpecID3, func(specID types.MetadataAddress) (stop bool) {
		contractSpec3Count++
		return false
	})
	s.Nil(errContractSpec3, "contract spec 3: should not have returned an error")
	s.Equal(0, contractSpec3Count, "contract spec 3: iteration count")

	// Make sure the stop bool is being recognized.
	countStop := 0
	errStop := s.app.MetadataKeeper.IterateScopeSpecsForContractSpec(s.ctx, s.contractSpecID1, func(specID types.MetadataAddress) (stop bool) {
		countStop++
		if countStop == 2 {
			return true
		}
		return false
	})
	s.Nil(errStop, "stop bool: should not have returned an error")
	s.Equal(2, countStop, "stop bool: iteration count")
}

func (s *SpecKeeperTestSuite) TestValidateScopeSpecUpdate() {
	// Trick the store into thinking that s.contractSpecID1 and s.contractSpecID2 exists.
	metadataStoreKey := s.app.GetKey(types.StoreKey)
	store := s.ctx.KVStore(metadataStoreKey)
	store.Set(s.contractSpecID1, []byte{0x01})
	store.Set(s.contractSpecID2, []byte{0x01})

	otherScopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())
	otherContractSpecID := types.ContractSpecMetadataAddress(uuid.New())
	tests := []struct {
		name        string
		existing    *types.ScopeSpecification
		proposed    *types.ScopeSpecification
		signers     []string
		want        string
	}{
		{
			"existing specificationID does not match proposed specificationID causes error",
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			types.NewScopeSpecification(
				otherScopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			[]string{s.user1Addr.String()},
			fmt.Sprintf("cannot update scope spec identifier. expected %s, got %s",
				s.scopeSpecID, otherScopeSpecID),
		},
		{
			"proposed basic validation causes error",
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			[]string{s.user1Addr.String()},
			"the ScopeSpecification must have at least one owner",
		},
		{
			"basic validation not done on existing",
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			[]string{s.user1Addr.String()},
			"",
		},
		{
			"changing owner, only signed by new owner - error",
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user2Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			[]string{s.user2Addr.String()},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1Addr.String()),
		},
		{
			"adding signer, only existing owner needs to sign",
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String(), s.user2Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			[]string{s.user1Addr.String()},
			"",
		},
		{
			"adding signer, both signed - ok too",
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String(), s.user2Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			"",
		},
		{
			"new entry, no signers required",
			nil,
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			[]string{},
			"",
		},
		{
			"adding unknown contract spec - error",
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1, otherContractSpecID},
			),
			[]string{s.user1Addr.String()},
			fmt.Sprintf("no contract spec exists with id %s", otherContractSpecID),
		},
		{
			"adding known contract spec - ok",
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1, s.contractSpecID2},
			),
			[]string{s.user1Addr.String()},
			"",
		},
		{
			"changing to known contract spec - ok",
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID1},
			),
			types.NewScopeSpecification(
				s.scopeSpecID,
				nil,
				[]string{s.user1Addr.String()},
				[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				[]types.MetadataAddress{s.contractSpecID2},
			),
			[]string{s.user1Addr.String()},
			"",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateScopeSpecUpdate(s.ctx, tt.existing, *tt.proposed, tt.signers)
			if err != nil {
				require.Equal(t, tt.want, err.Error(), "ScopeSpec Keeper ValidateScopeSpecUpdate error")
			} else if len(tt.want) > 0 {
				t.Errorf("ScopeSpec Keeper ValidateScopeSpecUpdate error = nil, expected: %s", tt.want)
			}
		})
	}

	// I'm not really sure what all gets shared between unit tests.
	// So just to be on the safe side...
	store.Delete(s.contractSpecID1)
	store.Delete(s.contractSpecID2)
}

func (s *SpecKeeperTestSuite) TestValidateAllOwnersAreSigners() {
	tests := []struct {
		name        string
		owners      []string
		signers     []string
		want        string
	}{
		{
			"Scope Spec with 1 owner: no signers - error",
			[]string{s.user1Addr.String()},
			[]string{},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1Addr.String()),
		},
		{
			"Scope Spec with 1 owner: not in signers list - error",
			[]string{s.user1Addr.String()},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1Addr.String()),
		},
		{
			"Scope Spec with 1 owner: in signers list with non-owners - ok",
			[]string{s.user1Addr.String()},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1Addr.String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			"",
		},
		{
			"Scope Spec with 1 owner: only signer in list - ok",
			[]string{s.user1Addr.String()},
			[]string{s.user1Addr.String()},
			"",
		},
		{
			"Scope Spec with 2 owners: no signers - error",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1Addr.String()),
		},
		{
			"Scope Spec with 2 owners: neither in signers list - error",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1Addr.String()),
		},
		{
			"Scope Spec with 2 owners: one in signers list with non-owners - error",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1Addr.String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2Addr.String()),
		},
		{
			"Scope Spec with 2 owners: the other in signers list with non-owners - error",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user2Addr.String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1Addr.String()),
		},
		{
			"Scope Spec with 2 owners: both in signers list with non-owners - ok",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user2Addr.String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1Addr.String()},
			"",
		},
		{
			"Scope Spec with 2 owners: only both in signers list - ok",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			"",
		},
		{
			"Scope Spec with 2 owners: only both in signers list, opposite order - ok",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{s.user2Addr.String(), s.user1Addr.String()},
			"",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateAllOwnersAreSigners(tt.owners, tt.signers)
			if err != nil {
				require.Equal(t, tt.want, err.Error(), "ScopeSpec Keeper ValidateScopeSpecAllOwnersAreSigners error")
			} else if len(tt.want) > 0 {
				t.Errorf("ScopeSpec Keeper ValidateAllOwnersAreSigners error = nil, expected: %s", tt.want)
			}
		})
	}
}
