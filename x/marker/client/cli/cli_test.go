package cli_test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdktypes "github.com/cosmos/cosmos-sdk/codec/types"
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
	markercli "github.com/provenance-io/provenance/x/marker/client/cli"
	"github.com/provenance-io/provenance/x/marker/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

const (
	oneYear = 365 * 24 * 60 * 60
	oneHour = 60 * 60
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg              testnet.Config
	testnet          *testnet.Network
	keyring          keyring.Keyring
	keyringDir       string
	accountAddresses []sdk.AccAddress

	holderDenom string
	holderCount int
	markerCount int
}

func (s *IntegrationTestSuite) GenerateAccountsWithKeyrings(number int) {
	path := hd.CreateHDPath(118, 0, 0).String()
	s.keyringDir = s.T().TempDir()
	kr, err := keyring.New(s.T().Name(), "test", s.keyringDir, nil, s.cfg.Codec)
	s.Require().NoError(err)
	s.keyring = kr
	for i := 0; i < number; i++ {
		keyId := fmt.Sprintf("test_key%v", i)
		info, _, err := kr.NewMnemonic(keyId, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err)
		addr, err := info.GetAddress()
		s.Require().NoError(err, "getting keyring address")
		s.accountAddresses = append(s.accountAddresses, addr)
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	pioconfig.SetProvenanceConfig("", 0)
	cfg := testutil.DefaultTestNetworkConfig()

	genesisState := cfg.GenesisState
	cfg.NumValidators = 1
	s.cfg = cfg
	s.GenerateAccountsWithKeyrings(4)

	// Configure Genesis auth data for adding test accounts
	var genAccounts []authtypes.GenesisAccount
	var authData authtypes.GenesisState
	authData.Params = authtypes.DefaultParams()
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[0], nil, 3, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[1], nil, 4, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[2], nil, 5, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[3], nil, 6, 0))
	accounts, err := authtypes.PackAccounts(genAccounts)
	s.Require().NoError(err)
	authData.Accounts = accounts
	authDataBz, err := cfg.Codec.MarshalJSON(&authData)
	s.Require().NoError(err)
	genesisState[authtypes.ModuleName] = authDataBz

	s.holderDenom = "hodlercoin"
	s.holderCount = 4

	// Configure Genesis bank data for test accounts
	var genBalances []banktypes.Balance
	genBalances = append(genBalances, banktypes.Balance{Address: s.accountAddresses[0].String(), Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		sdk.NewCoin("authzhotdog", sdk.NewInt(100)),
		sdk.NewCoin(s.holderDenom, sdk.NewInt(123)),
	).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: s.accountAddresses[1].String(), Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		sdk.NewCoin("authzhotdog", sdk.NewInt(100)),
		sdk.NewCoin(s.holderDenom, sdk.NewInt(234)),
	).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: s.accountAddresses[2].String(), Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		sdk.NewCoin(s.holderDenom, sdk.NewInt(345)),
	).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: s.accountAddresses[3].String(), Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		sdk.NewCoin(s.holderDenom, sdk.NewInt(456)),
	).Sort()})

	genBalances = append(genBalances, banktypes.Balance{Address: markertypes.MustGetMarkerAddress("testcoin").String(), Coins: sdk.NewCoins(
		sdk.NewCoin("testcoin", sdk.NewInt(1000)),
	).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: markertypes.MustGetMarkerAddress("lockedcoin").String(), Coins: sdk.NewCoins(
		sdk.NewCoin("lockedcoin", sdk.NewInt(1000)),
	).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: markertypes.MustGetMarkerAddress("propcoin").String(), Coins: sdk.NewCoins(
		sdk.NewCoin("propcoin", sdk.NewInt(1000)),
	).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: markertypes.MustGetMarkerAddress("authzhotdog").String(), Coins: sdk.NewCoins(
		sdk.NewCoin("authzhotdog", sdk.NewInt(800)),
	).Sort()})

	var bankGenState banktypes.GenesisState
	bankGenState.Params = banktypes.DefaultParams()
	bankGenState.Balances = genBalances
	bankDataBz, err := cfg.Codec.MarshalJSON(&bankGenState)
	s.Require().NoError(err)
	genesisState[banktypes.ModuleName] = bankDataBz

	s.markerCount = 20

	// Configure Genesis data for marker module
	var markerData markertypes.GenesisState
	markerData.Params.EnableGovernance = true
	markerData.Params.MaxTotalSupply = 1000000
	// Note: These account numbers get ignored.
	markerData.Markers = []markertypes.MarkerAccount{
		{
			BaseAccount: &authtypes.BaseAccount{
				Address:       markertypes.MustGetMarkerAddress("testcoin").String(),
				AccountNumber: 100,
				Sequence:      0,
			},
			Status:                 markertypes.StatusActive,
			SupplyFixed:            true,
			MarkerType:             markertypes.MarkerType_Coin,
			AllowGovernanceControl: false,
			Supply:                 sdk.NewInt(1000),
			Denom:                  "testcoin",
			AllowForcedTransfer:    false,
		},
		{
			BaseAccount: &authtypes.BaseAccount{
				Address:       markertypes.MustGetMarkerAddress("lockedcoin").String(),
				AccountNumber: 110,
				Sequence:      0,
			},
			Status:                 markertypes.StatusActive,
			SupplyFixed:            true,
			MarkerType:             markertypes.MarkerType_RestrictedCoin,
			AllowGovernanceControl: false,
			Supply:                 sdk.NewInt(1000),
			Denom:                  "lockedcoin",
			AllowForcedTransfer:    false,
		},
		{
			BaseAccount: &authtypes.BaseAccount{
				Address:       markertypes.MustGetMarkerAddress("propcoin").String(),
				AccountNumber: 120,
				Sequence:      0,
			},
			Status:                 markertypes.StatusActive,
			SupplyFixed:            true,
			MarkerType:             markertypes.MarkerType_Coin,
			AllowGovernanceControl: true,
			Supply:                 sdk.NewInt(1000),
			Denom:                  "propcoin",
			AllowForcedTransfer:    false,
		},
		{
			BaseAccount: &authtypes.BaseAccount{
				Address:       markertypes.MustGetMarkerAddress(cfg.BondDenom).String(),
				AccountNumber: 130,
				Sequence:      0,
			},
			Status:                 markertypes.StatusActive,
			SupplyFixed:            false,
			MarkerType:             markertypes.MarkerType_Coin,
			AllowGovernanceControl: true,
			Supply:                 cfg.BondedTokens.Mul(sdk.NewInt(int64(cfg.NumValidators))),
			Denom:                  cfg.BondDenom,
			AllowForcedTransfer:    false,
		},
		{
			BaseAccount: &authtypes.BaseAccount{
				Address:       markertypes.MustGetMarkerAddress("authzhotdog").String(),
				AccountNumber: 140,
				Sequence:      0,
			},
			Status:                 markertypes.StatusActive,
			SupplyFixed:            true,
			MarkerType:             markertypes.MarkerType_RestrictedCoin,
			AllowGovernanceControl: false,
			Supply:                 sdk.NewInt(1000),
			Denom:                  "authzhotdog",
			AccessControl: []markertypes.AccessGrant{
				*markertypes.NewAccessGrant(s.accountAddresses[0], []markertypes.Access{markertypes.Access_Transfer, markertypes.Access_Admin}),
				*markertypes.NewAccessGrant(s.accountAddresses[1], []markertypes.Access{markertypes.Access_Transfer, markertypes.Access_Admin}),
				*markertypes.NewAccessGrant(s.accountAddresses[2], []markertypes.Access{markertypes.Access_Transfer, markertypes.Access_Admin}),
			},
			AllowForcedTransfer: false,
		},
		{
			BaseAccount: &authtypes.BaseAccount{
				Address:       markertypes.MustGetMarkerAddress("hodlercoin").String(),
				AccountNumber: 150,
				Sequence:      0,
			},
			Status:                 markertypes.StatusActive,
			SupplyFixed:            false,
			MarkerType:             markertypes.MarkerType_RestrictedCoin,
			AllowGovernanceControl: false,
			Supply:                 sdk.NewInt(3000),
			Denom:                  "hodlercoin",
			AllowForcedTransfer:    true,
		},
	}
	for i := len(markerData.Markers); i < s.markerCount; i++ {
		denom := toWritten(i + 1)
		markerData.Markers = append(markerData.Markers,
			markertypes.MarkerAccount{
				BaseAccount: &authtypes.BaseAccount{
					Address:       markertypes.MustGetMarkerAddress(denom).String(),
					AccountNumber: uint64(i * 10),
					Sequence:      0,
				},
				Status:                 markertypes.StatusActive,
				SupplyFixed:            false,
				MarkerType:             markertypes.MarkerType_Coin,
				AllowGovernanceControl: true,
				Supply:                 sdk.NewInt(int64(i * 100000)),
				Denom:                  denom,
				AllowForcedTransfer:    false,
			},
		)
	}
	markerDataBz, err := cfg.Codec.MarshalJSON(&markerData)
	s.Require().NoError(err)
	genesisState[markertypes.ModuleName] = markerDataBz

	cfg.GenesisState = genesisState
	cfg.ChainID = antewrapper.SimAppChainID

	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), cfg)
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

