package keeper_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SpecKeeperTestSuite struct {
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
		if !containsMetadataAddress(arr1, v2) {
			return false
		}
	}
	return true
}

func containsRecSpec(arr []*types.RecordSpecification, newVal *types.RecordSpecification) bool {
	found := false
	for _, v := range arr {
		if v.SpecificationId.Equals(newVal.SpecificationId) {
			found = true
			break
		}
	}
	return found
}

func areEquivalentSetsOfRecSpecs(arr1 []*types.RecordSpecification, arr2 []*types.RecordSpecification) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for _, v2 := range arr2 {
		if !containsRecSpec(arr1, v2) {
			return false
		}
	}
	return true
}

func asRecSpecAddrOrPanic(ma types.MetadataAddress, name string) types.MetadataAddress {
	retval, err := ma.AsRecordSpecAddress(name)
	if err != nil {
		panic(err)
	}
	return retval
}

func (s *SpecKeeperTestSuite) TestGetSetRemoveRecordSpecification() {
	recordName := "record name"
	recSpecID := types.RecordSpecMetadataAddress(s.contractSpecUUID1, recordName)
	newSpec := types.NewRecordSpecification(
		recSpecID,
		recordName,
		[]*types.InputSpecification{
			types.NewInputSpecification(
				"input name 1", "type name 1",
				types.NewInputSpecificationSourceHash("source hash 1"),
			),
			types.NewInputSpecification(
				"input name 2", "type name 2",
				types.NewInputSpecificationSourceHash("source hash 2"),
			),
		},
		"type name",
		types.DefinitionType_DEFINITION_TYPE_RECORD,
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
	)
	require.NotNil(s.T(), newSpec, "test setup failure: NewRecordSpecification should not return nil")

	spec1, found1 := s.app.MetadataKeeper.GetRecordSpecification(s.ctx, recSpecID)
	s.False(found1, "1: get record spec should return false before it has been saved")
	s.NotNil(spec1, "1: get record spec should always return a non-nil record spec")

	s.app.MetadataKeeper.SetRecordSpecification(s.ctx, *newSpec)

	spec2, found2 := s.app.MetadataKeeper.GetRecordSpecification(s.ctx, recSpecID)
	s.True(found2, "get record spec should return true after it has been saved")
	s.NotNil(spec2, "get record spec should always return a non-nil record spec")
	s.Equal(recSpecID, spec2.SpecificationId, "2: get record spec should return a spec containing id provided")

	spec3, found3 := s.app.MetadataKeeper.GetRecordSpecification(s.ctx, types.RecordSpecMetadataAddress(s.contractSpecUUID1, "nope"))
	s.False(found3, "3: get record spec should return false for an unknown address")
	s.NotNil(spec3, "3: get record spec should always return a non-nil record spec")

	spec4, found4 := s.app.MetadataKeeper.GetRecordSpecification(s.ctx, types.RecordSpecMetadataAddress(uuid.New(), recordName))
	s.False(found4, "4: get record spec should return false for an unknown address")
	s.NotNil(spec4, "4: get record spec should always return a non-nil record spec")

	remErr1 := s.app.MetadataKeeper.RemoveRecordSpecification(s.ctx, newSpec.SpecificationId)
	s.Nil(remErr1, "5: remove should not return any error")

	spec6, found6 := s.app.MetadataKeeper.GetRecordSpecification(s.ctx, recSpecID)
	s.False(found6, "6: get record spec should return false after it has been removed")
	s.NotNil(spec6, "6: get record spec should always return a non-nil record spec")

	remErr2 := s.app.MetadataKeeper.RemoveRecordSpecification(s.ctx, recSpecID)
	s.Equal(
		fmt.Errorf("record specification with id %s not found", recSpecID),
		remErr2,
		"7: remove error message when not found",
	)
}

