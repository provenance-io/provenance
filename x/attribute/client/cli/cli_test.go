package cli_test

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/attribute/client/cli"
	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	namecli "github.com/provenance-io/provenance/x/name/client/cli"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg        testnet.Config
	testnet    *testnet.Network
	keyring    keyring.Keyring
	keyringDir string

	accountAddresses []sdk.AccAddress

	account1Addr sdk.AccAddress
	account1Str  string

	account2Addr sdk.AccAddress
	account2Str  string

	account3Addr sdk.AccAddress
	account3Str  string

	account4Addr sdk.AccAddress
	account4Str  string

	account5Addr sdk.AccAddress
	account5Str  string

	account6Addr sdk.AccAddress
	account6Str  string

	account7Addr sdk.AccAddress
	account7Str  string

	accAttrCount int
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// Generate a number of accounts with keyring entries.
// To use these, do .WithKeyring(s.keyring) when getting the client context.
func (s *IntegrationTestSuite) generateKeyringAndAccounts(number int) {
	path := hd.CreateHDPath(118, 0, 0).String()
	s.keyringDir = s.T().TempDir()
	var err error
	s.keyring, err = keyring.New(s.T().Name(), "test", s.keyringDir, nil, s.cfg.Codec)
	s.Require().NoError(err, "creating keyring")

	s.accountAddresses = make([]sdk.AccAddress, number)
	for i := range s.accountAddresses {
		keyID := fmt.Sprintf("test_key%v", i+1)
		info, _, err := s.keyring.NewMnemonic(keyID, keyring.English, path, "", hd.Secp256k1)
		s.Require().NoError(err, "NewMnemonic(%q)", keyID)
		addr, err := info.GetAddress()
		s.Require().NoError(err, "getting keyring address")
		s.accountAddresses[i] = addr
	}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	pioconfig.SetProvenanceConfig("", 0)
	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.NumValidators = 1
	s.generateKeyringAndAccounts(7)

	s.account1Addr = s.accountAddresses[0]
	s.account1Str = s.account1Addr.String()
	s.account2Addr = s.accountAddresses[1]
	s.account2Str = s.account2Addr.String()
	s.account3Addr = s.accountAddresses[2]
	s.account3Str = s.account3Addr.String()
	s.account4Addr = s.accountAddresses[3]
	s.account4Str = s.account4Addr.String()
	s.account5Addr = s.accountAddresses[4]
	s.account5Str = s.account5Addr.String()
	s.account6Addr = s.accountAddresses[5]
	s.account6Str = s.account6Addr.String()
	s.account7Addr = s.accountAddresses[6]
	s.account7Str = s.account7Addr.String()

	genesisState := s.cfg.GenesisState

	// Configure Genesis data for name module
	attrModAddr := authtypes.NewModuleAddress(attributetypes.ModuleName)
	var nameData nametypes.GenesisState
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("attribute", s.account1Addr, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("example.attribute", s.account1Addr, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord(attributetypes.AccountDataName, attrModAddr, true))
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 3
	nameData.Params.MinSegmentLength = 3
	nameData.Params.MaxSegmentLength = 12
	nameDataBz, err := s.cfg.Codec.MarshalJSON(&nameData)
	s.Require().NoError(err)
	genesisState[nametypes.ModuleName] = nameDataBz

	var authData authtypes.GenesisState
	s.Require().NoError(s.cfg.Codec.UnmarshalJSON(genesisState[authtypes.ModuleName], &authData))
	genAccount1, err := codectypes.NewAnyWithValue(&authtypes.BaseAccount{
		Address:       s.account1Str,
		AccountNumber: 1,
		Sequence:      0,
	})
	s.Require().NoError(err, "NewAnyWithValue genAccount1")
	genAccount5, err := codectypes.NewAnyWithValue(&authtypes.BaseAccount{
		Address:       s.account5Str,
		AccountNumber: 2,
		Sequence:      0,
	})
	s.Require().NoError(err, "NewAnyWithValue genAccount5")
	genAccount6, err := codectypes.NewAnyWithValue(&authtypes.BaseAccount{
		Address:       s.account6Str,
		AccountNumber: 3,
		Sequence:      0,
	})
	s.Require().NoError(err, "NewAnyWithValue genAccount6")
	genAccount7, err := codectypes.NewAnyWithValue(&authtypes.BaseAccount{
		Address:       s.account7Str,
		AccountNumber: 4,
		Sequence:      0,
	})
	s.Require().NoError(err, "NewAnyWithValue genAccount7")
	authData.Accounts = append(authData.Accounts, genAccount1, genAccount5, genAccount6, genAccount7)
	authDataBz, err := s.cfg.Codec.MarshalJSON(&authData)
	s.Require().NoError(err)
	genesisState[authtypes.ModuleName] = authDataBz

	balances := sdk.NewCoins(
		sdk.NewCoin(s.cfg.BondDenom, s.cfg.AccountTokens),
	).Sort()
	var bankData banktypes.GenesisState
	s.Require().NoError(s.cfg.Codec.UnmarshalJSON(genesisState[banktypes.ModuleName], &bankData))
	bankData.Balances = append(bankData.Balances,
		banktypes.Balance{Address: s.account1Str, Coins: balances},
		banktypes.Balance{Address: s.account5Str, Coins: balances},
		banktypes.Balance{Address: s.account6Str, Coins: balances},
		banktypes.Balance{Address: s.account7Str, Coins: balances},
	)
	bankDataBz, err := s.cfg.Codec.MarshalJSON(&bankData)
	s.Require().NoError(err)
	genesisState[banktypes.ModuleName] = bankDataBz

	// Configure Genesis data for attribute module
	var attributeData attributetypes.GenesisState
	attributeData.Attributes = append(attributeData.Attributes,
		attributetypes.NewAttribute(
			"example.attribute",
			s.account1Str,
			attributetypes.AttributeType_String,
			[]byte("example attribute value string"),
			nil),
		attributetypes.NewAttribute(
			"example.attribute.count",
			s.account1Str,
			attributetypes.AttributeType_Int,
			[]byte("2"),
			nil),
		attributetypes.NewAttribute(
			attributetypes.AccountDataName,
			s.account1Str,
			attributetypes.AttributeType_String,
			[]byte("accountdata set at genesis"),
			nil),
		attributetypes.NewAttribute(
			attributetypes.AccountDataName,
			s.account7Str,
			attributetypes.AttributeType_String,
			[]byte("more accountdata set at genesis"),
			nil),
	)
	s.accAttrCount = 500
	for i := 0; i < s.accAttrCount; i++ {
		attributeData.Attributes = append(attributeData.Attributes,
			attributetypes.NewAttribute(
				fmt.Sprintf("example.attribute.%s", toWritten(i)),
				s.account3Str,
				attributetypes.AttributeType_Int,
				[]byte(fmt.Sprintf("%d", i)),
				nil),
			attributetypes.NewAttribute(
				"example.attribute.overload",
				s.account4Str,
				attributetypes.AttributeType_String,
				[]byte(toWritten(i)),
				nil),
		)
	}
	attributeData.Params.MaxValueLength = 128
	attributeDataBz, err := s.cfg.Codec.MarshalJSON(&attributeData)
	s.Require().NoError(err)
	genesisState[attributetypes.ModuleName] = attributeDataBz

	s.cfg.GenesisState = genesisState

	s.cfg.ChainID = antewrapper.SimAppChainID
	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "creating testnet")

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err, "waiting for height 1")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	testutil.CleanUp(s.testnet, s.T())
}