// markerSorter implements sort.Interface for []MarkerAccount
// Sorts by .Denom only.
type markerSorter []markertypes.MarkerAccount

func (a markerSorter) Len() int {
	return len(a)
}
func (a markerSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a markerSorter) Less(i, j int) bool {
	return a[i].Denom < a[j].Denom
}

// balanceSorter implements sort.Interface for []Balance
// Sorts by .Address, then .Coins length.
type balanceSorter []markertypes.Balance

func (a balanceSorter) Len() int {
	return len(a)
}
func (a balanceSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a balanceSorter) Less(i, j int) bool {
	if a[i].Address != a[j].Address {
		return a[i].Address < a[j].Address
	}
	return len(a[i].Coins) < len(a[j].Coins)
}

func appendMarkers(a []markertypes.MarkerAccount, toAdd ...*sdktypes.Any) []markertypes.MarkerAccount {
	for _, n := range toAdd {
		var ma markertypes.MarkerAccount
		err := ma.Unmarshal(n.Value)
		if err != nil {
			panic(err.Error())
		}
		a = append(a, ma)
	}
	return a
}

func (s *IntegrationTestSuite) TestMarkerQueryCommands() {
	testCases := []struct {
		name           string
		cmd            *cobra.Command
		args           []string
		expectedOutput string
	}{
		{
			"get marker params json",
			markercli.QueryParamsCmd(),
			[]string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			`{"max_total_supply":"1000000","enable_governance":true,"unrestricted_denom_regex":""}`,
		},
		{
			"get testcoin marker json",
			markercli.MarkerCmd(),
			[]string{
				"testcoin",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			`{"marker":{"@type":"/provenance.marker.v1.MarkerAccount","base_account":{"address":"cosmos1p3sl9tll0ygj3flwt5r2w0n6fx9p5ngq2tu6mq","pub_key":null,"account_number":"8","sequence":"0"},"manager":"","access_control":[],"status":"MARKER_STATUS_ACTIVE","denom":"testcoin","supply":"1000","marker_type":"MARKER_TYPE_COIN","supply_fixed":true,"allow_governance_control":false,"allow_forced_transfer":false,"required_attributes":[]}}`,
		},
		{
			"get testcoin marker test",
			markercli.MarkerCmd(),
			[]string{
				"testcoin",
				fmt.Sprintf("--%s=text", tmcli.OutputFlag),
			},
			`marker:
  '@type': /provenance.marker.v1.MarkerAccount
  access_control: []
  allow_forced_transfer: false
  allow_governance_control: false
  base_account:
    account_number: "8"
    address: cosmos1p3sl9tll0ygj3flwt5r2w0n6fx9p5ngq2tu6mq
    pub_key: null
    sequence: "0"
  denom: testcoin
  manager: ""
  marker_type: MARKER_TYPE_COIN
  required_attributes: []
  status: MARKER_STATUS_ACTIVE
  supply: "1000"
  supply_fixed: true`,
		},
		{
			"query non existent marker",
			markercli.MarkerCmd(),
			[]string{
				"doesntexist",
			},
			"",
		},
		{
			"get restricted coin marker",
			markercli.MarkerCmd(),
			[]string{
				"lockedcoin",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			`{"marker":{"@type":"/provenance.marker.v1.MarkerAccount","base_account":{"address":"cosmos16437wt0xtqtuw0pn4vt8rlf8gr2plz2det0mt2","pub_key":null,"account_number":"9","sequence":"0"},"manager":"","access_control":[],"status":"MARKER_STATUS_ACTIVE","denom":"lockedcoin","supply":"1000","marker_type":"MARKER_TYPE_RESTRICTED","supply_fixed":true,"allow_governance_control":false,"allow_forced_transfer":false,"required_attributes":[]}}`,
		},
		{
			"get restricted coin marker with forced transfer",
			markercli.MarkerCmd(),
			[]string{
				s.holderDenom,
			},

			`marker:
  '@type': /provenance.marker.v1.MarkerAccount
  access_control: []
  allow_forced_transfer: true
  allow_governance_control: false
  base_account:
    account_number: "13"
    address: cosmos1ae2206l700zfkxyqvd6cwn3gddas3rjy6z6g4u
    pub_key: null
    sequence: "0"
  denom: hodlercoin
  manager: ""
  marker_type: MARKER_TYPE_RESTRICTED
  required_attributes: []
  status: MARKER_STATUS_ACTIVE
  supply: "3000"
  supply_fixed: false`,
		},
		{
			"query access",
			markercli.MarkerAccessCmd(),
			[]string{
				s.cfg.BondDenom,
			},
			"accounts: []",
		},
		{
			"query escrow",
			markercli.MarkerEscrowCmd(),
			[]string{
				s.cfg.BondDenom,
			},
			"escrow: []",
		},
		{
			"query supply",
			markercli.MarkerSupplyCmd(),
			[]string{
				s.cfg.BondDenom,
			},
			fmt.Sprintf("amount:\n  amount: \"%s\"\n  denom: %s", s.cfg.BondedTokens.Mul(sdk.NewInt(int64(s.cfg.NumValidators))), s.cfg.BondDenom),
		},
	}
	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestMarkerTxCommands() {
	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"create a new marker",
			markercli.GetCmdAddMarker(),
			[]string{
				"1000hotdog",
				fmt.Sprintf("--%s=%s", markercli.FlagType, "RESTRICTED"),
				fmt.Sprintf("--%s=%s", markercli.FlagSupplyFixed, "true"),
				fmt.Sprintf("--%s=%s", markercli.FlagAllowGovernanceControl, "true"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"create a new marker with dashes and periods",
			markercli.GetCmdAddMarker(),
			[]string{
				"1000cat-scratch-fever.bobcat",
				fmt.Sprintf("--%s=%s", markercli.FlagType, "RESTRICTED"),
				fmt.Sprintf("--%s=%s", markercli.FlagSupplyFixed, "true"),
				fmt.Sprintf("--%s=%s", markercli.FlagAllowGovernanceControl, "true"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"fail to create add marker, incorrect allow governance value",
			markercli.GetCmdAddMarker(),
			[]string{
				"1000hotdog",
				fmt.Sprintf("--%s=%s", markercli.FlagType, "RESTRICTED"),
				fmt.Sprintf("--%s=%s", markercli.FlagSupplyFixed, "false"),
				fmt.Sprintf("--%s=%s", markercli.FlagAllowGovernanceControl, "wrong"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"fail to create add marker, incorrect supply fixed value",
			markercli.GetCmdAddMarker(),
			[]string{
				"1000hotdog",
				fmt.Sprintf("--%s=%s", markercli.FlagType, "RESTRICTED"),
				fmt.Sprintf("--%s=%s", markercli.FlagSupplyFixed, "wrong"),
				fmt.Sprintf("--%s=%s", markercli.FlagAllowGovernanceControl, "true"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"fail to create feegrant not admin",
			markercli.GetCmdFeeGrant(),
			[]string{
				"hotdog",
				s.testnet.Validators[0].Address.String(),
				s.accountAddresses[0].String(),
				fmt.Sprintf("--%s=%s", markercli.FlagSpendLimit, sdk.NewCoin("stake", sdk.NewInt(100))),
				fmt.Sprintf("--%s=%s", markercli.FlagExpiration, getFormattedExpiration(oneYear)),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 4,
		},
		{
			"add single access",
			markercli.GetCmdAddAccess(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				"hotdog",
				"admin",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"add multiple access",
			markercli.GetCmdAddAccess(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				"hotdog",
				"mint,burn,transfer,withdraw",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"mint supply",
			markercli.GetCmdMint(),
			[]string{
				"100hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"burn supply",
			markercli.GetCmdBurn(),
			[]string{
				"100hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"finalize",
			markercli.GetCmdFinalize(),
			[]string{
				"hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"activate",
			markercli.GetCmdActivate(),
			[]string{
				"hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"create simple feegrant allowance",
			markercli.GetCmdFeeGrant(),
			[]string{
				"hotdog",
				s.testnet.Validators[0].Address.String(),
				s.accountAddresses[0].String(),
				fmt.Sprintf("--%s=%s", markercli.FlagSpendLimit, sdk.NewCoin("stake", sdk.NewInt(100))),
				fmt.Sprintf("--%s=%s", markercli.FlagExpiration, getFormattedExpiration(oneYear)),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"create periodic feegrant allowance",
			markercli.GetCmdFeeGrant(),
			[]string{
				"hotdog",
				s.testnet.Validators[0].Address.String(),
				s.accountAddresses[0].String(),
				fmt.Sprintf("--%s=%v", markercli.FlagPeriod, oneHour),
				fmt.Sprintf("--%s=%s", markercli.FlagPeriodLimit, sdk.NewCoin("stake", sdk.NewInt(100))),
				fmt.Sprintf("--%s=%s", markercli.FlagExpiration, getFormattedExpiration(oneYear)),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"withdraw, fail to parse coins",
			markercli.GetCmdWithdrawCoins(),
			[]string{
				"hotdog",
				"incorrect-denom-blah",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"withdraw, fail to parse recipient address",
			markercli.GetCmdWithdrawCoins(),
			[]string{
				"hotdog",
				"40hotdog",
				"invalid-recipient",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"withdraw, successful withdraw to a recipient",
			markercli.GetCmdWithdrawCoins(),
			[]string{
				"hotdog",
				"40hotdog",
				s.accountAddresses[0].String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"withdraw, successful withdraw to caller's account",
			markercli.GetCmdWithdrawCoins(),
			[]string{
				"hotdog",
				"200hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"transfer, fail to transfer invalid from address",
			markercli.GetNewTransferCmd(),
			[]string{
				"invalid-from",
				s.testnet.Validators[0].Address.String(),
				"100hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"transfer, fail to transfer invalid to address",
			markercli.GetNewTransferCmd(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				"not-to",
				"100hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"transfer, fail to transfer invalid coin parse",
			markercli.GetNewTransferCmd(),
			[]string{
				s.accountAddresses[0].String(),
				s.testnet.Validators[0].Address.String(),
				"hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"transfer, fail to transfer invalid coin count",
			markercli.GetNewTransferCmd(),
			[]string{
				s.accountAddresses[0].String(),
				s.testnet.Validators[0].Address.String(),
				"100hotdog,200koinz",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"transfer, successfully transfer",
			markercli.GetNewTransferCmd(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				s.accountAddresses[0].String(),
				"100hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"remove access",
			markercli.GetCmdDeleteAccess(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				"hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

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

func (s *IntegrationTestSuite) TestMarkerIbcTransfer() {
	testCases := []struct {
		name                       string
		srcPort                    string
		srcChannel                 string
		sender                     string
		receiver                   string
		amount                     string
		flagPacketTimeoutHeight    string
		flagPacketTimeoutTimestamp string
		flagAbsoluteTimeouts       string
		flagMemo                   string
		expectedErr                string
	}{
		{
			name:        "should fail on invalid coin",
			srcPort:     "port",
			srcChannel:  "channel",
			sender:      "sender",
			receiver:    "receiver",
			amount:      "not-a-valid-coin-amount",
			expectedErr: "invalid decimal coin expression: not-a-valid-coin-amount",
		},
		{
			name:                    "should fail on invalid packet timeout height",
			srcPort:                 "port",
			srcChannel:              "channel",
			sender:                  "sender",
			receiver:                "receiver",
			amount:                  "10jackthecat",
			flagPacketTimeoutHeight: "invalidtimeoutheight",
			expectedErr:             "expected height string format: {revision}-{height}. Got: invalidtimeoutheight: invalid height",
		},
		{
			name:                 "should fail on parsing absolute timeouts boolean",
			srcPort:              "port",
			srcChannel:           "channel",
			sender:               "sender",
			receiver:             "receiver",
			amount:               "10jackthecat",
			flagAbsoluteTimeouts: "not-a-bool",
			expectedErr:          `invalid argument "not-a-bool" for "--absolute-timeouts" flag: strconv.ParseBool: parsing "not-a-bool": invalid syntax`,
		},
		{
			name:        "should pass basic validation with optional flag memo",
			srcPort:     "port",
			srcChannel:  "channel-1",
			sender:      "sender",
			receiver:    "receiver",
			amount:      "10jackthecat",
			flagMemo:    "testing",
			expectedErr: `rpc error: code = NotFound desc = rpc error: code = NotFound desc = port-id: port, channel-id: channel-1: channel not found: key not found`,
		},
		{
			name:        "should pass basic validation without optional flag memo",
			srcPort:     "port",
			srcChannel:  "channel-1",
			sender:      "sender",
			receiver:    "receiver",
			amount:      "10jackthecat",
			expectedErr: `rpc error: code = NotFound desc = rpc error: code = NotFound desc = port-id: port, channel-id: channel-1: channel not found: key not found`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			args := []string{
				tc.srcPort,
				tc.srcChannel,
				tc.sender,
				tc.receiver,
				tc.amount,
			}
			args = append(args, []string{fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			}...)
			if len(tc.flagPacketTimeoutHeight) > 0 {
				args = append(args, fmt.Sprintf("--%s=%s", markercli.FlagPacketTimeoutHeight, tc.flagPacketTimeoutHeight))
			}
			if len(tc.flagAbsoluteTimeouts) > 0 {
				args = append(args, fmt.Sprintf("--%s=%s", markercli.FlagAbsoluteTimeouts, tc.flagAbsoluteTimeouts))
			}
			if len(tc.flagMemo) > 0 {
				args = append(args, fmt.Sprintf("--%s=%s", markercli.FlagMemo, tc.flagMemo))
			}
			_, err := clitestutil.ExecTestCLICmd(clientCtx, markercli.GetIbcTransferTxCmd(), args)
			if len(tc.expectedErr) > 0 {
				s.Assert().EqualError(err, tc.expectedErr)
			} else {
				s.Assert().NoError(err, tc.name)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestMarkerAuthzTxCommands() {
	testCases := []struct {
		name         string
		args         []string
		expectedErr  string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			name: "successfully grant authz for account without allow list",
			args: []string{
				s.accountAddresses[1].String(),
				"transfer",
				fmt.Sprintf("--%s=%s", markercli.FlagTransferLimit, "10authzhotdog"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
			},
			expectedErr:  "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "successfully grant authz for account with allow list",
			args: []string{
				s.accountAddresses[1].String(),
				"transfer",
				fmt.Sprintf("--%s=%s", markercli.FlagAllowList, s.accountAddresses[0].String()),
				fmt.Sprintf("--%s=%s", markercli.FlagTransferLimit, "10authzhotdog"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
			},
			expectedErr:  "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "fail to grant authz for account invalid allow list address",
			args: []string{
				s.accountAddresses[1].String(),
				"transfer",
				fmt.Sprintf("--%s=%s", markercli.FlagAllowList, "invalid"),
				fmt.Sprintf("--%s=%s", markercli.FlagTransferLimit, "10authzhotdog"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
			},
			expectedErr:  "decoding bech32 failed: invalid bech32 string length 7",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "fail to grant authz for account invalid denom for transfer limit",
			args: []string{
				s.accountAddresses[1].String(),
				"transfer",
				fmt.Sprintf("--%s=%s", markercli.FlagTransferLimit, "invalid"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
			},
			expectedErr:  "invalid decimal coin expression: invalid",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "fail to grant authz for account invalid action type",
			args: []string{
				s.accountAddresses[1].String(),
				"invalid",
				fmt.Sprintf("--%s=%s", markercli.FlagTransferLimit, "10authzhotdog"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
			},
			expectedErr:  "invalid authorization type, invalid",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)

			tc.args = append(tc.args, fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation))
			tc.args = append(tc.args, fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock))
			tc.args = append(tc.args, fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()))

			out, err := clitestutil.ExecTestCLICmd(clientCtx, markercli.GetCmdGrantAuthorization(), tc.args)
			if len(tc.expectedErr) > 0 {
				s.Assert().EqualError(err, tc.expectedErr)
			} else {
				s.Assert().NoError(err)
				s.Assert().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Assert().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestMarkerTxGovProposals() {
	testCases := []struct {
		name         string
		proposaltype string
		proposal     string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"invalid proposal type",
			"Invalid",
			"",
			true, &sdk.TxResponse{}, 0,
		},
		{
			"invalid proposal json",
			"Invalid",
			`{"title":"test add marker","description"`,
			true, &sdk.TxResponse{}, 0,
		},
		{
			"mint marker proposal",
			"IncreaseSupply",
			fmt.Sprintf(`{"title":"test mint marker","description":"description","manager":"%s",
			"amount":{"denom":"propcoin","amount":"10"}}`, s.testnet.Validators[0].Address.String()),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"burn marker proposal",
			"DecreaseSupply",
			fmt.Sprintf(`{"title":"test burn marker","description":"description","manager":"%s",
			"amount":{"denom":"propcoin","amount":"10"}}`, s.testnet.Validators[0].Address.String()),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"change status marker proposal",
			"ChangeStatus",
			`{"title":"test change marker status","description":"description","denom":"propcoin",
			"new_status":"MARKER_STATUS_CANCELLED"}`,
			false, &sdk.TxResponse{}, 0,
		},
		{
			"add admin marker proposal",
			"SetAdministrator",
			fmt.Sprintf(`{"title":"test add admin to marker","description":"description",
			"denom":"propcoin","access":[{"address":"%s", "permissions": [1,2,3,4,5,6]}]}`,
				s.testnet.Validators[0].Address.String()),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"remove admin marker proposal",
			"RemoveAdministrator",
			fmt.Sprintf(`{"title":"test remove marker admin","description":"description",
			"denom":"propcoin","removed_address":["%s"]}`,
				s.testnet.Validators[0].Address.String()),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"withdraw escrow marker proposal",
			"WithdrawEscrow",
			fmt.Sprintf(`{"title":"test withdraw marker","description":"description","target_address":"%s",
			"denom":"%s", "amount":[{"denom":"%s","amount":"1"}]}`, s.testnet.Validators[0].Address.String(),
				s.cfg.BondDenom, s.cfg.BondDenom),
			false, &sdk.TxResponse{}, 0x9,
			// The gov module now has its own set of errors.
			// This /should/ fail due to insufficient funds, and it does, but then the gov module erroneously wraps it again.
			// Insufficient funds is 0x5 in the main SDK's set of errors.
			// However, the governance module erroneously wraps this error in a 0x9, "no handler exists for proposal type"
			// So we're looking for a 0x9 here.
			// Here's the expected error (from the rawlog):
			// 	0stake is smaller than 1stake: insufficient funds: invalid proposal content: no handler exists for proposal type
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			p, err := ioutil.TempFile("", "*")
			tmpFile := p.Name()

			s.Require().NoError(err)
			_, err = p.WriteString(tc.proposal)
			s.Require().NoError(err)
			s.Require().NoError(p.Sync())
			s.Require().NoError(p.Close())

			args := []string{
				tc.proposaltype,
				tmpFile,
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=%s", flags.FlagGas, "500000"),
			}
			s.T().Logf("args: %q", args)
			out, err := clitestutil.ExecTestCLICmd(clientCtx, markercli.GetCmdMarkerProposal(), args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, txResp.RawLog)
			}

			s.Require().NoError(os.Remove(tmpFile))
		})
	}
}

func (s *IntegrationTestSuite) TestPaginationWithPageKey() {
	asJson := fmt.Sprintf("--%s=json", tmcli.OutputFlag)

	// Because other tests might have run before this and added markers,
	// the s.markerCount variable isn't necessarily how many markers exist right now.
	// So we'll do a quick AllMarkersCmd query to count them all for us.
	cout, cerr := clitestutil.ExecTestCLICmd(
		s.testnet.Validators[0].ClientCtx,
		markercli.AllMarkersCmd(),
		[]string{limitArg(1), asJson, "--count-total"},
	)
	s.Require().NoError(cerr, "count marker cmd error")
	var cresult markertypes.QueryAllMarkersResponse
	mmerr := s.cfg.Codec.UnmarshalJSON(cout.Bytes(), &cresult)
	s.Require().NoError(mmerr, "count marker unmarshal error")
	s.Require().Greater(cresult.Pagination.Total, uint64(0), "count markers pagination total")
	s.markerCount = int(cresult.Pagination.Total)

	s.T().Run("AllMarkersCmd", func(t *testing.T) {
		// Choosing page size = 7 because it a) isn't the default, b) doesn't evenly divide 20.
		pageSize := 7
		expectedCount := s.markerCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]markertypes.MarkerAccount, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{pageSizeArg, asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := markercli.AllMarkersCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result markertypes.QueryAllMarkersResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultMarkerCount := len(result.Markers)
			if page != pageCount {
				require.Equalf(t, pageSize, resultMarkerCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultMarkerCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = appendMarkers(results, result.Markers...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of markers returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(markerSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two markers should be equal here")
		}
	})

	s.T().Run("AllHoldersCmd denom", func(t *testing.T) {
		// Choosing page size = 3 because it a) isn't the default, b) doesn't evenly divide 4.
		pageSize := 3
		expectedCount := s.holderCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]markertypes.Balance, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{s.holderDenom, pageSizeArg, asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := markercli.AllHoldersCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result markertypes.QueryHoldingResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultCount := len(result.Balances)
			if page != pageCount {
				require.Equalf(t, pageSize, resultCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = append(results, result.Balances...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of balances returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Sort(balanceSorter(results))
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two balances should be equal here")
		}
	})
}

func getFormattedExpiration(duration int64) string {
	return time.Now().Add(time.Duration(duration) * time.Second).Format(time.RFC3339)
}

func (s *IntegrationTestSuite) TestAddFinalizeActivateMarkerTxCommands() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"create a new marker, finalize it and activate it.",
			markercli.GetCmdAddFinalizeActivateMarker(),
			[]string{
				"1000newhotdog",
				getAccessGrantString(s.testnet.Validators[0].Address, s.accountAddresses[1]),
				fmt.Sprintf("--%s=%s", markercli.FlagType, "RESTRICTED"),
				fmt.Sprintf("--%s=%s", markercli.FlagSupplyFixed, "true"),
				fmt.Sprintf("--%s=%s", markercli.FlagAllowGovernanceControl, "true"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"create a new marker, finalize it and activate it, with access grant to one address ",
			markercli.GetCmdAddFinalizeActivateMarker(),
			[]string{
				"1000newhotdog1",
				getAccessGrantString(s.testnet.Validators[0].Address, nil),
				fmt.Sprintf("--%s=%s", markercli.FlagType, "RESTRICTED"),
				fmt.Sprintf("--%s=%s", markercli.FlagSupplyFixed, "true"),
				fmt.Sprintf("--%s=%s", markercli.FlagAllowGovernanceControl, "true"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"create a new marker with no access grant ",
			markercli.GetCmdAddFinalizeActivateMarker(),
			[]string{
				"1000newhotdog1",
				fmt.Sprintf("--%s=%s", markercli.FlagType, "RESTRICTED"),
				fmt.Sprintf("--%s=%s", markercli.FlagSupplyFixed, "true"),
				fmt.Sprintf("--%s=%s", markercli.FlagAllowGovernanceControl, "true"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"create a new marker, finalize it and activate it  with dashes and periods",
			markercli.GetCmdAddFinalizeActivateMarker(),
			[]string{
				"1000newcat-scratch-fever.bobcat",
				getAccessGrantString(s.testnet.Validators[0].Address, nil),
				fmt.Sprintf("--%s=%s", markercli.FlagType, "RESTRICTED"),
				fmt.Sprintf("--%s=%s", markercli.FlagSupplyFixed, "true"),
				fmt.Sprintf("--%s=%s", markercli.FlagAllowGovernanceControl, "true"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"fail to create add/finalize/activate marker, incorrect allow governance value",
			markercli.GetCmdAddFinalizeActivateMarker(),
			[]string{
				"1000hotdog",
				getAccessGrantString(s.testnet.Validators[0].Address, nil),
				fmt.Sprintf("--%s=%s", markercli.FlagType, "RESTRICTED"),
				fmt.Sprintf("--%s=%s", markercli.FlagSupplyFixed, "false"),
				fmt.Sprintf("--%s=%s", markercli.FlagAllowGovernanceControl, "wrong"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"fail to create add/finalize/activate marker, incorrect supply fixed value",
			markercli.GetCmdAddFinalizeActivateMarker(),
			[]string{
				"1000hotdog",
				getAccessGrantString(s.testnet.Validators[0].Address, nil),
				fmt.Sprintf("--%s=%s", markercli.FlagType, "RESTRICTED"),
				fmt.Sprintf("--%s=%s", markercli.FlagSupplyFixed, "wrong"),
				fmt.Sprintf("--%s=%s", markercli.FlagAllowGovernanceControl, "true"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

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

func (s *IntegrationTestSuite) TestUpdateRequiredAttributesTxCommand() {

	testCases := []struct {
		name          string
		cmd           *cobra.Command
		args          []string
		expectedError string
	}{
		{
			name: "should fail, both add and remove lists are empty",
			cmd:  markercli.GetCmdUpdateRequiredAttributes(),
			args: []string{
				"newhotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedError: "both add and remove lists cannot be empty",
		},
		{
			name: "should fail, invalid gov proposal no deposit set",
			cmd:  markercli.GetCmdUpdateRequiredAttributes(),
			args: []string{
				"newhotdog",
				fmt.Sprintf("--%s=%s", markercli.FlagGovProposal, "true"),
				fmt.Sprintf("--%s=%s", markercli.FlagAdd, "foo.provenance.io"),
				fmt.Sprintf("--%s=%s", markercli.FlagRemove, "bar.provenance.io"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedError: "deposit for gov proposal was not set.  Use deposit flag to set deposit",
		},
		{
			name: "should fail, invalid gov proposal deposit denom",
			cmd:  markercli.GetCmdUpdateRequiredAttributes(),
			args: []string{
				"newhotdog",
				fmt.Sprintf("--%s=%s", markercli.FlagGovProposal, "true"),
				fmt.Sprintf("--%s=%s", markercli.FlagAdd, "foo.provenance.io"),
				fmt.Sprintf("--%s=%s", markercli.FlagRemove, "bar.provenance.io"),
				fmt.Sprintf("--%s=%s", markercli.FlagDeposit, "blah"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedError: "invalid decimal coin expression: blah",
		},
		{
			name: "should succeed, gov proposal should succeed",
			cmd:  markercli.GetCmdUpdateRequiredAttributes(),
			args: []string{
				"newhotdog",
				fmt.Sprintf("--%s=%s", markercli.FlagGovProposal, "true"),
				fmt.Sprintf("--%s=%s", markercli.FlagAdd, "foo.provenance.io"),
				fmt.Sprintf("--%s=%s", markercli.FlagRemove, "bar.provenance.io"),
				fmt.Sprintf("--%s=%s", markercli.FlagDeposit, "100jackthecat"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
		},
		{
			name: "should succeed, send regular tx",
			cmd:  markercli.GetCmdUpdateRequiredAttributes(),
			args: []string{
				"newhotdog",
				fmt.Sprintf("--%s=%s", markercli.FlagAdd, "foo.provenance.io"),
				fmt.Sprintf("--%s=%s", markercli.FlagRemove, "bar.provenance.io"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx

			_, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)
			if len(tc.expectedError) > 0 {
				s.Require().EqualError(err, tc.expectedError)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func getAccessGrantString(address sdk.AccAddress, anotherAddress sdk.AccAddress) string {
	if anotherAddress != nil {
		accessGrant := address.String() + ",mint,admin;" + anotherAddress.String() + ",burn"
		return accessGrant
	} else {
		accessGrant := address.String() + ",mint,admin;"
		return accessGrant
	}
}

func (s *IntegrationTestSuite) TestParseAccessGrantFromString() {
	testCases := []struct {
		name              string
		accessGrantString string
		expectPanic       bool
		expectedResult    []types.AccessGrant
	}{
		{
			"successfully parses empty string",
			"",
			false,
			[]types.AccessGrant{},
		},
		{
			"fails parsing invalid string",
			"blah",
			true,
			[]types.AccessGrant{},
		},
		{
			"should fail empty list of permissions",
			",,;",
			true,
			[]types.AccessGrant{},
		},
		{
			"should fail address is not valid",
			"NotAnAddress,mint;",
			true,
			[]types.AccessGrant{},
		},
		{
			"should succeed to add access type",
			fmt.Sprintf("%s,mint;", s.accountAddresses[0].String()),
			false,
			[]types.AccessGrant{markertypes.AccessGrant{Address: s.accountAddresses[0].String(), Permissions: []markertypes.Access{markertypes.Access_Mint}}},
		},
		{
			"should succeed to add access type",
			fmt.Sprintf("%s,mint;", s.accountAddresses[0].String()),
			false,
			[]types.AccessGrant{markertypes.AccessGrant{Address: s.accountAddresses[0].String(), Permissions: []markertypes.Access{markertypes.Access_Mint}}},
		},
	}
	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			if tc.expectPanic {
				panicFunc := func() { markercli.ParseAccessGrantFromString(tc.accessGrantString) }
				s.Assert().Panics(panicFunc)

			} else {
				result := markercli.ParseAccessGrantFromString(tc.accessGrantString)
				s.Assert().ElementsMatch(result, tc.expectedResult)
			}
		})
	}
}

func TestParseNewMarkerFlags(t *testing.T) {
	getTestCmd := func() *cobra.Command {
		cmd := &cobra.Command{
			Use: "testing",
			Run: func(cmd *cobra.Command, args []string) {
				t.Logf("test command called with args: '%s'", strings.Join(args, "', '"))
			},
		}
		markercli.AddNewMarkerFlags(cmd)
		return cmd
	}

	argType := "--" + markercli.FlagType
	argFixed := "--" + markercli.FlagSupplyFixed
	argGov := "--" + markercli.FlagAllowGovernanceControl
	argForce := "--" + markercli.FlagAllowForceTransfer
	argRequiredAtt := "--" + markercli.FlagRequiredAttributes

	tests := []struct {
		name   string
		cmd    *cobra.Command
		args   []string
		exp    *markercli.NewMarkerFlagValues
		expErr []string
	}{
		{
			name: "no args",
			cmd:  getTestCmd(),
			args: []string{},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        false,
				AllowGovControl:    false,
				AllowForceTransfer: false,
				RequiredAttributes: []string{},
			},
		},
		{
			name: "coin type",
			cmd:  getTestCmd(),
			args: []string{argType, "COIN"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        false,
				AllowGovControl:    false,
				AllowForceTransfer: false,
				RequiredAttributes: []string{},
			},
		},
		{
			name: "restricted type upper",
			cmd:  getTestCmd(),
			args: []string{argType, "RESTRICTED"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_RestrictedCoin,
				SupplyFixed:        false,
				AllowGovControl:    false,
				AllowForceTransfer: false,
				RequiredAttributes: []string{},
			},
		},
		{
			name: "restricted type lower",
			cmd:  getTestCmd(),
			args: []string{argType, "restricted"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_RestrictedCoin,
				SupplyFixed:        false,
				AllowGovControl:    false,
				AllowForceTransfer: false,
				RequiredAttributes: []string{},
			},
		},
		{
			name:   "other type",
			cmd:    getTestCmd(),
			args:   []string{argType, "OOPS"},
			expErr: []string{"invalid marker type", "OOPS", "COIN", "RESTRICTED"},
		},
		{
			name: "supply fixed flag no value",
			cmd:  getTestCmd(),
			args: []string{argFixed},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        true,
				AllowGovControl:    false,
				AllowForceTransfer: false,
				RequiredAttributes: []string{},
			},
		},
		{
			name: "supply fixed true",
			cmd:  getTestCmd(),
			args: []string{argFixed + "=true"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        true,
				AllowGovControl:    false,
				AllowForceTransfer: false,
				RequiredAttributes: []string{},
			},
		},
		{
			name: "supply fixed false",
			cmd:  getTestCmd(),
			args: []string{argFixed + "=false"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        false,
				AllowGovControl:    false,
				AllowForceTransfer: false,
				RequiredAttributes: []string{},
			},
		},
		{
			name: "gov control flag no value",
			cmd:  getTestCmd(),
			args: []string{argGov},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        false,
				AllowGovControl:    true,
				AllowForceTransfer: false,
				RequiredAttributes: []string{},
			},
		},
		{
			name: "gov control true",
			cmd:  getTestCmd(),
			args: []string{argGov + "=true"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        false,
				AllowGovControl:    true,
				AllowForceTransfer: false,
				RequiredAttributes: []string{},
			},
			expErr: nil,
		},
		{
			name: "gov control false",
			cmd:  getTestCmd(),
			args: []string{argGov + "=false"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        false,
				AllowGovControl:    false,
				AllowForceTransfer: false,
				RequiredAttributes: []string{},
			},
			expErr: nil,
		},
		{
			name: "force transfer flag no value",
			cmd:  getTestCmd(),
			args: []string{argForce},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        false,
				AllowGovControl:    false,
				AllowForceTransfer: true,
				RequiredAttributes: []string{},
			},
		},
		{
			name: "force transfer true",
			cmd:  getTestCmd(),
			args: []string{argForce + "=true"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        false,
				AllowGovControl:    false,
				AllowForceTransfer: true,
				RequiredAttributes: []string{},
			},
		},
		{
			name: "force transfer false",
			cmd:  getTestCmd(),
			args: []string{argForce + "=false"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        false,
				AllowGovControl:    false,
				AllowForceTransfer: false,
				RequiredAttributes: []string{},
			},
		},
		{
			name: "required attributes present",
			cmd:  getTestCmd(),
			args: []string{argRequiredAtt + "=jack.the.cat.io,george.the.dog.io"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        false,
				AllowGovControl:    false,
				AllowForceTransfer: false,
				RequiredAttributes: []string{"jack.the.cat.io", "george.the.dog.io"},
			},
		},
		{
			name: "everything",
			cmd:  getTestCmd(),
			args: []string{argForce, argGov, argType, "RESTRICTED", argFixed, argRequiredAtt, "jack.the.cat.io,george.the.dog.io"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_RestrictedCoin,
				SupplyFixed:        true,
				AllowGovControl:    true,
				AllowForceTransfer: true,
				RequiredAttributes: []string{"jack.the.cat.io", "george.the.dog.io"},
			},
		},
		// Note: I can't figure out a way to make cmd.Flags().GetBool return an error.
		// If someone provides an invalid value, e.g. --supplyFixed=bananas, it will fail
		// when the flags are being parsed by cobra, which is before they'd be accessed in ParseFlags.
		// That's why such test cases have been omitted.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, tc.cmd.ParseFlags(tc.args), "applying args to command")
			actual, err := markercli.ParseNewMarkerFlags(tc.cmd)
			if len(tc.expErr) != 0 {
				if assert.Error(t, err, "ParseNewMarkerFlags") {
					for _, exp := range tc.expErr {
						assert.ErrorContainsf(t, err, exp, "ParseNewMarkerFlags")
					}
				}
			} else {
				require.NoError(t, err, "ParseNewMarkerFlags")
				assert.Equal(t, tc.exp, actual, "NewMarkerFlagValues from '%s'", strings.Join(tc.args, "', '"))
			}
		})
	}
}
