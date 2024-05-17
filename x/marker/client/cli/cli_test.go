package cli_test

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtcli "github.com/cometbft/cometbft/libs/cli"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdktypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/testutil/assertions"
	testcli "github.com/provenance-io/provenance/testutil/cli"
	"github.com/provenance-io/provenance/testutil/queries"
	attrcli "github.com/provenance-io/provenance/x/attribute/client/cli"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
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
	govv1.DefaultMinDepositRatio = sdkmath.LegacyZeroDec()
	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.NumValidators = 1
	s.cfg.ChainID = antewrapper.SimAppChainID
	s.GenerateAccountsWithKeyrings(4)

	s.holderDenom = "hodlercoin"
	s.holderCount = 4
	s.markerCount = 20

	// Configure Genesis auth data for adding test accounts
	testutil.MutateGenesisState(s.T(), &s.cfg, authtypes.ModuleName, &authtypes.GenesisState{}, func(authData *authtypes.GenesisState) *authtypes.GenesisState {
		var genAccounts []authtypes.GenesisAccount
		authData.Params = authtypes.DefaultParams()
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[0], nil, 3, 0))
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[1], nil, 4, 0))
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[2], nil, 5, 0))
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[3], nil, 6, 0))
		accounts, err := authtypes.PackAccounts(genAccounts)
		s.Require().NoError(err)
		authData.Accounts = accounts
		return authData
	})

	// Configure Genesis bank data for test accounts
	testutil.MutateGenesisState(s.T(), &s.cfg, banktypes.ModuleName, &banktypes.GenesisState{}, func(bankGenState *banktypes.GenesisState) *banktypes.GenesisState {
		bondCoin := sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens)
		bal := func(addr sdk.AccAddress, coins ...sdk.Coin) banktypes.Balance {
			return banktypes.Balance{
				Address: addr.String(),
				Coins:   sdk.NewCoins(coins...),
			}
		}
		coin := func(amount int64, denom string) sdk.Coin {
			return sdk.NewInt64Coin(denom, amount)
		}

		bankGenState.Balances = append(bankGenState.Balances,
			bal(s.accountAddresses[0], bondCoin, coin(100, "authzhotdog"), coin(123, s.holderDenom)),
			bal(s.accountAddresses[1], bondCoin, coin(100, "authzhotdog"), coin(234, s.holderDenom)),
			bal(s.accountAddresses[2], bondCoin, coin(345, s.holderDenom)),
			bal(s.accountAddresses[3], bondCoin, coin(456, s.holderDenom)),

			bal(markertypes.MustGetMarkerAddress("testcoin"), coin(1000, "testcoin")),
			bal(markertypes.MustGetMarkerAddress("lockedcoin"), coin(1000, "lockedcoin")),
			bal(markertypes.MustGetMarkerAddress("propcoin"), coin(1000, "propcoin")),
			bal(markertypes.MustGetMarkerAddress("authzhotdog"), coin(800, "authzhotdog")),
		)

		return bankGenState
	})

	// Configure Genesis data for marker module
	testutil.MutateGenesisState(s.T(), &s.cfg, markertypes.ModuleName, &markertypes.GenesisState{}, func(markerData *markertypes.GenesisState) *markertypes.GenesisState {
		markerData.Params.EnableGovernance = true
		markerData.Params.MaxTotalSupply = 1000000
		markerData.Params.MaxSupply = sdkmath.NewInt(1000000)

		// Define some specific markers to use in the tests.
		// Note: These account numbers get ignored.
		newMarkers := []markertypes.MarkerAccount{
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
				Supply:                 sdkmath.NewInt(1000),
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
				Supply:                 sdkmath.NewInt(1000),
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
				Supply:                 sdkmath.NewInt(1000),
				Denom:                  "propcoin",
				AllowForcedTransfer:    false,
			},
			{
				BaseAccount: &authtypes.BaseAccount{
					Address:       markertypes.MustGetMarkerAddress(s.cfg.BondDenom).String(),
					AccountNumber: 130,
					Sequence:      0,
				},
				Status:                 markertypes.StatusActive,
				SupplyFixed:            false,
				MarkerType:             markertypes.MarkerType_Coin,
				AllowGovernanceControl: true,
				Supply:                 s.cfg.BondedTokens.MulRaw(int64(s.cfg.NumValidators)),
				Denom:                  s.cfg.BondDenom,
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
				Supply:                 sdkmath.NewInt(1000),
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
					Address:       markertypes.MustGetMarkerAddress(s.holderDenom).String(),
					AccountNumber: 150,
					Sequence:      0,
				},
				Status:                 markertypes.StatusActive,
				SupplyFixed:            false,
				MarkerType:             markertypes.MarkerType_RestrictedCoin,
				AllowGovernanceControl: false,
				Supply:                 sdkmath.NewInt(3000),
				Denom:                  s.holderDenom,
				AllowForcedTransfer:    true,
			},
		}
		markerData.Markers = append(markerData.Markers, newMarkers...)

		// And define a NAV for each new marker.
		for _, marker := range newMarkers {
			var mNav types.MarkerNetAssetValues
			mNav.Address = marker.GetAddress().String()
			mNav.NetAssetValues = []types.NetAssetValue{types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 100), 100)}
			markerData.NetAssetValues = append(markerData.NetAssetValues, mNav)
		}

		// Now create more markers (and their navs) until we have s.markerCount of them.
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
					Supply:                 sdkmath.NewInt(int64(i * 100000)),
					Denom:                  denom,
					AllowForcedTransfer:    false,
				},
			)
			var mNav types.MarkerNetAssetValues
			mNav.Address = markertypes.MustGetMarkerAddress(denom).String()
			mNav.NetAssetValues = []types.NetAssetValue{types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 100), 100)}
			markerData.NetAssetValues = append(markerData.NetAssetValues, mNav)
		}

		return markerData
	})

	// Pre-define an accountdata entry
	testutil.MutateGenesisState(s.T(), &s.cfg, attrtypes.ModuleName, &attrtypes.GenesisState{}, func(attrData *attrtypes.GenesisState) *attrtypes.GenesisState {
		attrData.Attributes = append(attrData.Attributes,
			attrtypes.Attribute{
				Name:          attrtypes.AccountDataName,
				Value:         []byte("Do not sell this coin."),
				AttributeType: attrtypes.AttributeType_String,
				Address:       markertypes.MustGetMarkerAddress(s.holderDenom).String(),
			},
		)
		return attrData
	})

	var err error
	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "creating testnet")

	_, err = testutil.WaitForHeight(s.testnet, 1)
	s.Require().NoError(err, "waiting for height 1")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	testutil.Cleanup(s.testnet, s.T())
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
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			},
			`{"max_total_supply":"1000000","enable_governance":true,"unrestricted_denom_regex":"[a-zA-Z][a-zA-Z0-9\\-\\.]{2,83}","max_supply":"1000000"}`,
		},
		{
			"get testcoin marker json",
			markercli.MarkerCmd(),
			[]string{
				"testcoin",
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			},
			`{"marker":{"@type":"/provenance.marker.v1.MarkerAccount","base_account":{"address":"cosmos1p3sl9tll0ygj3flwt5r2w0n6fx9p5ngq2tu6mq","pub_key":null,"account_number":"8","sequence":"0"},"manager":"","access_control":[],"status":"MARKER_STATUS_ACTIVE","denom":"testcoin","supply":"1000","marker_type":"MARKER_TYPE_COIN","supply_fixed":true,"allow_governance_control":false,"allow_forced_transfer":false,"required_attributes":[]}}`,
		},
		{
			"get testcoin marker test",
			markercli.MarkerCmd(),
			[]string{
				"testcoin",
				fmt.Sprintf("--%s=text", cmtcli.OutputFlag),
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
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
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
  denom: ` + s.holderDenom + `
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
			fmt.Sprintf("amount:\n  amount: \"%s\"\n  denom: %s", s.cfg.BondedTokens.MulRaw(int64(s.cfg.NumValidators)), s.cfg.BondDenom),
		},
		{
			name:           "account data",
			cmd:            markercli.AccountDataCmd(),
			args:           []string{s.holderDenom},
			expectedOutput: "value: Do not sell this coin.",
		},
		{
			name:           "marker net asset value query",
			cmd:            markercli.NetAssetValuesCmd(),
			args:           []string{"testcoin"},
			expectedOutput: "net_asset_values:\n- price:\n    amount: \"100\"\n    denom: usd\n  updated_block_height: \"0\"\n  volume: \"100\"",
		},
	}
	for _, tc := range testCases {
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", markercli.FlagSpendLimit, sdk.NewInt64Coin("stake", 100)),
				fmt.Sprintf("--%s=%s", markercli.FlagExpiration, getFormattedExpiration(oneYear)),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"add multiple access",
			markercli.GetCmdAddAccess(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				"hotdog",
				"mint,burn,transfer,withdraw,deposit",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", markercli.FlagSpendLimit, sdk.NewInt64Coin("stake", 100)),
				fmt.Sprintf("--%s=%s", markercli.FlagExpiration, getFormattedExpiration(oneYear)),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"create periodic feegrant allowance",
			markercli.GetCmdFeeGrant(),
			[]string{
				"hotdog",
				s.testnet.Validators[0].Address.String(),
				s.accountAddresses[1].String(),
				fmt.Sprintf("--%s=%v", markercli.FlagPeriod, oneHour),
				fmt.Sprintf("--%s=%s", markercli.FlagPeriodLimit, sdk.NewInt64Coin("stake", 100)),
				fmt.Sprintf("--%s=%s", markercli.FlagExpiration, getFormattedExpiration(oneYear)),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			name: "set account data",
			cmd:  markercli.GetCmdSetAccountData(),
			args: []string{
				"hotdog",
				fmt.Sprintf("--%s", attrcli.FlagValue), "Not as good as corndog.",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			"remove access",
			markercli.GetCmdDeleteAccess(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				"hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			name: "set account data via gov prop",
			cmd:  markercli.GetCmdSetAccountData(),
			args: []string{
				"hotdog",
				fmt.Sprintf("--%s", attrcli.FlagValue), "Better than corndog.",
				fmt.Sprintf("--%s", markercli.FlagGovProposal),
				"--title", "Set hotdog account data", "--summary", "Something unique to help identify this proposal. B65B",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			testcli.NewTxExecutor(tc.cmd, tc.args).
				WithExpErr(tc.expectErr).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}

	// Now check some stuff to make sure it actually happened.

	checks := []struct {
		name   string
		cmd    *cobra.Command
		args   []string
		expOut []string
	}{
		{
			name:   "get account data with marker command",
			cmd:    markercli.AccountDataCmd(),
			args:   []string{"hotdog"},
			expOut: []string{"value: Not as good as corndog."},
		},
		{
			name:   "get account data with attribute command",
			cmd:    attrcli.GetAccountDataCmd(),
			args:   []string{markertypes.MustGetMarkerAddress("hotdog").String()},
			expOut: []string{"value: Not as good as corndog."},
		},
		{
			name: "gov prop created for account data cmd",
			cmd:  queries.CmdGetAllGovProps(s.testnet),
			expOut: []string{
				"'@type': /provenance.marker.v1.MsgSetAccountDataRequest",
				"denom: hotdog",
				"signer: " + authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				"value: Better than corndog.",
			},
		},
	}

	for _, check := range checks {
		s.Run(check.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			cmdName := check.cmd.Name()
			if check.args == nil {
				check.args = []string{}
			}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, check.cmd, check.args)
			outBz := out.Bytes()
			s.T().Logf("ExecTestCLICmd %q %q\nOutput:\n%s", cmdName, check.args, string(outBz))

			s.Require().NoError(err, "ExecTestCLICmd %s %q", cmdName, check.args)
			for _, exp := range check.expOut {
				s.Assert().Contains(string(outBz), exp, "%s command output", cmdName)
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
		s.Run(tc.name, func() {
			cmd := markercli.GetIbcTransferTxCmd()
			args := []string{
				tc.srcPort,
				tc.srcChannel,
				tc.sender,
				tc.receiver,
				tc.amount,
			}
			args = append(args, []string{fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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

			testcli.NewTxExecutor(cmd, args).
				WithExpErrMsg(tc.expectedErr).
				Execute(s.T(), s.testnet)
		})
	}
}

func (s *IntegrationTestSuite) TestMarkerAuthzTxCommands() {
	curClientCtx := s.testnet.Validators[0].ClientCtx
	defer func() {
		s.testnet.Validators[0].ClientCtx = curClientCtx
	}()
	s.testnet.Validators[0].ClientCtx = s.testnet.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)

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
		s.Run(tc.name, func() {
			cmd := markercli.GetCmdGrantAuthorization()
			tc.args = append(tc.args,
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			)
			testcli.NewTxExecutor(cmd, tc.args).
				WithExpErrMsg(tc.expectedErr).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}
}

func (s *IntegrationTestSuite) TestPaginationWithPageKey() {
	asJson := fmt.Sprintf("--%s=json", cmtcli.OutputFlag)

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
	getAccessGrantString := func(address sdk.AccAddress, anotherAddress sdk.AccAddress) string {
		if anotherAddress != nil {
			return address.String() + ",mint,admin,transfer;" + anotherAddress.String() + ",burn"
		}
		return address.String() + ",mint,admin,transfer;"
	}

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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			testcli.NewTxExecutor(tc.cmd, tc.args).
				WithExpErr(tc.expectErr).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			expectedError: "both add and remove lists cannot be empty",
		},
		{
			name: "should fail, invalid gov proposal deposit denom",
			cmd:  markercli.GetCmdUpdateRequiredAttributes(),
			args: []string{
				"newhotdog",
				fmt.Sprintf("--%s=%s", markercli.FlagGovProposal, "true"),
				fmt.Sprintf("--%s=%s", markercli.FlagAdd, "foo.provenance.io"),
				fmt.Sprintf("--%s=%s", markercli.FlagRemove, "bar.provenance.io"),
				fmt.Sprintf("--%s=%s", govcli.FlagDeposit, "blah"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			expectedError: "invalid deposit: invalid decimal coin expression: blah",
		},
		{
			name: "should succeed, gov proposal should succeed",
			cmd:  markercli.GetCmdUpdateRequiredAttributes(),
			args: []string{
				"newhotdog",
				fmt.Sprintf("--%s=%s", markercli.FlagGovProposal, "true"),
				"--title", "Update newhotdog req attrs", "--summary", "See title.",
				fmt.Sprintf("--%s=%s", markercli.FlagAdd, "foo.provenance.io"),
				fmt.Sprintf("--%s=%s", markercli.FlagRemove, "bar.provenance.io"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
		},
		{
			name: "should succeed, send regular tx",
			cmd:  markercli.GetCmdUpdateRequiredAttributes(),
			args: []string{
				"newhotdog",
				fmt.Sprintf("--%s=%s", markercli.FlagAdd, "foo.provenance.io"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			testcli.NewTxExecutor(tc.cmd, tc.args).
				WithExpErrMsg(tc.expectedError).
				Execute(s.T(), s.testnet)
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdUpdateForcedTransfer() {
	denom := "updateftcoin"
	s.Run("add a new marker for this", func() {
		cmd := markercli.GetCmdAddFinalizeActivateMarker()
		args := []string{
			"1000" + denom,
			s.testnet.Validators[0].Address.String() + ",mint,burn,deposit,withdraw,delete,admin,transfer",
			fmt.Sprintf("--%s=%s", markercli.FlagType, "RESTRICTED"),
			"--" + markercli.FlagSupplyFixed,
			"--" + markercli.FlagAllowGovernanceControl,
			fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
		}
		testcli.NewTxExecutor(cmd, args).Execute(s.T(), s.testnet)
	})
	if s.T().Failed() {
		s.FailNow("Stopping due to setup error")
	}

	argsWStdFlags := func(denom string, args ...string) []string {
		var rv []string
		rv = append(rv, denom)
		rv = append(rv, args...)
		rv = append(rv,
			"--title", "Update ft of "+denom, "--summary", "whatever",
			fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
		)
		return rv
	}

	tests := []struct {
		name        string
		args        []string
		expErr      string
		expCode     uint32
		expInRawLog string
	}{
		{
			name:        "invalid denom",
			args:        argsWStdFlags("x", "true"),
			expCode:     12,
			expInRawLog: "invalid denom: x",
		},
		{
			name:   "invalid bool",
			args:   argsWStdFlags(denom, "farse"),
			expErr: "invalid boolean string: \"farse\"",
		},
		{
			name: "set to true",
			args: argsWStdFlags(denom, "true"),
		},
		{
			name: "set to false",
			args: argsWStdFlags(denom, "false"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			txResp := testcli.NewTxExecutor(markercli.GetCmdUpdateForcedTransfer(), tc.args).
				WithExpErrMsg(tc.expErr).
				WithExpCode(tc.expCode).
				Execute(s.T(), s.testnet)

			if txResp != nil && txResp.Code == 0 {
				expAttrs := []abci.EventAttribute{
					{
						Key:   "action",
						Value: "/cosmos.gov.v1.MsgSubmitProposal",
						Index: true,
					},
					{
						Key:   "proposal_messages",
						Value: ",/provenance.marker.v1.MsgUpdateForcedTransferRequest",
						Index: true,
					},
				}

				var actAttrs []abci.EventAttribute
				for _, event := range txResp.Events {
					actAttrs = append(actAttrs, event.Attributes...)
				}

				var missingAttrs []abci.EventAttribute
				for _, exp := range expAttrs {
					if !s.Assert().Contains(actAttrs, exp) {
						missingAttrs = append(missingAttrs, exp)
					}
				}
				if len(missingAttrs) > 0 {
					s.T().Logf("Events:\n%s", strings.Join(assertions.ABCIEventsToStrings(txResp.Events), "\n"))
					s.T().Logf("Missing Expected Attributes:\n%s", strings.Join(assertions.AttrsToStrings(missingAttrs), "\n"))
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdAddNetAssetValues() {
	denom := "updatenavcoin"
	argsWStdFlags := func(args ...string) []string {
		return append(args,
			fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
		)
	}

	s.Run("add a new marker for this", func() {
		cmd := markercli.GetCmdAddFinalizeActivateMarker()
		args := argsWStdFlags(
			"1000"+denom,
			s.testnet.Validators[0].Address.String()+",mint,burn,deposit,withdraw,delete,admin,transfer",
			fmt.Sprintf("--%s=%s", markercli.FlagType, "RESTRICTED"),
			"--"+markercli.FlagSupplyFixed,
			"--"+markercli.FlagAllowGovernanceControl,
		)
		testcli.NewTxExecutor(cmd, args).Execute(s.T(), s.testnet)
	})
	if s.T().Failed() {
		s.FailNow("Stopping due to setup error")
	}

	tests := []struct {
		name   string
		args   []string
		expErr string
	}{
		{
			name:   "invalid net asset string",
			args:   argsWStdFlags(denom, "invalid"),
			expErr: "invalid net asset value, expected coin,volume",
		},
		{
			name:   "validate basic fail",
			args:   argsWStdFlags("x", "1usd,1"),
			expErr: "invalid denom: x",
		},
		{
			name: "successful",
			args: argsWStdFlags(denom, "1usd,1"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			testcli.NewTxExecutor(markercli.GetCmdAddNetAssetValues(), tc.args).
				WithExpErrMsg(tc.expErr).
				Execute(s.T(), s.testnet)
		})
	}
}

func (s *IntegrationTestSuite) TestParseAccessGrantFromString() {
	testCases := []struct {
		name              string
		accessGrantString string
		expPanic          bool
		expResult         []types.AccessGrant
	}{
		{
			name:              "successfully parses empty string",
			accessGrantString: "",
			expPanic:          false,
			expResult:         []types.AccessGrant{},
		},
		{
			name:              "fails parsing invalid string",
			accessGrantString: "blah",
			expPanic:          true,
			expResult:         []types.AccessGrant{},
		},
		{
			name:              "should fail empty list of permissions",
			accessGrantString: ",,;",
			expPanic:          true,
			expResult:         []types.AccessGrant{},
		},
		{
			name:              "should fail address is not valid",
			accessGrantString: "NotAnAddress,mint;",
			expPanic:          true,
			expResult:         []types.AccessGrant{},
		},
		{
			name:              "should succeed to add access type",
			accessGrantString: fmt.Sprintf("%s,mint;", s.accountAddresses[0].String()),
			expPanic:          false,
			expResult:         []types.AccessGrant{{Address: s.accountAddresses[0].String(), Permissions: []markertypes.Access{markertypes.Access_Mint}}},
		},
		{
			name:              "should succeed to add access type",
			accessGrantString: fmt.Sprintf("%s,mint;", s.accountAddresses[0].String()),
			expPanic:          false,
			expResult:         []types.AccessGrant{{Address: s.accountAddresses[0].String(), Permissions: []markertypes.Access{markertypes.Access_Mint}}},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var actual []types.AccessGrant
			testFunc := func() {
				actual = markercli.ParseAccessGrantFromString(tc.accessGrantString)
			}
			if tc.expPanic {
				s.Require().Panics(testFunc, "ParseAccessGrantFromString")
			} else {
				s.Require().NotPanics(testFunc, "ParseAccessGrantFromString")
				s.Assert().ElementsMatch(actual, tc.expResult)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestParseNetAssertValueString() {
	testCases := []struct {
		name           string
		netAssetValues string
		expErr         string
		expResult      []types.NetAssetValue
	}{
		{
			name:           "successfully parses empty string",
			netAssetValues: "",
			expErr:         "",
			expResult:      []types.NetAssetValue{},
		},
		{
			name:           "invalid coin",
			netAssetValues: "notacoin,1",
			expErr:         "invalid net asset value coin : notacoin",
			expResult:      []types.NetAssetValue{},
		},
		{
			name:           "invalid volume string",
			netAssetValues: "1hotdog,invalidvolume",
			expErr:         "invalid volume invalidvolume",
			expResult:      []types.NetAssetValue{},
		},
		{
			name:           "invalid amount of args",
			netAssetValues: "1hotdog,invalidvolume,notsupposedtobehere",
			expErr:         "invalid net asset value, expected coin,volume",
			expResult:      []types.NetAssetValue{},
		},
		{
			name:           "successfully parse single nav",
			netAssetValues: "1hotdog,10",
			expErr:         "",
			expResult:      []types.NetAssetValue{{Price: sdk.NewInt64Coin("hotdog", 1), Volume: 10}},
		},
		{
			name:           "successfully parse multi nav",
			netAssetValues: "1hotdog,10;20jackthecat,40",
			expErr:         "",
			expResult:      []types.NetAssetValue{{Price: sdk.NewInt64Coin("hotdog", 1), Volume: 10}, {Price: sdk.NewInt64Coin("jackthecat", 20), Volume: 40}},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result, err := markercli.ParseNetAssetValueString(tc.netAssetValues)
			if len(tc.expErr) > 0 {
				s.Assert().Equal(tc.expErr, err.Error())
				s.Assert().Empty(result)
			} else {
				s.Assert().NoError(err)
				s.Assert().ElementsMatch(result, tc.expResult)
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
	argUsdMills := "--" + markercli.FlagUsdMills
	argVolume := "--" + markercli.FlagVolume

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
			name:   "usd mills present without volume",
			cmd:    getTestCmd(),
			args:   []string{argUsdMills + "=10"},
			expErr: []string{"incorrect value for volume flag.  Must be positive number if usd-mills flag has been set to positive value"},
		},
		{
			name: "volume present",
			cmd:  getTestCmd(),
			args: []string{argVolume + "=11"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        false,
				AllowGovControl:    false,
				AllowForceTransfer: false,
				RequiredAttributes: []string{},
				UsdMills:           0,
				Volume:             11,
			},
		},
		{
			name: "usd-mills and volume present",
			cmd:  getTestCmd(),
			args: []string{argVolume + "=11", argUsdMills + "=1"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_Coin,
				SupplyFixed:        false,
				AllowGovControl:    false,
				AllowForceTransfer: false,
				RequiredAttributes: []string{},
				UsdMills:           1,
				Volume:             11,
			},
		},
		{
			name: "everything",
			cmd:  getTestCmd(),
			args: []string{argForce, argGov, argType, "RESTRICTED", argFixed, argRequiredAtt, "jack.the.cat.io,george.the.dog.io", argUsdMills, "10", argVolume, "12"},
			exp: &markercli.NewMarkerFlagValues{
				MarkerType:         types.MarkerType_RestrictedCoin,
				SupplyFixed:        true,
				AllowGovControl:    true,
				AllowForceTransfer: true,
				RequiredAttributes: []string{"jack.the.cat.io", "george.the.dog.io"},
				UsdMills:           10,
				Volume:             12,
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

func TestParseBoolStrict(t *testing.T) {
	trueCases := []string{
		"true", "TRUE", "True", "tRuE",
	}
	falseCases := []string{
		"false", "FALSE", "False", "FalSe",
	}
	errCases := []string{
		"yes", "no", "y", "n", "0", "1",
		"t rue", "fa lse", "truetrue", "true false", "false true", "tru", "fals", "T", "F",
	}

	type testCase struct {
		input  string
		exp    bool
		expErr bool
	}

	tests := []testCase(nil)

	for _, tc := range trueCases {
		tests = append(tests, []testCase{
			{input: tc, exp: true},
			{input: " " + tc, exp: true},
			{input: tc + " ", exp: true},
			{input: "   " + tc + "   ", exp: true},
		}...)
	}

	for _, tc := range falseCases {
		tests = append(tests, []testCase{
			{input: tc, exp: false},
			{input: " " + tc, exp: false},
			{input: tc + " ", exp: false},
			{input: "   " + tc + "   ", exp: false},
		}...)
	}

	for _, tc := range errCases {
		tests = append(tests, []testCase{
			{input: tc, expErr: true},
			{input: " " + tc, expErr: true},
			{input: tc + " ", expErr: true},
			{input: "   " + tc + "   ", expErr: true},
		}...)
	}

	tests = append(tests, testCase{input: "", expErr: true})
	tests = append(tests, testCase{input: " ", expErr: true})
	tests = append(tests, testCase{input: "   ", expErr: true})

	for _, tc := range tests {
		name := tc.input
		if len(name) == 0 {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			actual, err := markercli.ParseBoolStrict(tc.input)
			if tc.expErr {
				exp := fmt.Sprintf("invalid boolean string: %q", tc.input)
				assert.EqualError(t, err, exp, "ParseBoolStrict(%q) error", tc.input)
			} else {
				assert.NoError(t, err, "ParseBoolStrict(%q) error", tc.input)
			}
			assert.Equal(t, tc.exp, actual, "ParseBoolStrict(%q) value", tc.input)
		})
	}
}

func (s *IntegrationTestSuite) TestSupplyDecreaseProposal() {
	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name: "success - submit supply decrease proposal",
			args: []string{
				"1000stake",
			},
			expectedCode: 0,
			signer:       s.testnet.Validators[0].Address.String(),
		},
		{
			name: "failure - invalid amount",
			args: []string{
				"invalidamountstake",
			},
			expectErrMsg: "invalid coin invalidamountstake",
			signer:       s.testnet.Validators[0].Address.String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := markercli.GetCmdSupplyDecreaseProposal()
			tc.args = append(tc.args,
				"--title", fmt.Sprintf("title: %v", tc.name),
				"--summary", fmt.Sprintf("summary: %v", tc.name),
				"--deposit=1000000stake",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			)

			testcli.NewTxExecutor(cmd, tc.args).
				WithExpErrMsg(tc.expectErrMsg).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}
}

func (s *IntegrationTestSuite) TestSupplyIncreaseProposal() {
	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name: "success - submit supply increase proposal",
			args: []string{
				"1000stake",
			},
			expectedCode: 0,
			signer:       s.testnet.Validators[0].Address.String(),
		},
		{
			name: "success - submit supply increase proposal with target address",
			args: []string{
				"1000stake",
				"--target-address=" + s.accountAddresses[1].String(),
			},
			expectedCode: 0,
			signer:       s.testnet.Validators[0].Address.String(),
		},
		{
			name: "failure - invalid amount",
			args: []string{
				"invalidamountstake",
			},
			expectErrMsg: "invalid coin invalidamountstake",
			signer:       s.testnet.Validators[0].Address.String(),
		},
		{
			name: "failure - invalid target address",
			args: []string{
				"1000stake",
				"--target-address=invalidaddress",
			},
			expectErrMsg: "invalid target address invalidaddress: decoding bech32 failed: invalid separator index -1",
			signer:       s.testnet.Validators[0].Address.String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := markercli.GetCmdSupplyIncreaseProposal()
			tc.args = append(tc.args,
				"--title", fmt.Sprintf("title: %v", tc.name),
				"--summary", fmt.Sprintf("summary: %v", tc.name),
				"--deposit=1000000stake",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			)

			testcli.NewTxExecutor(cmd, tc.args).
				WithExpErrMsg(tc.expectErrMsg).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}
}

func (s *IntegrationTestSuite) TestSetAdministratorProposal() {
	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name: "success - submit set administrator proposal",
			args: []string{
				"mycoin",
				fmt.Sprintf("%s,admin,mint;%s,transfer", s.accountAddresses[0].String(), s.accountAddresses[1].String()),
			},
			expectedCode: 0,
			signer:       s.testnet.Validators[0].Address.String(),
		},
		{
			name: "failure - invalid access grant format",
			args: []string{
				"mycoin",
				"invalidaccessgrant",
			},
			expectErrMsg: "invalid access grants invalidaccessgrant: at least one grant should be provided with address",
			signer:       s.testnet.Validators[0].Address.String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := markercli.GetCmdSetAdministratorProposal()
			tc.args = append(tc.args,
				"--title", fmt.Sprintf("title: %v", tc.name),
				"--summary", fmt.Sprintf("summary: %v", tc.name),
				"--deposit=1000000stake",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			)

			testcli.NewTxExecutor(cmd, tc.args).
				WithExpErrMsg(tc.expectErrMsg).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}
}

func (s *IntegrationTestSuite) TestRemoveAdministratorProposal() {
	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name: "success - submit remove administrator proposal",
			args: []string{
				"mycoin",
				fmt.Sprintf("%s,%s", s.accountAddresses[0].String(), s.accountAddresses[1].String()),
			},
			expectedCode: 0,
			signer:       s.testnet.Validators[0].Address.String(),
		},
		{
			name: "failure - invalid address format",
			args: []string{
				"mycoin",
				"invalidaddress",
			},
			expectErrMsg: "invalid address invalidaddress: decoding bech32 failed: invalid separator index -1",
			signer:       s.testnet.Validators[0].Address.String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := markercli.GetCmdRemoveAdministratorProposal()
			tc.args = append(tc.args,
				"--title", fmt.Sprintf("title: %v", tc.name),
				"--summary", fmt.Sprintf("summary: %v", tc.name),
				"--deposit=1000000stake",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			)

			testcli.NewTxExecutor(cmd, tc.args).
				WithExpErrMsg(tc.expectErrMsg).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}
}

func (s *IntegrationTestSuite) TestChangeStatusProposal() {
	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name: "success - submit change status proposal to active",
			args: []string{
				"mycoin",
				"ACTIVE",
			},
			expectedCode: 0,
			signer:       s.testnet.Validators[0].Address.String(),
		},
		{
			name: "success - submit change status proposal to proposed",
			args: []string{
				"mycoin",
				"proposed",
			},
			expectedCode: 0,
			signer:       s.testnet.Validators[0].Address.String(),
		},
		{
			name: "failure - invalid marker status",
			args: []string{
				"mycoin",
				"INVALIDSTATUS",
			},
			expectErrMsg: "invalid status: INVALIDSTATUS",
			signer:       s.testnet.Validators[0].Address.String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := markercli.GetCmdChangeStatusProposal()
			tc.args = append(tc.args,
				"--title", fmt.Sprintf("title: %v", tc.name),
				"--summary", fmt.Sprintf("summary: %v", tc.name),
				"--deposit=1000000stake",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			)

			testcli.NewTxExecutor(cmd, tc.args).
				WithExpErrMsg(tc.expectErrMsg).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}
}

func (s *IntegrationTestSuite) TestWithdrawEscrowProposal() {
	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name: "success - submit withdraw escrow proposal",
			args: []string{
				"mycoin",
				"100stake",
				s.accountAddresses[1].String(),
			},
			expectedCode: 0,
			signer:       s.testnet.Validators[0].Address.String(),
		},
		{
			name: "failure - invalid amount",
			args: []string{
				"mycoin",
				"invalidamount",
				s.accountAddresses[1].String(),
			},
			expectErrMsg: "invalid amount invalidamount: invalid decimal coin expression: invalidamount",
			signer:       s.testnet.Validators[0].Address.String(),
		},
		{
			name: "failure - invalid target address",
			args: []string{
				"mycoin",
				"100stake",
				"invalidaddress",
			},
			expectErrMsg: "invalid target address invalidaddress: decoding bech32 failed: invalid separator index -1",
			signer:       s.testnet.Validators[0].Address.String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := markercli.GetCmdWithdrawEscrowProposal()
			tc.args = append(tc.args,
				"--title", fmt.Sprintf("title: %v", tc.name),
				"--summary", fmt.Sprintf("summary: %v", tc.name),
				"--deposit=1000000stake",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			)

			testcli.NewTxExecutor(cmd, tc.args).
				WithExpErrMsg(tc.expectErrMsg).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}
}

func (s *IntegrationTestSuite) TestSetDenomMetadataProposal() {
	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name: "success - submit set denom metadata proposal",
			args: []string{
				"mycoin", "My Coin", "MYC", "My coin description", "myc", "6",
			},
			expectedCode: 0,
			signer:       s.testnet.Validators[0].Address.String(),
		},
		{
			name: "failure - invalid exponent",
			args: []string{
				"mycoin", "My Coin", "MYC", "My coin description", "myc", "invalidexponent",
			},
			expectErrMsg: "invalid exponent: invalidexponent",
			signer:       s.testnet.Validators[0].Address.String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := markercli.GetCmdSetDenomMetadataProposal()
			tc.args = append(tc.args,
				"--title", fmt.Sprintf("title: %v", tc.name),
				"--summary", fmt.Sprintf("summary: %v", tc.name),
				"--deposit=1000000stake",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			)

			testcli.NewTxExecutor(cmd, tc.args).
				WithExpErrMsg(tc.expectErrMsg).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}
}
