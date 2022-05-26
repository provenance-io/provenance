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
	"github.com/provenance-io/provenance/testutil"
	markercli "github.com/provenance-io/provenance/x/marker/client/cli"
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
	kr, err := keyring.New(s.T().Name(), "test", s.keyringDir, nil)
	s.Require().NoError(err)
	s.keyring = kr
	for i := 0; i < number; i++ {
		keyId := fmt.Sprintf("test_key%v", i)
		info, _, err := kr.NewMnemonic(keyId, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err)
		s.accountAddresses = append(s.accountAddresses, info.GetAddress())
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultTestNetworkConfig()

	genesisState := cfg.GenesisState
	cfg.NumValidators = 1
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
			},
		)
	}
	markerDataBz, err := cfg.Codec.MarshalJSON(&markerData)
	s.Require().NoError(err)
	genesisState[markertypes.ModuleName] = markerDataBz

	cfg.GenesisState = genesisState

	s.cfg = cfg
	cfg.ChainID = antewrapper.SimAppChainID

	s.testnet = testnet.New(s.T(), cfg)

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err)
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
			`{"marker":{"@type":"/provenance.marker.v1.MarkerAccount","base_account":{"address":"cosmos1p3sl9tll0ygj3flwt5r2w0n6fx9p5ngq2tu6mq","pub_key":null,"account_number":"11","sequence":"0"},"manager":"","access_control":[],"status":"MARKER_STATUS_ACTIVE","denom":"testcoin","supply":"1000","marker_type":"MARKER_TYPE_COIN","supply_fixed":true,"allow_governance_control":false}}`,
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
  allow_governance_control: false
  base_account:
    account_number: "11"
    address: cosmos1p3sl9tll0ygj3flwt5r2w0n6fx9p5ngq2tu6mq
    pub_key: null
    sequence: "0"
  denom: testcoin
  manager: ""
  marker_type: MARKER_TYPE_COIN
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
			`{"marker":{"@type":"/provenance.marker.v1.MarkerAccount","base_account":{"address":"cosmos16437wt0xtqtuw0pn4vt8rlf8gr2plz2det0mt2","pub_key":null,"account_number":"12","sequence":"0"},"manager":"","access_control":[],"status":"MARKER_STATUS_ACTIVE","denom":"lockedcoin","supply":"1000","marker_type":"MARKER_TYPE_RESTRICTED","supply_fixed":true,"allow_governance_control":false}}`,
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
				s.Require().NoError(clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestMarkerAuthzTxCommands() {
	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"grant authz transfer permissions to account 1 for account 0 tranfer limit of 10",
			markercli.GetCmdGrantAuthorization(),
			[]string{
				s.accountAddresses[1].String(),
				"transfer",
				fmt.Sprintf("--%s=%s", markercli.FlagTransferLimit, "10authzhotdog"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"grant authz transfer permissions to account 0 for account 1 transfer limit 20",
			markercli.GetCmdGrantAuthorization(),
			[]string{
				s.accountAddresses[0].String(),
				"transfer",
				fmt.Sprintf("--%s=%s", markercli.FlagTransferLimit, "20authzhotdog"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[1].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"marker transfer successful, account 1 as grantee and account 0 as granter",
			markercli.GetNewTransferCmd(),
			[]string{
				s.accountAddresses[0].String(),
				s.accountAddresses[1].String(),
				"4authzhotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[1].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"marker transfer failed,  account 1 over transfer limit",
			markercli.GetNewTransferCmd(),
			[]string{
				s.accountAddresses[0].String(),
				s.accountAddresses[1].String(),
				"7authzhotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[1].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 5,
		},
		{
			"marker transfer failed, account 1 not granted rights by account 2",
			markercli.GetNewTransferCmd(),
			[]string{
				s.accountAddresses[2].String(),
				s.accountAddresses[1].String(),
				"9authzhotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[1].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 1,
		},
		{
			"grantee successful transfer, removed from auth for reaching transfer limit",
			markercli.GetNewTransferCmd(),
			[]string{
				s.accountAddresses[1].String(),
				s.accountAddresses[0].String(),
				"20authzhotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"revoke authz transfer from grantee",
			markercli.GetCmdRevokeAuthorization(),
			[]string{
				s.accountAddresses[1].String(),
				"transfer",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"transfer should fail, due to account 1's revoked access",
			markercli.GetNewTransferCmd(),
			[]string{
				s.accountAddresses[0].String(),
				s.accountAddresses[1].String(),
				"1hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[1].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 1,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
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
			"add marker proposal",
			"AddMarker",
			fmt.Sprintf(`{"title":"test add marker","description":"description","manager":"%s",
			"amount":{"denom":"testpropmarker","amount":"1"},"status":"MARKER_STATUS_ACTIVE","marker_type":1,
			"supply_fixed":true,"allow_governance_control":true}`, s.testnet.Validators[0].Address.String()),
			false, &sdk.TxResponse{}, 0,
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
			false, &sdk.TxResponse{}, 0x5, // request is good, NSF on account though
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
			}
			out, err := clitestutil.ExecTestCLICmd(clientCtx, markercli.GetCmdMarkerProposal(), args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, txResp.Logs.String())
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