// toWritten converts an integer to a written string version.
// Originally, this was the full written string, e.g. 38 => "thirtyEight" but that ended up being too long for
// an attribute name segment, so it got trimmed down, e.g. 115 => "onehun15".
func toWritten(i int) string {
	if i < 0 || i > 999 {
		panic("cannot convert negative numbers or numbers larger than 999 to written string")
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
		default:
			r = i % 100
			l = toWritten(i/100) + "hun"
		}
		if r == 0 {
			return l
		}
		return l + fmt.Sprintf("%d", r)
	}
}

func limitArg(pageSize int) string {
	return fmt.Sprintf("--limit=%d", pageSize)
}

func pageKeyArg(nextKey string) string {
	return fmt.Sprintf("--page-key=%s", nextKey)
}

// attrSorter implements sort.Interface for []Attribute
// Sorts by .Name then .AttributeType then .Value, then .Address.
type attrSorter []attributetypes.Attribute

func (a attrSorter) Len() int {
	return len(a)
}
func (a attrSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a attrSorter) Less(i, j int) bool {
	// Sort by Name first
	if a[i].Name != a[j].Name {
		return a[i].Name < a[j].Name
	}
	// Then by AttributeType
	if a[i].AttributeType != a[j].AttributeType {
		return a[i].AttributeType < a[j].AttributeType
	}
	// Then by Value.
	// Since this is unit tests, just use the raw byte values rather than going through the trouble of using the AttributeType and converting them.
	for _, vbi := range a[i].Value {
		for _, vbj := range a[j].Value {
			if vbi != vbj {
				return vbi < vbj
			}
		}
	}
	// Then by Address.
	return a[i].Address < a[j].Address
}