func (s *SpecKeeperTestSuite) TestIterateRecordSpecs() {
	size := 10
	specs := make([]*types.RecordSpecification, size)
	for i := 0; i < size; i++ {
		contractSpecUUID := uuid.New()
		recordName := fmt.Sprintf("record name %d", i)
		specs[i] = types.NewRecordSpecification(
			types.RecordSpecMetadataAddress(contractSpecUUID, recordName),
			recordName,
			[]*types.InputSpecification{
				types.NewInputSpecification(
					fmt.Sprintf("input name [%d] 1", i),
					fmt.Sprintf("type name [%d] 1", i),
					types.NewInputSpecificationSourceHash(fmt.Sprintf("source hash [%d] 1", i)),
				),
				types.NewInputSpecification(
					fmt.Sprintf("input name [%d] 2", i),
					fmt.Sprintf("type name [%d] 2", i),
					types.NewInputSpecificationSourceHash(fmt.Sprintf("source hash [%d] 2", i)),
				),
				types.NewInputSpecification(
					fmt.Sprintf("input name [%d] 3", i),
					fmt.Sprintf("type name [%d] 3", i),
					types.NewInputSpecificationSourceRecordID(types.RecordMetadataAddress(uuid.New(), recordName)),
				),
			},
			fmt.Sprintf("type name %d", i),
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		)
		s.app.MetadataKeeper.SetRecordSpecification(s.ctx, *specs[i])
	}

	visitedRecordSpecIDs := make([]types.MetadataAddress, size)
	count := 0
	err1 := s.app.MetadataKeeper.IterateRecordSpecs(s.ctx, func(spec types.RecordSpecification) (stop bool) {
		if containsMetadataAddress(visitedRecordSpecIDs, spec.SpecificationId) {
			s.Fail("function IterateRecordSpecs visited the same record specification twice: %s", spec.SpecificationId.String())
		}
		visitedRecordSpecIDs[count] = spec.SpecificationId
		count++
		return false
	})
	s.Nil(err1, "1: function IterateRecordSpecs should not have returned an error")
	s.Equal(size, count, "number of record specs iterated through")

	shortCount := 0
	err2 := s.app.MetadataKeeper.IterateRecordSpecs(s.ctx, func(spec types.RecordSpecification) (stop bool) {
		shortCount++
		if shortCount == 5 {
			return true
		}
		return false
	})
	s.Nil(err2, "2: function IterateRecordSpecs should not have returned an error")
	s.Equal(5, shortCount, "function IterateRecordSpecs ignored (stop bool) return value")
}

func (s *SpecKeeperTestSuite) TestIterateRecordSpecsForOwner() {
	// Create 3 contract specs. One owned by user1, One owned by user2, and One owned by both user1 and user2.
	contractSpecs := []*types.ContractSpecification{
		types.NewContractSpecification(
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
		),
		types.NewContractSpecification(
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
		),
		types.NewContractSpecification(
			types.ContractSpecMetadataAddress(uuid.New()),
			types.NewDescription(
				"TestIterateContractSpecsForOwner[2]",
				"A description for a unit test contract specification - owners: user1, user2",
				"http://test.net",
				"http://test.net/ico.png",
			),
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
			types.NewContractSpecificationSourceHash("somehash2"),
			"someclass_2",
		),
	}
	for _, spec := range contractSpecs {
		s.app.MetadataKeeper.SetContractSpecification(s.ctx, *spec)
	}

	// Create 2 record specifications for each contract specification
	recordSpecs := []*types.RecordSpecification{
		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[0].SpecificationId, "contractspec0recspec0"),
			"contractspec0recspec0",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec0", "inputspectypename0",
					types.NewInputSpecificationSourceHash("sourcehash0"),
				),
			},
			"typename0",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),
		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[0].SpecificationId, "contractspec0recspec1"),
			"contractspec0recspec1",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec1", "inputspectypename1",
					types.NewInputSpecificationSourceHash("sourcehash1"),
				),
			},
			"typename1",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),

		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[1].SpecificationId, "contractspec1recspec2"),
			"contractspec1recspec2",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec2", "inputspectypename2",
					types.NewInputSpecificationSourceHash("sourcehash2"),
				),
			},
			"typename2",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),
		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[1].SpecificationId, "contractspec1recspec3"),
			"contractspec1recspec3",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec3", "inputspectypename3",
					types.NewInputSpecificationSourceHash("sourcehash3"),
				),
			},
			"typename3",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),

		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[2].SpecificationId, "contractspec2recspec4"),
			"contractspec2recspec4",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec4", "inputspectypename4",
					types.NewInputSpecificationSourceHash("sourcehash4"),
				),
			},
			"typename4",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),
		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[2].SpecificationId, "contractspec2recspec5"),
			"contractspec2recspec5",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec5", "inputspectypename5",
					types.NewInputSpecificationSourceHash("sourcehash5"),
				),
			},
			"typename5",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),
	}
	for _, spec := range recordSpecs {
		s.app.MetadataKeeper.SetRecordSpecification(s.ctx, *spec)
	}

	user1SpecIDs := []types.MetadataAddress{
		recordSpecs[0].SpecificationId, recordSpecs[1].SpecificationId,
		recordSpecs[4].SpecificationId, recordSpecs[5].SpecificationId,
	}
	user2SpecIDs := []types.MetadataAddress{
		recordSpecs[2].SpecificationId, recordSpecs[3].SpecificationId,
		recordSpecs[4].SpecificationId, recordSpecs[5].SpecificationId,
	}

	// Make sure all user1 record specs are iterated over
	user1SpecIDsIterated := []types.MetadataAddress{}
	errUser1 := s.app.MetadataKeeper.IterateRecordSpecsForOwner(s.ctx, s.user1Addr, func(specID types.MetadataAddress) (stop bool) {
		user1SpecIDsIterated = append(user1SpecIDsIterated, specID)
		return false
	})
	s.Nil(errUser1, "user1: should not have returned an error")
	s.Equal(4, len(user1SpecIDsIterated), "user1: iteration count")
	s.True(areEquivalentSetsOfMetaAddresses(user1SpecIDs, user1SpecIDsIterated),
		"user1: iterated over unexpected record specs:\n  expected: %v\n    actual: %v\n",
		user1SpecIDs, user1SpecIDsIterated)

	// Make sure all user2 record specs are iterated over
	user2SpecIDsIterated := []types.MetadataAddress{}
	errUser2 := s.app.MetadataKeeper.IterateRecordSpecsForOwner(s.ctx, s.user2Addr, func(specID types.MetadataAddress) (stop bool) {
		user2SpecIDsIterated = append(user2SpecIDsIterated, specID)
		return false
	})
	s.Nil(errUser2, "user2: should not have returned an error")
	s.Equal(4, len(user2SpecIDsIterated), "user2: iteration count")
	s.True(areEquivalentSetsOfMetaAddresses(user2SpecIDs, user2SpecIDsIterated),
		"user2: iterated over unexpected record specs:\n  expected: %v\n    actual: %v\n",
		user2SpecIDs, user2SpecIDsIterated)

	// Make sure an unknown user address results in zero iterations.
	user3Addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	user3Count := 0
	errUser3 := s.app.MetadataKeeper.IterateRecordSpecsForOwner(s.ctx, user3Addr, func(specID types.MetadataAddress) (stop bool) {
		user3Count++
		return false
	})
	s.Nil(errUser3, "user3: should not have returned an error")
	s.Equal(0, user3Count, "user3: iteration count")

	// Make sure the stop bool is being recognized.
	countStop := 0
	errStop := s.app.MetadataKeeper.IterateRecordSpecsForOwner(s.ctx, s.user1Addr, func(specID types.MetadataAddress) (stop bool) {
		countStop++
		if countStop == 2 {
			return true
		}
		return false
	})
	s.Nil(errStop, "stop bool: should not have returned an error")
	s.Equal(2, countStop, "stop bool: iteration count")
}

