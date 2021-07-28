package cli_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/testutil"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
)

type IntegrationCLIPageTestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network

	asJson string

	accountKey     *secp256k1.PrivKey
	accountAddr    sdk.AccAddress
	accountAddrStr string

	user1Key     *secp256k1.PrivKey
	user1Addr    sdk.AccAddress
	user1AddrStr string

	scopeSpecCount    int
	contractSpecCount int
	recordSpecCount   int
	scopeCount        int
	sessionCount      int
	recordCount       int
	osLocatorCount    int

	accountScopesOwned      int
	user1ScopesOwned        int
	accountScopesValueOwned int
	user1ScopesValueOwned   int
}

func TestIntegrationCLIPageTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationCLIPageTestSuite))
}

func (s *IntegrationCLIPageTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	var err error

	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.NumValidators = 1

	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("account"))
	s.accountAddr, err = sdk.AccAddressFromHex(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err, "getting accountAddr from s.accountKey")
	s.accountAddrStr = s.accountAddr.String()

	s.user1Key = secp256k1.GenPrivKeyFromSecret([]byte("user1"))
	s.user1Addr, err = sdk.AccAddressFromHex(s.user1Key.PubKey().Address().String())
	s.Require().NoError(err, "getting user1Addr from s.user1Key")
	s.user1AddrStr = s.user1Addr.String()

	s.asJson = fmt.Sprintf("--%s=json", tmcli.OutputFlag)

	var metadataData metadatatypes.GenesisState
	s.Require().NoError(s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[metadatatypes.ModuleName], &metadataData), "unmarshalling JSON metadataData")

	// Create 20 contract specifications, each with 2 record specifications.
	cSpecs := map[string]*metadatatypes.ContractSpecification{}
	rSpecs := map[string][]*metadatatypes.RecordSpecification{}
	for i := 0; i < 20; i++ {
		written := toWritten(i)
		cSpec := metadatatypes.ContractSpecification{
			SpecificationId: metadatatypes.ContractSpecMetadataAddress(uuid.New()),
			Description: &metadatatypes.Description{
				Name:        fmt.Sprintf("Contract Spec %d", i),
				Description: fmt.Sprintf("The contract specification with number %s", written),
				WebsiteUrl:  "",
				IconUrl:     "",
			},
			OwnerAddresses:  []string{s.accountAddrStr},
			PartiesInvolved: []metadatatypes.PartyType{metadatatypes.PartyType_PARTY_TYPE_OWNER},
			Source:          metadatatypes.NewContractSpecificationSourceHash(written),
			ClassName:       toClassName(written),
		}
		rSpec1 := metadatatypes.RecordSpecification{
			SpecificationId: nil,
			Name:            fmt.Sprintf("record %s 1 of 2", written),
			Inputs: []*metadatatypes.InputSpecification{
				{
					Name:     written + "-1",
					TypeName: "string",
					Source:   metadatatypes.NewInputSpecificationSourceHash(written),
				},
			},
			TypeName:           toClassName(toWritten(100 + i)),
			ResultType:         metadatatypes.DefinitionType_DEFINITION_TYPE_RECORD_LIST,
			ResponsibleParties: []metadatatypes.PartyType{metadatatypes.PartyType_PARTY_TYPE_OWNER},
		}
		rSpec1.SpecificationId = cSpec.SpecificationId.MustGetAsRecordSpecAddress(rSpec1.Name)
		rSpec2 := metadatatypes.RecordSpecification{
			SpecificationId: nil,
			Name:            fmt.Sprintf("record %s 2 of 2", written),
			Inputs: []*metadatatypes.InputSpecification{
				{
					Name:     written + "-2",
					TypeName: "string",
					Source:   metadatatypes.NewInputSpecificationSourceHash(written),
				},
			},
			TypeName:           toClassName(toWritten(100 + i)),
			ResultType:         metadatatypes.DefinitionType_DEFINITION_TYPE_RECORD_LIST,
			ResponsibleParties: []metadatatypes.PartyType{metadatatypes.PartyType_PARTY_TYPE_OWNER},
		}
		rSpec2.SpecificationId = cSpec.SpecificationId.MustGetAsRecordSpecAddress(rSpec2.Name)
		metadataData.ContractSpecifications = append(metadataData.ContractSpecifications, cSpec)
		metadataData.RecordSpecifications = append(metadataData.RecordSpecifications, rSpec1, rSpec2)
		cSpecs[cSpec.SpecificationId.String()] = &cSpec
		rSpecs[cSpec.SpecificationId.String()] = []*metadatatypes.RecordSpecification{&rSpec1, &rSpec2}
	}
	// Create 20 scope specifications, each with two contract specifications.
	for i := 0; i < 20; i++ {
		written := toWritten(i)
		sSpec := metadatatypes.ScopeSpecification{
			SpecificationId: metadatatypes.ScopeSpecMetadataAddress(uuid.New()),
			Description: &metadatatypes.Description{
				Name:        fmt.Sprintf("Scope Spec %d", i),
				Description: fmt.Sprintf("The scope specification with number %s", written),
				WebsiteUrl:  "",
				IconUrl:     "",
			},
			OwnerAddresses:  []string{s.accountAddrStr},
			PartiesInvolved: []metadatatypes.PartyType{metadatatypes.PartyType_PARTY_TYPE_OWNER},
			ContractSpecIds: []metadatatypes.MetadataAddress{
				metadataData.ContractSpecifications[i].SpecificationId,
				metadataData.ContractSpecifications[(i+1)%20].SpecificationId,
			},
		}
		metadataData.ScopeSpecifications = append(metadataData.ScopeSpecifications, sSpec)
	}
	// Create 100 scopes, each with 2 sessions, and each session with 2 records.
	// Scopes:
	//    Use scope specification i % 20.
	//    i % 5 == 0: 20 of them are owned by s.accountAddr and value owner is s.accountAddr.
	//    i % 5 == 1: 20 of them are owned by s.accountAddr and value owner is s.user1Addr.
	//    i % 5 == 2: 20 of them are owned by s.accountAddr + s.user1Addr and value owner is s.accountAddr
	//    i % 5 == 3: 20 of them are owned by s.accountAddr + s.user1Addr and value owner is s.user1Addr.
	//    i % 5 == 4: 20 of them are owned by s.user1Addr and value owner is s.user1Addr.
	//        Result:
	//            s.user1Addr is owner of 60.
	//            s.accountAddr is owner of 80
	//            s.user1Addr is value owner of 60
	//            s.accountAddr is value owner of 40
	// Sessions:
	//    Use each c spec on the scope spec
	// Records:
	//    Use each record spec in the contract spec
	s.user1ScopesOwned = 60
	s.accountScopesOwned = 80
	s.user1ScopesValueOwned = 60
	s.accountScopesValueOwned = 40
	accountOwnerParty := metadatatypes.Party{Address: s.accountAddrStr, Role: metadatatypes.PartyType_PARTY_TYPE_OWNER}
	user1OwnerParty := metadatatypes.Party{Address: s.accountAddrStr, Role: metadatatypes.PartyType_PARTY_TYPE_OWNER}
	for si := 0; si < 100; si++ {
		scopeSpec := metadataData.ScopeSpecifications[si%20]
		records := []metadatatypes.Record{}
		sessions := []metadatatypes.Session{}
		scope := metadatatypes.Scope{
			ScopeId:           metadatatypes.ScopeMetadataAddress(uuid.New()),
			SpecificationId:   scopeSpec.SpecificationId,
			DataAccess:        []string{s.accountAddrStr, s.user1AddrStr},
			Owners:            []metadatatypes.Party{},
			ValueOwnerAddress: "",
		}
		switch si % 5 {
		case 0:
			scope.Owners = append(scope.Owners, accountOwnerParty)
			scope.ValueOwnerAddress = s.accountAddrStr
		case 1:
			scope.Owners = append(scope.Owners, accountOwnerParty)
			scope.ValueOwnerAddress = s.user1AddrStr
		case 2:
			scope.Owners = append(scope.Owners, accountOwnerParty, user1OwnerParty)
			scope.ValueOwnerAddress = s.accountAddrStr
		case 3:
			scope.Owners = append(scope.Owners, accountOwnerParty, user1OwnerParty)
			scope.ValueOwnerAddress = s.user1AddrStr
		case 4:
			scope.Owners = append(scope.Owners, user1OwnerParty)
			scope.ValueOwnerAddress = s.user1AddrStr
		}
		for ci, cSpecID := range scopeSpec.ContractSpecIds {
			session := metadatatypes.Session{
				SessionId:       scope.ScopeId.MustGetAsSessionAddress(uuid.New()),
				SpecificationId: cSpecs[cSpecID.String()].SpecificationId,
				Parties:         scope.Owners,
				Name:            cSpecs[cSpecID.String()].ClassName,
				Context:         []byte(toWritten(ci)),
				Audit: &metadatatypes.AuditFields{
					CreatedDate: time.Now(),
					CreatedBy:   s.accountAddrStr,
					Version:     1,
					Message:     "initial state",
				},
			}
			sessions = append(sessions, session)
			for ri, rSpec := range rSpecs[cSpecID.String()] {
				allWritten := fmt.Sprintf("%s%s%s", toWritten(si), strings.ToTitle(toWritten(ci)), strings.ToTitle(toWritten(ri)))
				records = append(records, metadatatypes.Record{
					SpecificationId: rSpec.SpecificationId,
					Name:            rSpec.Name,
					SessionId:       session.SessionId,
					Process: metadatatypes.Process{
						ProcessId: &metadatatypes.Process_Hash{Hash: allWritten},
						Name:      toClassName(allWritten),
						Method:    allWritten,
					},
					Inputs: []metadatatypes.RecordInput{
						{
							Name:     rSpec.Inputs[0].Name,
							Source:   &metadatatypes.RecordInput_Hash{Hash: rSpec.Inputs[0].GetHash()},
							TypeName: rSpec.Inputs[0].TypeName,
							Status:   metadatatypes.RecordInputStatus_Record,
						},
					},
					Outputs: []metadatatypes.RecordOutput{
						{
							Hash:   allWritten + "Out",
							Status: metadatatypes.ResultStatus_RESULT_STATUS_PASS,
						},
					},
				})
			}
		}
		metadataData.Scopes = append(metadataData.Scopes, scope)
		metadataData.Sessions = append(metadataData.Sessions, sessions...)
		metadataData.Records = append(metadataData.Records, records...)
	}

	// Create 60 Object Store Locators
	// i % 3 == 0 or 1: 40 of them with a unique URI
	// i % 3 == 2: 20 of them with the URI https://provenance.io/
	for i := 0; i < 60; i++ {
		addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
		osl := metadatatypes.ObjectStoreLocator{
			Owner:         addr,
			EncryptionKey: addr,
			LocatorUri:    "",
		}
		switch i % 3 {
		case 0, 1:
			osl.LocatorUri = fmt.Sprintf("http://%s.corn", toClassName(toWritten(i)))
		case 2:
			osl.LocatorUri = "https://provenance.io/"
		}
		metadataData.ObjectStoreLocators = append(metadataData.ObjectStoreLocators, osl)
	}

	metadataDataBz, err := s.cfg.Codec.MarshalJSON(&metadataData)
	s.Require().NoError(err, "marshalling JSON metadataData")
	s.cfg.GenesisState[metadatatypes.ModuleName] = metadataDataBz

	s.scopeSpecCount = len(metadataData.ScopeSpecifications)
	s.contractSpecCount = len(metadataData.ContractSpecifications)
	s.recordSpecCount = len(metadataData.RecordSpecifications)
	s.scopeCount = len(metadataData.Scopes)
	s.sessionCount = len(metadataData.Sessions)
	s.recordCount = len(metadataData.Records)
	s.osLocatorCount = len(metadataData.ObjectStoreLocators)

	var authData authtypes.GenesisState
	s.Require().NoError(s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[authtypes.ModuleName], &authData), "unmarshalling JSON authData")
	genAccount, err := codectypes.NewAnyWithValue(authtypes.NewBaseAccount(s.accountAddr, s.accountKey.PubKey(), 1, 1))
	s.Require().NoError(err, "creating Any BaseAccount for s.accountKey")
	authData.Accounts = append(authData.Accounts, genAccount)

	authDataBz, err := s.cfg.Codec.MarshalJSON(&authData)
	s.Require().NoError(err, "marshalling json authData")
	s.cfg.GenesisState[authtypes.ModuleName] = authDataBz

	s.testnet = testnet.New(s.T(), s.cfg)

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err, "calling s.testnet.WaitForHeight(1)")
	s.Require().NoError(s.testnet.Validators[0].ClientCtx.Keyring.ImportPrivKey(s.accountAddr.String(), crypto.EncryptArmorPrivKey(s.accountKey, "pasSword0", "secp256k1"), "pasSword0"), "adding s.accountKey to keyring")
	s.T().Log("done setting up integration test suite")
}

