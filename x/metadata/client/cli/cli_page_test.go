package cli_test

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/testutil"
	mdcli "github.com/provenance-io/provenance/x/metadata/client/cli"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
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

	osLocatorURI string

	scopeSpecCount    int
	contractSpecCount int
	recordSpecCount   int
	scopeCount        int
	sessionCount      int
	recordCount       int
	osLocatorCount    int
	osLocatorURICount int

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
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 0)
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

	s.asJson = "--output=json"

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
	//            s.user1Addr is owner of 80 (either owner or value owner or data access)
	//            s.accountAddr is owner of 80 (either owner or value owner or data access)
	//            s.user1Addr is value owner of 60 (just looking at value owner field)
	//            s.accountAddr is value owner of 40 (just looking at value owner field)
	// Sessions:
	//    Use each c spec on the scope spec
	// Records:
	//    Use each record spec in the contract spec
	s.user1ScopesOwned = 80
	s.accountScopesOwned = 80
	s.user1ScopesValueOwned = 60
	s.accountScopesValueOwned = 40
	accountOwnerParty := metadatatypes.Party{Address: s.accountAddrStr, Role: metadatatypes.PartyType_PARTY_TYPE_OWNER}
	user1OwnerParty := metadatatypes.Party{Address: s.user1AddrStr, Role: metadatatypes.PartyType_PARTY_TYPE_OWNER}
	for si := 0; si < 100; si++ {
		scopeSpec := metadataData.ScopeSpecifications[si%20]
		records := []metadatatypes.Record{}
		sessions := []metadatatypes.Session{}
		scope := metadatatypes.Scope{
			ScopeId:           metadatatypes.ScopeMetadataAddress(uuid.New()),
			SpecificationId:   scopeSpec.SpecificationId,
			DataAccess:        []string{},
			Owners:            nil,
			ValueOwnerAddress: "",
		}
		switch si % 5 {
		case 0:
			scope.Owners = []metadatatypes.Party{accountOwnerParty}
			scope.ValueOwnerAddress = s.accountAddrStr
		case 1:
			scope.Owners = []metadatatypes.Party{accountOwnerParty}
			scope.ValueOwnerAddress = s.user1AddrStr
		case 2:
			scope.Owners = []metadatatypes.Party{accountOwnerParty, user1OwnerParty}
			scope.ValueOwnerAddress = s.accountAddrStr
		case 3:
			scope.Owners = []metadatatypes.Party{accountOwnerParty, user1OwnerParty}
			scope.ValueOwnerAddress = s.user1AddrStr
		case 4:
			scope.Owners = []metadatatypes.Party{user1OwnerParty}
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
	s.osLocatorURI = "https://provenance.io/"
	s.osLocatorURICount = 0
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
			osl.LocatorUri = s.osLocatorURI
			s.osLocatorURICount++
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
	testutil.CleanUp(s.testnet, s.T())
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

func limitArg(pageSize int) string {
	return fmt.Sprintf("--limit=%d", pageSize)
}

func pageKeyArg(nextKey string) string {
	return fmt.Sprintf("--page-key=%s", nextKey)
}

// Convert a camel-case string to a class string by putting a period before each upper-case letter
// and then lower-casing the whole thing.
func toClassName(str string) string {
	return strings.ToLower(reUpperLetter.ReplaceAllStringFunc(str, func(s string) string {
		return s[0:1] + "." + s[1:]
	}))
}

// scopeSorter implements sort.Interface for []metadatatypes.Scope
// Sorts by .ScopeId
type scopeSorter []metadatatypes.Scope

func (a scopeSorter) Len() int {
	return len(a)
}
func (a scopeSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a scopeSorter) Less(i, j int) bool {
	return a[i].ScopeId.Compare(a[j].ScopeId) < 0
}

// sessionSorter implements sort.Interface for []metadatatypes.Session
// Sorts by .SessionId
type sessionSorter []metadatatypes.Session

func (a sessionSorter) Len() int {
	return len(a)
}
func (a sessionSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a sessionSorter) Less(i, j int) bool {
	return a[i].SessionId.Compare(a[j].SessionId) < 0
}

// recordSorter implements sort.Interface for []metadatatypes.Record
// Sorts by .SessionId then by .Name
type recordSorter []metadatatypes.Record

func (a recordSorter) Len() int {
	return len(a)
}
func (a recordSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a recordSorter) Less(i, j int) bool {
	scmp := a[i].SessionId.Compare(a[j].SessionId)
	if scmp != 0 {
		return scmp < 0
	}
	return a[i].Name < a[j].Name
}

// scopeSpecSorter implements sort.Interface for []metadatatypes.ScopeSpecification
// Sorts by .SpecificationId
type scopeSpecSorter []metadatatypes.ScopeSpecification

func (a scopeSpecSorter) Len() int {
	return len(a)
}
func (a scopeSpecSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a scopeSpecSorter) Less(i, j int) bool {
	return a[i].SpecificationId.Compare(a[j].SpecificationId) < 0
}

// contractSpecSorter implements sort.Interface for []metadatatypes.ContractSpecification
// Sorts by .SpecificationId
type contractSpecSorter []metadatatypes.ContractSpecification

func (a contractSpecSorter) Len() int {
	return len(a)
}
func (a contractSpecSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a contractSpecSorter) Less(i, j int) bool {
	return a[i].SpecificationId.Compare(a[j].SpecificationId) < 0
}

// recordSpecSorter implements sort.Interface for []metadatatypes.RecordSpecification
// Sorts by .SpecificationId
type recordSpecSorter []metadatatypes.RecordSpecification

func (a recordSpecSorter) Len() int {
	return len(a)
}
func (a recordSpecSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a recordSpecSorter) Less(i, j int) bool {
	return a[i].SpecificationId.Compare(a[j].SpecificationId) < 0
}

// osLocatorSorter implements sort.Interface for []metadatatypes.ObjectStoreLocator
// Sorts by .Owner
type osLocatorSorter []metadatatypes.ObjectStoreLocator

func (a osLocatorSorter) Len() int {
	return len(a)
}
func (a osLocatorSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a osLocatorSorter) Less(i, j int) bool {
	return a[i].Owner < a[j].Owner
}

func (s *IntegrationCLIPageTestSuite) TestScopesPagination() {
	s.T().Run("GetMetadataGetAllCmd scopes", func(t *testing.T) {
		// Choosing page size = 43 because it a) isn't the default, b) doesn't evenly divide 100.
		pageSize := 43
		expectedCount := s.scopeCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.Scope, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"scopes", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetMetadataGetAllCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.ScopesAllResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.Scopes)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			for _, s := range result.Scopes {
				results = append(results, *s.Scope)
			}
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of scopes returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(scopeSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two scopes should be equal here")
		}
	})

	s.T().Run("GetMetadataScopeCmd all", func(t *testing.T) {
		// Choosing page size = 43 because it a) isn't the default, b) doesn't evenly divide 100.
		pageSize := 43
		expectedCount := s.scopeCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.Scope, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"all", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetMetadataScopeCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.ScopesAllResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.Scopes)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			for _, s := range result.Scopes {
				results = append(results, *s.Scope)
			}
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of scopes returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(scopeSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two scopes should be equal here")
		}
	})

	s.T().Run("GetOwnershipCmd account", func(t *testing.T) {
		// Choosing page size = 17 because it a) isn't the default, b) doesn't evenly divide 80.
		pageSize := 17
		expectedCount := s.accountScopesOwned
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]string, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{s.accountAddrStr, pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetOwnershipCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.OwnershipResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.ScopeUuids)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = append(results, result.ScopeUuids...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of scope UUIDs returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Strings(results)
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two scope UUIDs should be equal here")
		}
	})

	s.T().Run("GetOwnershipCmd user1", func(t *testing.T) {
		// Choosing page size = 17 because it a) isn't the default, b) doesn't evenly divide 60.
		pageSize := 17
		expectedCount := s.user1ScopesOwned
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]string, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{s.user1AddrStr, pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetOwnershipCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.OwnershipResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.ScopeUuids)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = append(results, result.ScopeUuids...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of scope UUIDs returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Strings(results)
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two scope UUIDs should be equal here")
		}
	})

	s.T().Run("GetValueOwnershipCmd account", func(t *testing.T) {
		// Choosing page size = 17 because it a) isn't the default, b) doesn't evenly divide 40.
		pageSize := 17
		expectedCount := s.accountScopesValueOwned
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]string, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{s.accountAddrStr, pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetValueOwnershipCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.ValueOwnershipResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.ScopeUuids)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = append(results, result.ScopeUuids...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of scope UUIDs returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Strings(results)
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two scope UUIDs should be equal here")
		}
	})

	s.T().Run("GetValueOwnershipCmd user1", func(t *testing.T) {
		// Choosing page size = 17 because it a) isn't the default, b) doesn't evenly divide 60.
		pageSize := 17
		expectedCount := s.user1ScopesValueOwned
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]string, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{s.user1AddrStr, pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetValueOwnershipCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.ValueOwnershipResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.ScopeUuids)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = append(results, result.ScopeUuids...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of scope UUIDs returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Strings(results)
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two scope UUIDs should be equal here")
		}
	})
}