func (s *SpecKeeperTestSuite) TestIterateRecordSpecsForContractSpec() {
	// Create 3 contract specs.
	contractSpecs := []*types.ContractSpecification{
		types.NewContractSpecification(
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
		),
		types.NewContractSpecification(
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
		),
		types.NewContractSpecification(
			types.ContractSpecMetadataAddress(uuid.New()),
			types.NewDescription(
				"TestIterateContractSpecsForOwner[2]",
				"A description for a unit test contract specification - owner: user1, user2",
				"http://test.net",
				"http://test.net/ico.png",
			),
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
			types.NewContractSpecificationSourceHash("somehash2"),
			"someclass_2",
		),
	}
	for _, spec := range contractSpecs {
		s.app.MetadataKeeper.SetContractSpecification(s.ctx, *spec)
	}

	// Create 3 record specs for the 1st contract spec, 2 for the 2nd, and none for the 3rd.
	recordSpecs := []*types.RecordSpecification{
		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[0].SpecificationId, "contractspec0recspec0"),
			"contractspec0recspec0",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec0", "inputspectypename0",
					types.NewInputSpecificationSourceHash("sourcehash0"),
				),
			},
			"typename0",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),
		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[0].SpecificationId, "contractspec0recspec1"),
			"contractspec0recspec1",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec1", "inputspectypename1",
					types.NewInputSpecificationSourceHash("sourcehash1"),
				),
			},
			"typename1",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),
		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[0].SpecificationId, "contractspec1recspec2"),
			"contractspec0recspec2",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec2", "inputspectypename2",
					types.NewInputSpecificationSourceHash("sourcehash2"),
				),
			},
			"typename2",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),

		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[1].SpecificationId, "contractspec1recspec3"),
			"contractspec1recspec3",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec3", "inputspectypename3",
					types.NewInputSpecificationSourceHash("sourcehash3"),
				),
			},
			"typename3",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),
		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[1].SpecificationId, "contractspec1recspec4"),
			"contractspec1recspec4",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec4", "inputspectypename4",
					types.NewInputSpecificationSourceHash("sourcehash4"),
				),
			},
			"typename4",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),
	}
	for _, spec := range recordSpecs {
		s.app.MetadataKeeper.SetRecordSpecification(s.ctx, *spec)
	}

	contractSpec0RecSpecIDs := []types.MetadataAddress{
		recordSpecs[0].SpecificationId, recordSpecs[1].SpecificationId, recordSpecs[2].SpecificationId,
	}
	contractSpec1RecSpecIDs := []types.MetadataAddress{
		recordSpecs[3].SpecificationId, recordSpecs[4].SpecificationId,
	}
	contractSpec2RecSpecIDs := []types.MetadataAddress{}

	// Make sure all contract spec 0 record specs are iterated over
	contractSpec0SpecIDsIterated := []types.MetadataAddress{}
	errContractSpec0 := s.app.MetadataKeeper.IterateRecordSpecsForContractSpec(s.ctx, contractSpecs[0].SpecificationId, func(specID types.MetadataAddress) (stop bool) {
		contractSpec0SpecIDsIterated = append(contractSpec0SpecIDsIterated, specID)
		return false
	})
	s.Nil(errContractSpec0, "contract spec 0: should not have returned an error")
	s.Equal(3, len(contractSpec0SpecIDsIterated), "contract spec 0: iteration count")
	s.True(areEquivalentSetsOfMetaAddresses(contractSpec0RecSpecIDs, contractSpec0SpecIDsIterated),
		"contract spec 0: iterated over unexpected record specs:\n  expected: %v\n    actual: %v\n",
		contractSpec0RecSpecIDs, contractSpec0SpecIDsIterated)

	// Make sure all contract spec 1 record specs are iterated over
	contractSpec1SpecIDsIterated := []types.MetadataAddress{}
	errContractSpec1 := s.app.MetadataKeeper.IterateRecordSpecsForContractSpec(s.ctx, contractSpecs[1].SpecificationId, func(specID types.MetadataAddress) (stop bool) {
		contractSpec1SpecIDsIterated = append(contractSpec1SpecIDsIterated, specID)
		return false
	})
	s.Nil(errContractSpec1, "contract spec 1: should not have returned an error")
	s.Equal(2, len(contractSpec1SpecIDsIterated), "contract spec 1: iteration count")
	s.True(areEquivalentSetsOfMetaAddresses(contractSpec1RecSpecIDs, contractSpec1SpecIDsIterated),
		"contract spec 1: iterated over unexpected record specs:\n  expected: %v\n    actual: %v\n",
		contractSpec1RecSpecIDs, contractSpec1SpecIDsIterated)

	// Make sure no contract spec 2 record specs are iterated over
	contractSpec2SpecIDsIterated := []types.MetadataAddress{}
	errContractSpec2 := s.app.MetadataKeeper.IterateRecordSpecsForContractSpec(s.ctx, contractSpecs[2].SpecificationId, func(specID types.MetadataAddress) (stop bool) {
		contractSpec2SpecIDsIterated = append(contractSpec2SpecIDsIterated, specID)
		return false
	})
	s.Nil(errContractSpec2, "contract spec 2: should not have returned an error")
	s.Equal(0, len(contractSpec2SpecIDsIterated), "contract spec 2: iteration count")
	s.True(areEquivalentSetsOfMetaAddresses(contractSpec2RecSpecIDs, contractSpec2SpecIDsIterated),
		"contract spec 2: iterated over unexpected record specs:\n  expected: %v\n    actual: %v\n",
		contractSpec2RecSpecIDs, contractSpec2SpecIDsIterated)

	// Make sure an unknown contract spec results in zero iterations.
	unknownContractSpecID := types.ContractSpecMetadataAddress(uuid.New())
	unknownContractSpecIDCount := 0
	errUnknownContractSpecID := s.app.MetadataKeeper.IterateRecordSpecsForContractSpec(s.ctx, unknownContractSpecID, func(specID types.MetadataAddress) (stop bool) {
		unknownContractSpecIDCount++
		return false
	})
	s.Nil(errUnknownContractSpecID, "unknown contract spec id: should not have returned an error")
	s.Equal(0, unknownContractSpecIDCount, "unknown contract spec id: iteration count")

	// Make sure the stop bool is being recognized.
	countStop := 0
	errStop := s.app.MetadataKeeper.IterateRecordSpecsForContractSpec(s.ctx, contractSpecs[0].SpecificationId, func(specID types.MetadataAddress) (stop bool) {
		countStop++
		if countStop == 2 {
			return true
		}
		return false
	})
	s.Nil(errStop, "stop bool: should not have returned an error")
	s.Equal(2, countStop, "stop bool: iteration count")
}