func (s *IntegrationCLIPageTestSuite) TearDownSuite() {
	s.T().Log("teardown waiting for next block")
	s.Require().NoError(s.testnet.WaitForNextBlock(), "waiting for next block")
	s.T().Log("teardown cleaning up testnet")
	s.testnet.Cleanup()
	s.T().Log("teardown done")
}

// Converts an integer to a written version of it. E.g. 1 => one, 83 => eightyThree.
func toWritten(i int) string {
	if i > 999999 {
		panic("cannot convert number larger than 999,999 to written string")
	}
	if i < 0 {
		return "nagative" + strings.Title(toWritten(-1*i))
	}
	switch i {
	case 0:
		return "zero"
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	case 4:
		return "four"
	case 5:
		return "five"
	case 6:
		return "six"
	case 7:
		return "seven"
	case 8:
		return "eight"
	case 9:
		return "nine"
	case 10:
		return "ten"
	case 11:
		return "eleven"
	case 12:
		return "twelve"
	case 13:
		return "thirteen"
	case 14:
		return "fourteen"
	case 15:
		return "fifteen"
	case 16:
		return "sixteen"
	case 17:
		return "seventeen"
	case 18:
		return "eighteen"
	case 19:
		return "nineteen"
	case 20:
		return "twenty"
	case 30:
		return "thirty"
	case 40:
		return "forty"
	case 50:
		return "fifty"
	case 60:
		return "sixty"
	case 70:
		return "seventy"
	case 80:
		return "eighty"
	case 90:
		return "ninety"
	default:
		var r int
		var l string
		switch {
		case i < 100:
			r = i % 10
			l = toWritten(i - r)
		case i < 1000:
			r = i % 100
			l = toWritten(i/100) + "Hundred"
		default:
			r = i % 1000
			l = toWritten(i/1000) + "Thousand"
		}
		if r == 0 {
			return l
		}
		return l + strings.Title(toWritten(r))
	}
}