func (s *IntegrationCLIPageTestSuite) TestSessionsPagination() {
	s.T().Run("GetMetadataGetAllCmd sessions", func(t *testing.T) {
		// Choosing page size = 57 because it a) isn't the default, b) doesn't evenly divide 200.
		pageSize := 57
		expectedCount := s.sessionCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.Session, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"sessions", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetMetadataGetAllCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.SessionsAllResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.Sessions)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			for _, s := range result.Sessions {
				results = append(results, *s.Session)
			}
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of sessions returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(sessionSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two sessions should be equal here")
		}
	})

	s.T().Run("GetMetadataSessionCmd all", func(t *testing.T) {
		// Choosing page size = 57 because it a) isn't the default, b) doesn't evenly divide 200.
		pageSize := 57
		expectedCount := s.sessionCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.Session, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"all", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetMetadataSessionCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.SessionsAllResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.Sessions)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			for _, s := range result.Sessions {
				results = append(results, *s.Session)
			}
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of sessions returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(sessionSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two sessions should be equal here")
		}
	})
}

func (s *IntegrationCLIPageTestSuite) TestRecordsPagination() {
	s.T().Run("GetMetadataGetAllCmd records", func(t *testing.T) {
		// Choosing page size = 73 because it a) isn't the default, b) doesn't evenly divide 400.
		pageSize := 73
		expectedCount := s.recordCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.Record, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"records", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetMetadataGetAllCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.RecordsAllResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.Records)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			for _, s := range result.Records {
				results = append(results, *s.Record)
			}
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of records returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(recordSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two records should be equal here")
		}
	})

	s.T().Run("GetMetadataRecordCmd all", func(t *testing.T) {
		// Choosing page size = 73 because it a) isn't the default, b) doesn't evenly divide 400.
		pageSize := 73
		expectedCount := s.recordCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.Record, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"all", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetMetadataRecordCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.RecordsAllResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.Records)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			for _, s := range result.Records {
				results = append(results, *s.Record)
			}
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of records returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(recordSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two records should be equal here")
		}
	})
}