func (s *SpecKeeperTestSuite) TestGetRecordSpecificationsForContractSpecificationID() {
	// Create 3 contract specs.
	contractSpecs := []*types.ContractSpecification{
		types.NewContractSpecification(
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
		),
		types.NewContractSpecification(
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
		),
		types.NewContractSpecification(
			types.ContractSpecMetadataAddress(uuid.New()),
			types.NewDescription(
				"TestIterateContractSpecsForOwner[2]",
				"A description for a unit test contract specification - owner: user1, user2",
				"http://test.net",
				"http://test.net/ico.png",
			),
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
			types.NewContractSpecificationSourceHash("somehash2"),
			"someclass_2",
		),
	}
	for _, spec := range contractSpecs {
		s.app.MetadataKeeper.SetContractSpecification(s.ctx, *spec)
	}

	// Create 3 record specs for the 1st contract spec, 2 for the 2nd, and none for the 3rd.
	recordSpecs := []*types.RecordSpecification{
		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[0].SpecificationId, "contractspec0recspec0"),
			"contractspec0recspec0",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec0", "inputspectypename0",
					types.NewInputSpecificationSourceHash("sourcehash0"),
				),
			},
			"typename0",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),
		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[0].SpecificationId, "contractspec0recspec1"),
			"contractspec0recspec1",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec1", "inputspectypename1",
					types.NewInputSpecificationSourceHash("sourcehash1"),
				),
			},
			"typename1",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),
		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[0].SpecificationId, "contractspec1recspec2"),
			"contractspec0recspec2",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec2", "inputspectypename2",
					types.NewInputSpecificationSourceHash("sourcehash2"),
				),
			},
			"typename2",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),

		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[1].SpecificationId, "contractspec1recspec3"),
			"contractspec1recspec3",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec3", "inputspectypename3",
					types.NewInputSpecificationSourceHash("sourcehash3"),
				),
			},
			"typename3",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),
		types.NewRecordSpecification(
			asRecSpecAddrOrPanic(contractSpecs[1].SpecificationId, "contractspec1recspec4"),
			"contractspec1recspec4",
			[]*types.InputSpecification{
				types.NewInputSpecification(
					"inputspec4", "inputspectypename4",
					types.NewInputSpecificationSourceHash("sourcehash4"),
				),
			},
			"typename4",
			types.DefinitionType_DEFINITION_TYPE_RECORD,
			[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		),
	}
	for _, spec := range recordSpecs {
		s.app.MetadataKeeper.SetRecordSpecification(s.ctx, *spec)
	}

	contractSpec0Expected := recordSpecs[0:3]
	contractSpec1Expected := recordSpecs[3:5]
	contractSpec2Expected := []*types.RecordSpecification{}

	// Make sure all contract spec 0 record specs are returned
	contractSpecs0Actual, contractSpecs0ActualErr := s.app.MetadataKeeper.GetRecordSpecificationsForContractSpecificationID(s.ctx, contractSpecs[0].SpecificationId)
	s.Nil(contractSpecs0ActualErr, "contract spec 0: should not have returned an error")
	s.True(areEquivalentSetsOfRecSpecs(contractSpec0Expected, contractSpecs0Actual),
		"contract spec 0: unexpected record specs:\n  expected: %v\n    actual: %v\n",
		contractSpec0Expected, contractSpecs0Actual)

	// Make sure all contract spec 1 record specs are returned
	contractSpecs1Actual, contractSpecs1ActualErr := s.app.MetadataKeeper.GetRecordSpecificationsForContractSpecificationID(s.ctx, contractSpecs[1].SpecificationId)
	s.Nil(contractSpecs1ActualErr, "contract spec 1: should not have returned an error")
	s.True(areEquivalentSetsOfRecSpecs(contractSpec1Expected, contractSpecs1Actual),
		"contract spec 1: unexpected record specs:\n  expected: %v\n    actual: %v\n",
		contractSpec1Expected, contractSpecs1Actual)

	// Make sure all contract spec 2 record specs are returned (none)
	contractSpecs2Actual, contractSpecs2ActualErr := s.app.MetadataKeeper.GetRecordSpecificationsForContractSpecificationID(s.ctx, contractSpecs[2].SpecificationId)
	s.Nil(contractSpecs2ActualErr, "contract spec 2: should not have returned an error")
	s.True(areEquivalentSetsOfRecSpecs(contractSpec2Expected, contractSpecs2Actual),
		"contract spec 2: unexpected record specs:\n  expected: %v\n    actual: %v\n",
		contractSpec2Expected, contractSpecs2Actual)

	// Make sure an unknown contract spec returns empty.
	unknownContractSpecID := types.ContractSpecMetadataAddress(uuid.New())
	unknownContractSpecIDActual, unknownContractSpecIDActualErr := s.app.MetadataKeeper.GetRecordSpecificationsForContractSpecificationID(s.ctx, unknownContractSpecID)
	s.Nil(unknownContractSpecIDActualErr, "unknown contract spec id: should not have returned an error")
	s.Equal(0, len(unknownContractSpecIDActual), "unknown contract spec id: count")
}