func (s *IntegrationTestSuite) TestGetAccountAttributeCmd() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			name:           "should get attribute by name with json output",
			args:           []string{s.account1Addr.String(), "example.attribute", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			expectedOutput: fmt.Sprintf(`{"account":"%s","attributes":[{"name":"example.attribute","value":"ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n","attribute_type":"ATTRIBUTE_TYPE_STRING","address":"%s","expiration_date":null}],"pagination":{"next_key":null,"total":"0"}}`, s.account1Addr.String(), s.account1Addr.String()),
		},
		{
			name: "should get attribute by name with text output",
			args: []string{s.account1Addr.String(), "example.attribute", fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			expectedOutput: fmt.Sprintf(`account: %s
attributes:
- address: %s
  attribute_type: ATTRIBUTE_TYPE_STRING
  expiration_date: null
  name: example.attribute
  value: ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n
pagination:
  next_key: null
  total: "0"`, s.account1Addr.String(), s.account1Addr.String()),
		},
		{
			name: "should fail to find unknown attribute output",
			args: []string{s.account1Addr.String(), "example.none", fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			expectedOutput: fmt.Sprintf(`account: %s
attributes: []
pagination:
  next_key: null
  total: "0"`, s.account1Addr.String()),
		},
		{
			name:           "should fail to find unknown attribute by name with json output",
			args:           []string{s.account1Addr.String(), "example.none", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			expectedOutput: fmt.Sprintf(`{"account":"%s","attributes":[],"pagination":{"next_key":null,"total":"0"}}`, s.account1Addr.String()),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetAccountAttributeCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestScanAccountAttributesCmd() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			name:           "should get attribute by suffix with json output",
			args:           []string{s.account1Addr.String(), "attribute", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			expectedOutput: fmt.Sprintf(`{"account":"%s","attributes":[{"name":"example.attribute","value":"ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n","attribute_type":"ATTRIBUTE_TYPE_STRING","address":"%s","expiration_date":null}],"pagination":{"next_key":null,"total":"0"}}`, s.account1Addr.String(), s.account1Addr.String()),
		},
		{
			name: "should get attribute by suffix with text output",
			args: []string{s.account1Addr.String(), "attribute", fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			expectedOutput: fmt.Sprintf(`account: %s
attributes:
- address: %s
  attribute_type: ATTRIBUTE_TYPE_STRING
  expiration_date: null
  name: example.attribute
  value: ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n
pagination:
  next_key: null
  total: "0"`, s.account1Addr.String(), s.account1Addr.String()),
		},
		{
			name: "should fail to find unknown attribute suffix text output",
			args: []string{s.account1Addr.String(), "none", fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			expectedOutput: fmt.Sprintf(`account: %s
attributes: []
pagination:
  next_key: null
  total: "0"`, s.account1Addr.String()),
		},
		{
			name:           "should get attribute by suffix with json output",
			args:           []string{s.account1Addr.String(), "none", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			expectedOutput: fmt.Sprintf(`{"account":"%s","attributes":[],"pagination":{"next_key":null,"total":"0"}}`, s.account1Addr.String()),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.ScanAccountAttributesCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestListAccountAttributesCmd() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			name:           "should list all attributes for account with json output",
			args:           []string{s.account1Addr.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			expectedOutput: fmt.Sprintf(`{"account":"%[1]s","attributes":[{"name":"example.attribute.count","value":"Mg==","attribute_type":"ATTRIBUTE_TYPE_INT","address":"%[1]s","expiration_date":null},{"name":"example.attribute","value":"ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n","attribute_type":"ATTRIBUTE_TYPE_STRING","address":"%[1]s","expiration_date":null},{"name":"accountdata","value":"YWNjb3VudGRhdGEgc2V0IGF0IGdlbmVzaXM=","attribute_type":"ATTRIBUTE_TYPE_STRING","address":"%[1]s","expiration_date":null}],"pagination":{"next_key":null,"total":"0"}}`, s.account1Addr.String()),
		},
		{
			name: "should list all attributes for account text output",
			args: []string{s.account1Addr.String(), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			expectedOutput: fmt.Sprintf(`account: %[1]s
attributes:
- address: %[1]s
  attribute_type: ATTRIBUTE_TYPE_INT
  expiration_date: null
  name: example.attribute.count
  value: Mg==
- address: %[1]s
  attribute_type: ATTRIBUTE_TYPE_STRING
  expiration_date: null
  name: example.attribute
  value: ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n
- address: %[1]s
  attribute_type: ATTRIBUTE_TYPE_STRING
  expiration_date: null
  name: accountdata
  value: YWNjb3VudGRhdGEgc2V0IGF0IGdlbmVzaXM=
pagination:
  next_key: null
  total: "0"`,
				s.account1Addr.String()),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.ListAccountAttributesCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestGetAttributeParamsCmd() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			name:           "json output",
			args:           []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			expectedOutput: "{\"max_value_length\":128}",
		},
		{
			name:           "text output",
			args:           []string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			expectedOutput: "max_value_length: 128",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetAttributeParamsCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestAttributeAccountsCmd() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			name:           "successfully output result for attribute",
			args:           []string{"example.attribute"},
			expectedOutput: fmt.Sprintf("{\"accounts\":[\"%s\"],\"pagination\":{\"next_key\":null,\"total\":\"0\"}}", s.account1Addr),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetAttributeAccountsCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			tc.args = append(tc.args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestAttributeTxCommands() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			name: "bind a new attribute name for testing",
			cmd:  namecli.GetBindNameCmd(),
			args: []string{
				"txtest",
				s.testnet.Validators[0].Address.String(),
				"attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "set attribute, valid string",
			cmd:  cli.NewAddAccountAttributeCmd(),
			args: []string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "set attribute, invalid expiration",
			cmd:  cli.NewAddAccountAttributeCmd(),
			args: []string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"string",
				"test value with expiration",
				"foo",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "set attribute, valid expiration",
			cmd:  cli.NewAddAccountAttributeCmd(),
			args: []string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"string",
				"test value with expiration",
				"2050-01-15T00:00:00Z",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "set attribute, invalid bech32 address",
			cmd:  cli.NewAddAccountAttributeCmd(),
			args: []string{
				"txtest.attribute",
				"invalidbech32",
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "set attribute, invalid type",
			cmd:  cli.NewAddAccountAttributeCmd(),
			args: []string{
				"txtest.attribute",
				s.account2Addr.String(),
				"blah",
				"3.14159",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 1,
		},
		{
			name: "set attribute, cannot encode",
			cmd:  cli.NewAddAccountAttributeCmd(),
			args: []string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"bytes",
				"3.14159",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 1,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestAccountDataCmds() {
	jsonOut := "--" + tmcli.OutputFlag + "=json"
	txFlags := func(from string, firstArgs ...string) []string {
		return append(firstArgs,
			"--"+flags.FlagFrom, from,
			"--"+flags.FlagSkipConfirmation,
			"--"+flags.FlagBroadcastMode, flags.BroadcastBlock,
			"--"+flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
			jsonOut,
		)
	}
	queryFlags := func(args ...string) []string {
		return append(args, jsonOut)
	}

	tests := []struct {
		name      string
		cmd       *cobra.Command
		args      []string
		expErr    string
		respType  proto.Message // Don't set if the command should error before sending anything.
		expResp   proto.Message // Only applicable if respType is neither nil nor a *sdk.TxResponse.
		expCode   int           // Only applicable if respType is a *sdk.TxResponse.
		expRawLog string        // Only applicable if respType is a *sdk.TxResponse. Ignored if "".
	}{
		{
			name:     "query account-data account1",
			cmd:      cli.GetQueryCmd(),
			args:     queryFlags("account-data", s.account1Str),
			respType: &attributetypes.QueryAccountDataResponse{},
			expResp:  &attributetypes.QueryAccountDataResponse{Value: "accountdata set at genesis"},
		},
		{
			name:     "query accountdata account7",
			cmd:      cli.GetQueryCmd(),
			args:     queryFlags("accountdata", s.account7Str),
			respType: &attributetypes.QueryAccountDataResponse{},
			expResp:  &attributetypes.QueryAccountDataResponse{Value: "more accountdata set at genesis"},
		},
		{
			name:     "query ad account7",
			cmd:      cli.GetQueryCmd(),
			args:     queryFlags("ad", s.account7Str),
			respType: &attributetypes.QueryAccountDataResponse{},
			expResp:  &attributetypes.QueryAccountDataResponse{Value: "more accountdata set at genesis"},
		},
		{
			name:     "query account-data account5 (not yet set)",
			cmd:      cli.GetQueryCmd(),
			args:     queryFlags("account-data", s.account5Str),
			respType: &attributetypes.QueryAccountDataResponse{},
			expResp:  &attributetypes.QueryAccountDataResponse{Value: ""},
		},
		{
			name:   "tx account-data account5 but no value flags provided",
			cmd:    cli.NewTxCmd(),
			args:   txFlags(s.account5Str, "account-data"),
			expErr: "exactly one of these must be provided: {--value <value>|--file <file>|--delete}",
		},
		{
			name:     "tx account-data account5",
			cmd:      cli.NewTxCmd(),
			args:     txFlags(s.account5Str, "account-data", "--value", "This is account2's account data."),
			respType: &sdk.TxResponse{},
			expCode:  0,
		},
		{
			name:     "query account-data account5 after set",
			cmd:      cli.GetQueryCmd(),
			args:     queryFlags("account-data", s.account5Str),
			respType: &attributetypes.QueryAccountDataResponse{},
			expResp:  &attributetypes.QueryAccountDataResponse{Value: "This is account2's account data."},
		},
		{
			name:     "tx accountdata account7 overwrite",
			cmd:      cli.NewTxCmd(),
			args:     txFlags(s.account7Str, "accountdata", "--value", "This is account7's new account data."),
			respType: &sdk.TxResponse{},
			expCode:  0,
		},
		{
			name:     "query account-data account7 after overwrite",
			cmd:      cli.GetQueryCmd(),
			args:     queryFlags("account-data", s.account7Str),
			respType: &attributetypes.QueryAccountDataResponse{},
			expResp:  &attributetypes.QueryAccountDataResponse{Value: "This is account7's new account data."},
		},
		{
			name:     "tx ad account5 delete",
			cmd:      cli.NewTxCmd(),
			args:     txFlags(s.account5Str, "ad", "--delete"),
			respType: &sdk.TxResponse{},
			expCode:  0,
		},
		{
			name:     "query account-data account5 after delete",
			cmd:      cli.GetQueryCmd(),
			args:     queryFlags("account-data", s.account5Str),
			respType: &attributetypes.QueryAccountDataResponse{},
			expResp:  &attributetypes.QueryAccountDataResponse{Value: ""},
		},
		{
			name: "tx account-data account6 value too long",
			cmd:  cli.NewTxCmd(),
			// From SetupSuite: attributeData.Params.MaxValueLength = 128.
			args: txFlags(s.account6Str, "account-data", "--value",
				strings.Join([]string{
					"This value is going to be way too long.",                                  // 39 chars
					"It just has too many characters in it.",                                   // 38 chars
					"And I'm not talking about characters like Mickey Mouse or Darkwing Duck.", // 72 chars
					"I'm not even talking about Howard the Duck.",                              // 43 chars
					"Look. this thing is just way too long.",                                   // 38 chars
				}, "\n"), // 39 + 38 + 72 + 43 + 38 + 4 newlines = 234 characters total.
			),
			respType: &sdk.TxResponse{},
			expCode:  1,
			expRawLog: "failed to execute message; message index: 0: " +
				`could not set accountdata for "` + s.account6Str + `": attribute value length of 234 exceeds max length 128`,
		},
		{
			name:     "query account-data account6 after failed set",
			cmd:      cli.GetQueryCmd(),
			args:     queryFlags("account-data", s.account6Str),
			respType: &attributetypes.QueryAccountDataResponse{},
			expResp:  &attributetypes.QueryAccountDataResponse{Value: ""},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx.WithKeyring(s.keyring)
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)
			outBz := out.Bytes()
			outStr := string(outBz)
			s.T().Logf("Args: %q\nOutput:\n%s", tc.args, outStr)
			if len(tc.expErr) > 0 {
				s.Require().EqualError(err, tc.expErr, "cmd execution error")
				s.Require().Contains(outStr, tc.expErr, "cmd output")
			} else {
				s.Require().NoError(err, "cmd execution error")
			}

			if tc.respType != nil {
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(outBz, tc.respType), "marshalling output to %T", tc.respType)
				txResp, isTxResp := tc.respType.(*sdk.TxResponse)
				if isTxResp {
					s.Assert().Equal(tc.expCode, int(txResp.Code), "TxResponse code")
					if len(tc.expRawLog) > 0 {
						s.Assert().Equal(tc.expRawLog, txResp.RawLog, "txResp.RawLog")
					}
				} else {
					s.Assert().Equal(tc.expResp, tc.respType, "command response")
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateAccountAttributeTxCommands() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			name: "bind a new attribute name for delete testing",
			cmd:  namecli.GetBindNameCmd(),
			args: []string{
				"updatetest",
				s.testnet.Validators[0].Address.String(),
				"attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "add new attribute for updating",
			cmd:  cli.NewAddAccountAttributeCmd(),
			args: []string{
				"updatetest.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "fail to update attribute, account address failure",
			cmd:  cli.NewUpdateAccountAttributeCmd(),
			args: []string{
				"updatetest.attribute",
				"not-an-address",
				"string",
				"test value",
				"int",
				"10",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "fail to update attribute, incorrect original type",
			cmd:  cli.NewUpdateAccountAttributeCmd(),
			args: []string{
				"updatetest.attribute",
				s.account2Addr.String(),
				"invalid",
				"test value",
				"int",
				"10",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "fail to update attribute, incorrect update type",
			cmd:  cli.NewUpdateAccountAttributeCmd(),
			args: []string{
				"updatetest.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				"invalid",
				"10",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "fail to update attribute, validate basic fail",
			cmd:  cli.NewUpdateAccountAttributeCmd(),
			args: []string{
				"updatetest.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				"init",
				"nan",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "successful update of attribute",
			cmd:  cli.NewUpdateAccountAttributeCmd(),
			args: []string{
				"updatetest.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				"int",
				"10",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestDeleteDistinctAccountAttributeTxCommands() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			name: "bind a new attribute name for delete testing",
			cmd:  namecli.GetBindNameCmd(),
			args: []string{
				"distinct",
				s.testnet.Validators[0].Address.String(),
				"attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "add new attribute for delete testing",
			cmd:  cli.NewAddAccountAttributeCmd(),
			args: []string{
				"distinct.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "delete distinct attribute, should fail incorrect address",
			cmd:  cli.NewDeleteDistinctAccountAttributeCmd(),
			args: []string{
				"distinct.attribute",
				"not-a-address",
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "delete distinct attribute, should fail incorrect type",
			cmd:  cli.NewDeleteDistinctAccountAttributeCmd(),
			args: []string{
				"distinct.attribute",
				s.account2Addr.String(),
				"invalid",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "delete distinct attribute, should successfully delete",
			cmd:  cli.NewDeleteDistinctAccountAttributeCmd(),
			args: []string{
				"distinct.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestDeleteAccountAttributeTxCommands() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			name: "bind a new attribute name for delete testing",
			cmd:  namecli.GetBindNameCmd(),
			args: []string{
				"deletetest",
				s.testnet.Validators[0].Address.String(),
				"attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "add new attribute for delete testing",
			cmd:  cli.NewAddAccountAttributeCmd(),
			args: []string{
				"deletetest.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "delete attribute, should delete txtest.attribute",
			cmd:  cli.NewDeleteAccountAttributeCmd(),
			args: []string{
				"deletetest.attribute",
				s.account2Addr.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "delete attribute, should fail to find txtest.attribute",
			cmd:  cli.NewDeleteAccountAttributeCmd(),
			args: []string{
				"deletetest.attribute",
				s.account2Addr.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 1,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestPaginationWithPageKey() {
	asJson := fmt.Sprintf("--%s=json", tmcli.OutputFlag)

	s.T().Run("GetAccountAttribute", func(t *testing.T) {
		// Choosing page size = 35 because it a) isn't the default, b) doesn't evenly divide 500.
		pageSize := 35
		expectedCount := s.accAttrCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]attributetypes.Attribute, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			// account 4 = lots of attributes with the same name but different values.
			args := []string{s.account4Str, "example.attribute.overload", pageSizeArg, asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := cli.GetAccountAttributeCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result attributetypes.QueryAttributeResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.Attributes)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = append(results, result.Attributes...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of attributes returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(attrSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two attributes should be equal here")
		}
	})

	s.T().Run("ListAccountAttribute", func(t *testing.T) {
		// Choosing page size = 35 because it a) isn't the default, b) doesn't evenly divide 500.
		pageSize := 35
		expectedCount := s.accAttrCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]attributetypes.Attribute, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			// account 3 = lots of attributes different names
			args := []string{s.account3Str, pageSizeArg, asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := cli.ListAccountAttributesCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result attributetypes.QueryAttributesResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.Attributes)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = append(results, result.Attributes...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of attributes returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(attrSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two attributes should be equal here")
		}
	})

	s.T().Run("ScanAccountAttributesCmd different names", func(t *testing.T) {
		// Choosing page size = 35 because it a) isn't the default, b) doesn't evenly divide 48.
		// 48 comes from the number of attributes on account 3 that end with the character '7' (500/10 - "seven" - "seventeen").
		pageSize := 35
		expectedCount := 48
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]attributetypes.Attribute, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			// account 3 = lots of attributes different names
			args := []string{s.account3Str, "7", pageSizeArg, asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := cli.ScanAccountAttributesCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result attributetypes.QueryScanResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.Attributes)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = append(results, result.Attributes...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of attributes returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(attrSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two attributes should be equal here")
		}
	})

	s.T().Run("ScanAccountAttributesCmd different values", func(t *testing.T) {
		// Choosing page size = 35 because it a) isn't the default, b) doesn't evenly divide 500.
		pageSize := 35
		expectedCount := s.accAttrCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]attributetypes.Attribute, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			// account 4 = lots of attributes with the same name but different values.
			args := []string{s.account4Str, "load", pageSizeArg, asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := cli.ScanAccountAttributesCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result attributetypes.QueryScanResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultAttrCount := len(result.Attributes)
			if page != pageCount {
				require.Equalf(t, pageSize, resultAttrCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultAttrCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = append(results, result.Attributes...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of attributes returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(attrSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two attributes should be equal here")
		}
	})
}

func (s *IntegrationTestSuite) TestUpdateAccountAttributeExpirationCmd() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    string
		expectedCode int32
	}{
		{
			name: "bind a new attribute name for delete testing",
			cmd:  namecli.GetBindNameCmd(),
			args: []string{
				"expiration",
				s.testnet.Validators[0].Address.String(),
				"attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "add new attribute for updating expire date testing",
			cmd:  cli.NewAddAccountAttributeCmd(),
			args: []string{
				"expiration.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "update expire date, should fail incorrect address",
			cmd:  cli.NewUpdateAccountAttributeExpirationCmd(),
			args: []string{
				"expiration.attribute",
				"not-a-address",
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: `invalid address: must be either an account address or scope metadata address: "not-a-address"`,
		},
		{
			name: "update expire date, should fail incorrect date",
			cmd:  cli.NewUpdateAccountAttributeExpirationCmd(),
			args: []string{
				"expiration.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: `unable to parse time "test value" required format is RFC3339 (2006-01-02T15:04:05Z07:00): parsing time "test value" as "2006-01-02T15:04:05Z07:00": cannot parse "test value" as "2006"`,
		},
		{
			name: "update expire date, should succeed",
			cmd:  cli.NewUpdateAccountAttributeExpirationCmd(),
			args: []string{
				"expiration.attribute",
				s.account2Addr.String(),
				"test value",
				"2050-01-15T00:00:00Z",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "update expire date, should succeed removes expiration",
			cmd:  cli.NewUpdateAccountAttributeExpirationCmd(),
			args: []string{
				"expiration.attribute",
				s.account2Addr.String(),
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)

			if len(tc.expectErr) > 0 {
				s.Require().EqualError(err, tc.expectErr)
			} else {
				var response sdk.TxResponse
				s.Assert().NoError(err)
				s.Assert().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				s.Assert().Equal(tc.expectedCode, int32(response.Code), "")
			}
		})
	}
}