func (s *IntegrationCLIPageTestSuite) TestScopeSpecsPagination() {
	s.T().Run("GetMetadataGetAllCmd scopespecs", func(t *testing.T) {
		// Choosing page size = 3 because it a) isn't the default, b) doesn't evenly divide 20.
		pageSize := 3
		expectedCount := s.scopeSpecCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.ScopeSpecification, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"scopespecs", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetMetadataGetAllCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.ScopeSpecificationsAllResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.ScopeSpecifications)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			for _, s := range result.ScopeSpecifications {
				results = append(results, *s.Specification)
			}
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of scope specifications returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(scopeSpecSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two scope specifications should be equal here")
		}
	})

	s.T().Run("GetMetadataScopeSpecCmd all", func(t *testing.T) {
		// Choosing page size = 3 because it a) isn't the default, b) doesn't evenly divide 20.
		pageSize := 3
		expectedCount := s.scopeSpecCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.ScopeSpecification, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"all", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetMetadataScopeSpecCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.ScopeSpecificationsAllResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.ScopeSpecifications)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			for _, s := range result.ScopeSpecifications {
				results = append(results, *s.Specification)
			}
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of scope specifications returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(scopeSpecSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two scope specifications should be equal here")
		}
	})
}

func (s *IntegrationCLIPageTestSuite) TestContractSpecsPagination() {
	s.T().Run("GetMetadataGetAllCmd contractspecs", func(t *testing.T) {
		// Choosing page size = 3 because it a) isn't the default, b) doesn't evenly divide 20.
		pageSize := 3
		expectedCount := s.contractSpecCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.ContractSpecification, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"contractspecs", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetMetadataGetAllCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.ContractSpecificationsAllResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.ContractSpecifications)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			for _, s := range result.ContractSpecifications {
				results = append(results, *s.Specification)
			}
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of contract specifications returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(contractSpecSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two contract specifications should be equal here")
		}
	})

	s.T().Run("GetMetadataContractSpecCmd all", func(t *testing.T) {
		// Choosing page size = 3 because it a) isn't the default, b) doesn't evenly divide 20.
		pageSize := 3
		expectedCount := s.contractSpecCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.ContractSpecification, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"all", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetMetadataContractSpecCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.ContractSpecificationsAllResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.ContractSpecifications)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			for _, s := range result.ContractSpecifications {
				results = append(results, *s.Specification)
			}
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of contract specifications returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(contractSpecSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two contract specifications should be equal here")
		}
	})
}