func (s *SpecKeeperTestSuite) TestValidateRecordSpecUpdate() {
	contractSpecUUIDOther := uuid.New()
	tests := []struct {
		name     string
		existing *types.RecordSpecification
		proposed *types.RecordSpecification
		want     string
	}{
		{
			"validate basic called on proposed",
			nil,
			types.NewRecordSpecification(
				types.RecordSpecMetadataAddress(s.contractSpecUUID1, "name"),
				"name",
				[]*types.InputSpecification{},
				"",
				types.DefinitionType_DEFINITION_TYPE_RECORD,
				[]types.PartyType{types.PartyType_PARTY_TYPE_SERVICER},
			),
			"record specification type name cannot be empty",
		},
		{
			"validate basic not called on existing",
			types.NewRecordSpecification(
				types.RecordSpecMetadataAddress(s.contractSpecUUID1, "name"),
				"name",
				[]*types.InputSpecification{},
				"", // should cause error if ValidateBasic called on it
				types.DefinitionType_DEFINITION_TYPE_RECORD,
				[]types.PartyType{types.PartyType_PARTY_TYPE_SERVICER},
			),
			types.NewRecordSpecification(
				types.RecordSpecMetadataAddress(s.contractSpecUUID1, "name"),
				"name",
				[]*types.InputSpecification{},
				"typename",
				types.DefinitionType_DEFINITION_TYPE_RECORD,
				[]types.PartyType{types.PartyType_PARTY_TYPE_SERVICER},
			),
			"",
		},
		{
			"SpecificationIDs must match",
			types.NewRecordSpecification(
				types.RecordSpecMetadataAddress(s.contractSpecUUID1, "foo"),
				"foo",
				[]*types.InputSpecification{},
				"", // should cause error if ValidateBasic called on it
				types.DefinitionType_DEFINITION_TYPE_RECORD,
				[]types.PartyType{types.PartyType_PARTY_TYPE_SERVICER},
			),
			types.NewRecordSpecification(
				types.RecordSpecMetadataAddress(contractSpecUUIDOther, "foo"),
				"foo",
				[]*types.InputSpecification{},
				"typename",
				types.DefinitionType_DEFINITION_TYPE_RECORD,
				[]types.PartyType{types.PartyType_PARTY_TYPE_SERVICER},
			),
			fmt.Sprintf("cannot update record spec identifier. expected %s, got %s",
				types.RecordSpecMetadataAddress(s.contractSpecUUID1, "foo"),
				types.RecordSpecMetadataAddress(contractSpecUUIDOther, "foo")),
		},
		// Names must match - cannot be tested. A changed name will change the spec id.
		// So either ValidateBasic will catch that the hashed name doesn't match its part in the ID,
		// or the ValidateRecordSpecUpdate will catch the changing specification id.
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateRecordSpecUpdate(s.ctx, tt.existing, *tt.proposed)
			if err != nil {
				require.Equal(t, tt.want, err.Error(), "RecordSpec Keeper ValidateRecordSpecUpdate error")
			} else if len(tt.want) > 0 {
				t.Errorf("RecordSpec Keeper ValidateRecordSpecUpdate error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *SpecKeeperTestSuite) TestGetSetRemoveContractSpecification() {
	newSpec := types.NewContractSpecification(
		s.contractSpecID1,
		types.NewDescription(
			"TestGetSetRemoveContractSpecification",
			"A description for a unit test contract specification",
			"http://test.net",
			"http://test.net/ico.png",
		),
		[]string{s.user1Addr.String()},
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		types.NewContractSpecificationSourceHash("somehash"),
		"someclass",
	)
	require.NotNil(s.T(), newSpec, "test setup failure: NewContractSpecification should not return nil")

	spec1, found1 := s.app.MetadataKeeper.GetContractSpecification(s.ctx, s.contractSpecID1)
	s.False(found1, "1: get contract spec should return false before it has been saved")
	s.NotNil(spec1, "1: get contract spec should always return a non-nil contract spec")

	s.app.MetadataKeeper.SetContractSpecification(s.ctx, *newSpec)

	spec2, found2 := s.app.MetadataKeeper.GetContractSpecification(s.ctx, s.contractSpecID1)
	s.True(found2, "get contract spec should return true after it has been saved")
	s.NotNil(spec2, "get contract spec should always return a non-nil contract spec")
	s.Equal(s.contractSpecID1, spec2.SpecificationId, "2: get contract spec should return a spec containing id provided")

	spec3, found3 := s.app.MetadataKeeper.GetContractSpecification(s.ctx, types.ContractSpecMetadataAddress(uuid.New()))
	s.False(found3, "3: get contract spec should return false for an unknown address")
	s.NotNil(spec3, "3: get contract spec should always return a non-nil contract spec")

	remErr := s.app.MetadataKeeper.RemoveContractSpecification(s.ctx, newSpec.SpecificationId)
	s.Nil(remErr, "4: remove should not return any error")

	spec5, found5 := s.app.MetadataKeeper.GetContractSpecification(s.ctx, s.contractSpecID1)
	s.False(found5, "5: get contract spec should return false after it has been removed")
	s.NotNil(spec5, "5: get contract spec should always return a non-nil contract spec")

	remErr2 := s.app.MetadataKeeper.RemoveContractSpecification(s.ctx, s.contractSpecID1)
	s.Equal(
		fmt.Errorf("contract specification with id %s not found", s.contractSpecID1),
		remErr2,
		"6: remove error message when not found",
	)
}

func (s *SpecKeeperTestSuite) TestIterateContractSpecs() {
	size := 10
	specs := make([]*types.ContractSpecification, size)
	for i := 0; i < size; i++ {
		specs[i] = types.NewContractSpecification(
			types.ContractSpecMetadataAddress(uuid.New()),
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
		)
		s.app.MetadataKeeper.SetContractSpecification(s.ctx, *specs[i])
	}

	visitedContractSpecIDs := make([]types.MetadataAddress, size)
	count := 0
	err1 := s.app.MetadataKeeper.IterateContractSpecs(s.ctx, func(spec types.ContractSpecification) (stop bool) {
		if containsMetadataAddress(visitedContractSpecIDs, spec.SpecificationId) {
			s.Fail("function IterateContractSpecs visited the same contract specification twice: %s", spec.SpecificationId.String())
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
	)
	user1SpecIDs[2] = specs[4].SpecificationId
	user2SpecIDs[2] = specs[4].SpecificationId

	for _, spec := range specs {
		s.app.MetadataKeeper.SetContractSpecification(s.ctx, *spec)
	}

	// Make sure all user1 contract specs are iterated over
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

	// Make sure all user2 contract specs are iterated over
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
	otherContractSpecID := types.ContractSpecMetadataAddress(uuid.New())
	tests := []struct {
		name     string
		existing *types.ContractSpecification
		proposed *types.ContractSpecification
		want     string
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
			),
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
			),
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
			),
			"",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateContractSpecUpdate(s.ctx, tt.existing, *tt.proposed)
			if err != nil {
				require.Equal(t, tt.want, err.Error(), "ContractSpec Keeper ValidateContractSpecUpdate error")
			} else if len(tt.want) > 0 {
				t.Errorf("ContractSpec Keeper ValidateContractSpecUpdate error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *SpecKeeperTestSuite) TestContractSpecIndexing() {
	specID := types.ContractSpecMetadataAddress(uuid.New())

	// randomUser defined in scope_test.go
	ownerConstant := randomUser()
	ownerToAdd := randomUser()
	ownerToRemove := randomUser()

	specV1 := types.ContractSpecification{
		SpecificationId: specID,
		Description:     nil,
		OwnerAddresses:  []string{ownerConstant.Bech32, ownerToRemove.Bech32},
		PartiesInvolved: nil,
		Source:          nil,
		ClassName:       "",
	}
	specV2 := types.ContractSpecification{
		SpecificationId: specID,
		Description:     nil,
		OwnerAddresses:  []string{ownerConstant.Bech32, ownerToAdd.Bech32},
		PartiesInvolved: nil,
		Source:          nil,
		ClassName:       "",
	}

	store := s.ctx.KVStore(s.app.GetKey(types.ModuleName))

	s.T().Run("1 write new contract specification", func(t *testing.T) {
		expectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressContractSpecCacheKey(ownerConstant.Addr, specID), "ownerConstant address index"},
			{types.GetAddressContractSpecCacheKey(ownerToRemove.Addr, specID), "ownerToRemove address index"},
		}

		s.app.MetadataKeeper.SetContractSpecification(s.ctx, specV1)

		for _, expected := range expectedIndexes {
			assert.True(t, store.Has(expected.key), expected.name)
		}
	})

	s.T().Run("2 update contract specification", func(t *testing.T) {
		expectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressContractSpecCacheKey(ownerConstant.Addr, specID), "ownerConstant address index"},
			{types.GetAddressContractSpecCacheKey(ownerToAdd.Addr, specID), "ownerToAdd address index"},
		}
		unexpectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressContractSpecCacheKey(ownerToRemove.Addr, specID), "ownerToRemove address index"},
		}

		s.app.MetadataKeeper.SetContractSpecification(s.ctx, specV2)

		for _, expected := range expectedIndexes {
			assert.True(t, store.Has(expected.key), expected.name)
		}
		for _, unexpected := range unexpectedIndexes {
			assert.False(t, store.Has(unexpected.key), unexpected.name)
		}
	})

	s.T().Run("3 delete contract specification", func(t *testing.T) {
		unexpectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressContractSpecCacheKey(ownerConstant.Addr, specID), "ownerConstant address index"},
			{types.GetAddressContractSpecCacheKey(ownerToAdd.Addr, specID), "ownerToAdd address index"},
			{types.GetAddressContractSpecCacheKey(ownerToRemove.Addr, specID), "ownerToRemove address index"},
		}

		assert.NoError(t, s.app.MetadataKeeper.RemoveContractSpecification(s.ctx, specID), "removing contract spec")

		for _, unexpected := range unexpectedIndexes {
			assert.False(t, store.Has(unexpected.key), unexpected.name)
		}
	})
}