// Regex for finding any non period character followed by an upper-case letter.
var reUpperLetter = regexp.MustCompile(`[^.][[:upper:]]`)

// Convert a camel-case string to a class string by putting a period before each upper-case letter
// and then lower-casing the whole thing.
func toClassName(str string) string {
	return strings.ToLower(reUpperLetter.ReplaceAllStringFunc(str, func(s string) string {
		return s[0:1] + "." + s[1:]
	}))
}

// TODO: Delete this once there's actual tests in here.
func (s *IntegrationCLIPageTestSuite) TestSetupTeardown() {
	s.Assert().True(true, "this should be true")
}

// TODO: ScopesAllRequest - outputScopesAll: GetMetadataGetAllCmd all scopes, GetMetadataScopeCmd scope all
// TODO: SessionsAllRequest - outputSessionsAll: GetMetadataGetAllCmd all sessions, GetMetadataSessionCmd session all
// TODO: RecordsAllRequest - outputRecordsAll: GetMetadataGetAllCmd all records, GetMetadataRecordCmd record all
// TODO: ScopeSpecificationsAllRequest - outputScopeSpecsAll: GetMetadataGetAllCmd all scopespecs, GetMetadataScopeSpecCmd scopespec all
// TODO: ContractSpecificationsAllRequest - outputContractSpecsAll: GetMetadataGetAllCmd all contractspecs, GetMetadataContractSpecCmd contractspec all
// TODO: RecordSpecificationsAllRequest - outputRecordSpecsAll: GetMetadataGetAllCmd all recordspecs, GetMetadataRecordSpecCmd recordspec all
// TODO: OSLocatorsByURIRequest - outputOSLocatorsByURI: GetOSLocatorCmd locator uri
// TODO: OSAllLocatorsRequest - outputOSLocatorsAll: GetMetadataGetAllCmd all locators, GetOSLocatorCmd locator all
// TODO: OwnershipRequest - outputOwnership: GetOwnershipCmd owner address
// TODO: ValueOwnershipRequest - outputValueOwnership: GetValueOwnershipCmd valueowner address