func (s *IntegrationCLIPageTestSuite) TestRecordSpecsPagination() {
	s.T().Run("GetMetadataGetAllCmd recordspecs", func(t *testing.T) {
		// Choosing page size = 7 because it a) isn't the default, b) doesn't evenly divide 40.
		pageSize := 7
		expectedCount := s.recordSpecCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.RecordSpecification, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"recordspecs", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetMetadataGetAllCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.RecordSpecificationsAllResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.RecordSpecifications)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			for _, s := range result.RecordSpecifications {
				results = append(results, *s.Specification)
			}
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of record specifications returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(recordSpecSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two record specifications should be equal here")
		}
	})

	s.T().Run("GetMetadataRecordSpecCmd all", func(t *testing.T) {
		// Choosing page size = 7 because it a) isn't the default, b) doesn't evenly divide 40.
		pageSize := 7
		expectedCount := s.recordSpecCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.RecordSpecification, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"all", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetMetadataRecordSpecCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.RecordSpecificationsAllResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.RecordSpecifications)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			for _, s := range result.RecordSpecifications {
				results = append(results, *s.Specification)
			}
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of record specifications returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(recordSpecSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two record specifications should be equal here")
		}
	})
}

func (s *IntegrationCLIPageTestSuite) TestOSLocatorPagination() {
	s.T().Run("GetMetadataGetAllCmd locators", func(t *testing.T) {
		// Choosing page size = 7 because it a) isn't the default, b) doesn't evenly divide 60.
		pageSize := 7
		expectedCount := s.osLocatorCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.ObjectStoreLocator, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"locators", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetMetadataGetAllCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.OSAllLocatorsResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.Locators)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = append(results, result.Locators...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of object store locators returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(osLocatorSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two object store locators should be equal here")
		}
	})

	s.T().Run("GetOSLocatorCmd all", func(t *testing.T) {
		// Choosing page size = 7 because it a) isn't the default, b) doesn't evenly divide 60.
		pageSize := 7
		expectedCount := s.osLocatorCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.ObjectStoreLocator, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{"all", pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetOSLocatorCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.OSAllLocatorsResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.Locators)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = append(results, result.Locators...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of object store locators returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(osLocatorSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two object store locators should be equal here")
		}
	})

	s.T().Run("GetOSLocatorCmd uri", func(t *testing.T) {
		// Choosing page size = 8 because it a) isn't the default, b) doesn't evenly divide 20.
		pageSize := 8
		expectedCount := s.osLocatorURICount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]metadatatypes.ObjectStoreLocator, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{s.osLocatorURI, pageSizeArg, s.asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := mdcli.GetOSLocatorCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result metadatatypes.OSAllLocatorsResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.Locators)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = append(results, result.Locators...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of object store locators returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(osLocatorSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two object store locators should be equal here")
		}
	})
}