func (s *SpecKeeperTestSuite) TestGetSetRemoveScopeSpecification() {
	newSpec := types.NewScopeSpecification(
		s.scopeSpecID,
		types.NewDescription(
			"TestGetSetRemoveScopeSpecification",
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

	remErr := s.app.MetadataKeeper.RemoveScopeSpecification(s.ctx, newSpec.SpecificationId)
	s.Nil(remErr, "4: remove should not return any error")

	spec5, found5 := s.app.MetadataKeeper.GetScopeSpecification(s.ctx, s.scopeSpecID)
	s.False(found5, "5: get scope spec should return false after it has been removed")
	s.NotNil(spec5, "5: get scope spec should always return a non-nil scope spec")

	remErr2 := s.app.MetadataKeeper.RemoveScopeSpecification(s.ctx, s.scopeSpecID)
	s.Equal(
		fmt.Errorf("scope specification with id %s not found", s.scopeSpecID),
		remErr2,
		"6: remove error message when not found",
	)
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
		name     string
		existing *types.ScopeSpecification
		proposed *types.ScopeSpecification
		want     string
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
			"",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateScopeSpecUpdate(s.ctx, tt.existing, *tt.proposed)
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

func (s *SpecKeeperTestSuite) TestScopeSpecIndexing() {
	specID := types.ScopeSpecMetadataAddress(uuid.New())

	// randomUser defined in scope_test.go
	ownerConstant := randomUser()
	ownerToAdd := randomUser()
	ownerToRemove := randomUser()

	cSpecIDConstant := types.ContractSpecMetadataAddress(uuid.New())
	cSpecIDToAdd := types.ContractSpecMetadataAddress(uuid.New())
	cSpecIDToRemove := types.ContractSpecMetadataAddress(uuid.New())

	specV1 := types.ScopeSpecification{
		SpecificationId: specID,
		Description:     nil,
		OwnerAddresses:  []string{ownerConstant.Bech32, ownerToRemove.Bech32},
		PartiesInvolved: nil,
		ContractSpecIds: []types.MetadataAddress{cSpecIDConstant, cSpecIDToRemove},
	}
	specV2 := types.ScopeSpecification{
		SpecificationId: specID,
		Description:     nil,
		OwnerAddresses:  []string{ownerConstant.Bech32, ownerToAdd.Bech32},
		PartiesInvolved: nil,
		ContractSpecIds: []types.MetadataAddress{cSpecIDConstant, cSpecIDToAdd},
	}

	store := s.ctx.KVStore(s.app.GetKey(types.ModuleName))

	s.T().Run("1 write new scope specification", func(t *testing.T) {
		expectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeSpecCacheKey(ownerConstant.Addr, specID), "ownerConstant address index"},
			{types.GetAddressScopeSpecCacheKey(ownerToRemove.Addr, specID), "ownerToRemove address index"},

			{types.GetContractSpecScopeSpecCacheKey(cSpecIDConstant, specID), "cSpecIDConstant contract spec index"},
			{types.GetContractSpecScopeSpecCacheKey(cSpecIDToRemove, specID), "cSpecIDToRemove contract spec index"},
		}

		s.app.MetadataKeeper.SetScopeSpecification(s.ctx, specV1)

		for _, expected := range expectedIndexes {
			assert.True(t, store.Has(expected.key), expected.name)
		}
	})

	s.T().Run("2 update scope specification", func(t *testing.T) {
		expectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeSpecCacheKey(ownerConstant.Addr, specID), "ownerConstant address index"},
			{types.GetAddressScopeSpecCacheKey(ownerToAdd.Addr, specID), "ownerToAdd address index"},

			{types.GetContractSpecScopeSpecCacheKey(cSpecIDConstant, specID), "cSpecIDConstant contract spec index"},
			{types.GetContractSpecScopeSpecCacheKey(cSpecIDToAdd, specID), "cSpecIDToAdd contract spec index"},
		}
		unexpectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeSpecCacheKey(ownerToRemove.Addr, specID), "ownerToRemove address index"},

			{types.GetContractSpecScopeSpecCacheKey(cSpecIDToRemove, specID), "cSpecIDToRemove contract spec index"},
		}

		s.app.MetadataKeeper.SetScopeSpecification(s.ctx, specV2)

		for _, expected := range expectedIndexes {
			assert.True(t, store.Has(expected.key), expected.name)
		}
		for _, unexpected := range unexpectedIndexes {
			assert.False(t, store.Has(unexpected.key), unexpected.name)
		}
	})

	s.T().Run("3 delete scope specification", func(t *testing.T) {
		unexpectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeSpecCacheKey(ownerConstant.Addr, specID), "ownerConstant address index"},
			{types.GetAddressScopeSpecCacheKey(ownerToAdd.Addr, specID), "ownerToAdd address index"},
			{types.GetAddressScopeSpecCacheKey(ownerToRemove.Addr, specID), "ownerToRemove address index"},

			{types.GetContractSpecScopeSpecCacheKey(cSpecIDConstant, specID), "cSpecIDConstant contract spec index"},
			{types.GetContractSpecScopeSpecCacheKey(cSpecIDToAdd, specID), "cSpecIDToAdd contract spec index"},
			{types.GetContractSpecScopeSpecCacheKey(cSpecIDToRemove, specID), "cSpecIDToRemove contract spec index"},
		}

		assert.NoError(t, s.app.MetadataKeeper.RemoveScopeSpecification(s.ctx, specID), "removing scope spec")

		for _, unexpected := range unexpectedIndexes {
			assert.False(t, store.Has(unexpected.key), unexpected.name)
		}
	})
}
