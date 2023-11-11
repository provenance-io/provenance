package cli_test

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/client/cli"
)

const (
	flagInt         = "int"
	flagString      = "string"
	flagStringSlice = "string-slice"
	flagUintSlice   = "uint-slice"
)

func TestAddFlagsAdmin(t *testing.T) {
	expAnnotations := map[string][]string{
		mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
		oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
	}

	cmd := &cobra.Command{
		Use: "testing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("the command should not have been run")
		},
	}

	cmd.Flags().String(flags.FlagFrom, "", "The from flag")
	cli.AddFlagsAdmin(cmd)

	adminFlag := cmd.Flags().Lookup(cli.FlagAdmin)
	if assert.NotNil(t, adminFlag, "The --%s flag", cli.FlagAdmin) {
		expUsage := "The admin (defaults to --from account)"
		actUsage := adminFlag.Usage
		assert.Equal(t, expUsage, actUsage, "The --%s flag usage", cli.FlagAdmin)
		actAnnotations := adminFlag.Annotations
		assert.Equal(t, expAnnotations, actAnnotations, "The --%s flag annotations", cli.FlagAdmin)
	}

	authorityFlag := cmd.Flags().Lookup(cli.FlagAuthority)
	if assert.NotNil(t, authorityFlag, "The --%s flag", cli.FlagAuthority) {
		expUsage := "Use the governance module account for the admin"
		actUsage := authorityFlag.Usage
		assert.Equal(t, expUsage, actUsage, "The --%s flag usage", cli.FlagAuthority)
		actAnnotations := authorityFlag.Annotations
		assert.Equal(t, expAnnotations, actAnnotations, "The --%s flag annotations", cli.FlagAuthority)
	}

	flagFrom := cmd.Flags().Lookup(flags.FlagFrom)
	if assert.NotNil(t, flagFrom, "The --%s flag", flags.FlagFrom) {
		fromExpAnnotations := map[string][]string{oneReq: expAnnotations[oneReq]}
		actAnnotations := flagFrom.Annotations
		assert.Equal(t, fromExpAnnotations, actAnnotations, "The --%s flag annotations", flags.FlagFrom)
	}
}

func TestReadFlagsAdminOrFrom(t *testing.T) {
	goodFlagSet := func() *pflag.FlagSet {
		flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
		flagSet.String(cli.FlagAdmin, "", "The admin")
		flagSet.Bool(cli.FlagAuthority, false, "Use authority")
		return flagSet
	}

	tests := []struct {
		name      string
		flagSet   func() *pflag.FlagSet
		flags     []string
		clientCtx client.Context
		expAddr   string
		expErr    string
	}{
		{
			name: "wrong admin flag type",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.Int(cli.FlagAdmin, 0, "The admin")
				flagSet.Bool(cli.FlagAuthority, false, "Use authority")
				return flagSet
			},
			expErr: "trying to get string value of flag of type int",
		},
		{
			name: "wrong authority flag type",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagAdmin, "", "The admin")
				flagSet.Int(cli.FlagAuthority, 0, "Use authority")
				return flagSet

			},
			expErr: "trying to get bool value of flag of type int",
		},
		{
			name:    "admin flag given",
			flags:   []string{"--" + cli.FlagAdmin, "theadmin"},
			expAddr: "theadmin",
		},
		{
			name:    "authority flag given",
			flags:   []string{"--" + cli.FlagAuthority},
			expAddr: cli.AuthorityAddr.String(),
		},
		{
			name:      "from address given",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			expAddr:   sdk.AccAddress("FromAddress_________").String(),
		},
		{
			name:   "nothing given",
			expErr: "no admin provided",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.flagSet == nil {
				tc.flagSet = goodFlagSet
			}
			flagSet := tc.flagSet()
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var addr string
			testFunc := func() {
				addr, err = cli.ReadFlagsAdminOrFrom(tc.clientCtx, flagSet)
			}
			require.NotPanics(t, testFunc, "ReadFlagsAdminOrFrom")
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadFlagsAdminOrFrom error")
			assert.Equal(t, tc.expAddr, addr, "ReadFlagsAdminOrFrom address")
		})
	}
}

func TestReadFlagAuthority(t *testing.T) {
	goodFlagSet := func() *pflag.FlagSet {
		flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
		flagSet.String(cli.FlagAuthority, "", "The authority")
		return flagSet
	}

	tests := []struct {
		name    string
		flagSet func() *pflag.FlagSet
		flags   []string
		expAddr string
		expErr  string
	}{
		{
			name: "wrong flag type",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.Int(cli.FlagAuthority, 0, "The authority")
				return flagSet

			},
			expErr: "trying to get string value of flag of type int",
		},
		{
			name:    "provided",
			flagSet: goodFlagSet,
			flags:   []string{"--" + cli.FlagAuthority, "usemeinstead"},
			expAddr: "usemeinstead",
		},
		{
			name:    "not provided",
			flagSet: goodFlagSet,
			expAddr: cli.AuthorityAddr.String(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flagSet := tc.flagSet()
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var addr string
			testFunc := func() {
				addr, err = cli.ReadFlagAuthority(flagSet)
			}
			require.NotPanics(t, testFunc, "ReadFlagAuthority")
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadFlagAuthority error")
			assert.Equal(t, tc.expAddr, addr, "ReadFlagAuthority address")
		})
	}
}

func TestReadAddrFlagOrFrom(t *testing.T) {
	tests := []struct {
		testName  string
		flags     []string
		clientCtx client.Context
		name      string
		expAddr   string
		expErr    string
	}{
		{
			testName: "unknown flag",
			name:     "notsetup",
			expErr:   "flag accessed but not defined: notsetup",
		},
		{
			testName: "wrong flag type",
			name:     flagInt,
			expErr:   "trying to get string value of flag of type int",
		},
		{
			testName: "flag given",
			flags:    []string{"--" + flagString, "someaddr"},
			name:     flagString,
			expAddr:  "someaddr",
		},
		{
			testName:  "using from",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			name:      flagString,
			expAddr:   sdk.AccAddress("FromAddress_________").String(),
		},
		{
			testName: "not provided",
			name:     flagString,
			expErr:   "no string provided",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
			flagSet.String(flagString, "", "A string")
			flagSet.Int(flagInt, 0, "An int")
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var addr string
			testFunc := func() {
				addr, err = cli.ReadAddrFlagOrFrom(tc.clientCtx, flagSet, tc.name)
			}
			require.NotPanics(t, testFunc, "ReadAddrFlagOrFrom")
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadAddrFlagOrFrom error")
			assert.Equal(t, tc.expAddr, addr, "ReadAddrFlagOrFrom address")
		})
	}
}

func TestAddFlagsEnableDisable(t *testing.T) {
	expAnnotations := map[string][]string{
		mutExc: {cli.FlagEnable + " " + cli.FlagDisable},
		oneReq: {cli.FlagEnable + " " + cli.FlagDisable},
	}

	cmd := &cobra.Command{
		Use: "testing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("the command should not have been run")
		},
	}

	cli.AddFlagsEnableDisable(cmd, "unittest")

	enableFlag := cmd.Flags().Lookup(cli.FlagEnable)
	if assert.NotNil(t, enableFlag, "The --%s flag", cli.FlagEnable) {
		expUsage := "Set the market's unittest field to true"
		actusage := enableFlag.Usage
		assert.Equal(t, expUsage, actusage, "--%s flag usage", cli.FlagEnable)
		actAnnotations := enableFlag.Annotations
		assert.Equal(t, expAnnotations, actAnnotations, "--%s flag annotations", cli.FlagEnable)
	}

	disableFlag := cmd.Flags().Lookup(cli.FlagDisable)
	if assert.NotNil(t, disableFlag, "The --%s flag", cli.FlagDisable) {
		expUsage := "Set the market's unittest field to false"
		actusage := disableFlag.Usage
		assert.Equal(t, expUsage, actusage, "--%s flag usage", cli.FlagDisable)
		actAnnotations := disableFlag.Annotations
		assert.Equal(t, expAnnotations, actAnnotations, "--%s flag annotations", cli.FlagDisable)
	}
}

func TestReadFlagsEnableDisable(t *testing.T) {
	goodFlagSet := func() *pflag.FlagSet {
		flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
		flagSet.Bool(cli.FlagEnable, false, "Enable")
		flagSet.Bool(cli.FlagDisable, false, "Disable")
		return flagSet
	}

	tests := []struct {
		name    string
		flags   []string
		flagSet func() *pflag.FlagSet
		exp     bool
		expErr  string
	}{
		{
			name: "cannot read enable",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.Int(cli.FlagEnable, 0, "Enable")
				flagSet.Bool(cli.FlagDisable, false, "Disable")
				return flagSet
			},
			expErr: "trying to get bool value of flag of type int",
		},
		{
			name: "cannot read disable",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.Bool(cli.FlagEnable, false, "Enable")
				flagSet.Int(cli.FlagDisable, 0, "Disable")
				return flagSet
			},
			expErr: "trying to get bool value of flag of type int",
		},
		{
			name:    "enable",
			flags:   []string{"--" + cli.FlagEnable},
			flagSet: goodFlagSet,
			exp:     true,
		},
		{
			name:    "disable",
			flags:   []string{"--" + cli.FlagDisable},
			flagSet: goodFlagSet,
			exp:     false,
		},
		{
			name:    "neither",
			flagSet: goodFlagSet,
			expErr:  "exactly one of --enable or --disable must be provided",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flagSet := tc.flagSet()
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var act bool
			testFunc := func() {
				act, err = cli.ReadFlagsEnableDisable(flagSet)
			}
			require.NotPanics(t, testFunc, "ReadFlagsEnableDisable")
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadFlagsEnableDisable error")
			assert.Equal(t, tc.exp, act, "ReadFlagsEnableDisable bool")
		})
	}
}

func TestReadFlagsAsksBidsOpt(t *testing.T) {
	goodFlagSet := func() *pflag.FlagSet {
		flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
		flagSet.Bool(cli.FlagAsks, false, "Asks")
		flagSet.Bool(cli.FlagBids, false, "Bids")
		return flagSet
	}

	tests := []struct {
		name    string
		flags   []string
		flagSet func() *pflag.FlagSet
		expStr  string
		expErr  string
	}{
		{
			name: "cannot read asks",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.Int(cli.FlagAsks, 0, "Asks")
				flagSet.Bool(cli.FlagBids, false, "Bids")
				return flagSet
			},
			expErr: "trying to get bool value of flag of type int",
		},
		{
			name: "cannot read bids",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.Bool(cli.FlagAsks, false, "Asks")
				flagSet.Int(cli.FlagBids, 0, "Bids")
				return flagSet
			},
			expErr: "trying to get bool value of flag of type int",
		},
		{
			name:    "asks",
			flags:   []string{"--" + cli.FlagAsks},
			flagSet: goodFlagSet,
			expStr:  "ask",
		},
		{
			name:    "bids",
			flags:   []string{"--" + cli.FlagBids},
			flagSet: goodFlagSet,
			expStr:  "bid",
		},
		{
			name:    "neither",
			flagSet: goodFlagSet,
			expStr:  "",
			expErr:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flagSet := tc.flagSet()
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var str string
			testFunc := func() {
				str, err = cli.ReadFlagsAsksBidsOpt(flagSet)
			}
			require.NotPanics(t, testFunc, "ReadFlagsAsksBidsOpt")
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadFlagsAsksBidsOpt error")
			assert.Equal(t, tc.expStr, str, "ReadFlagsAsksBidsOpt string")
		})
	}
}

func TestReadFlagOrderOrArg(t *testing.T) {
	theFlag := cli.FlagOrder
	goodFlagSet := func() *pflag.FlagSet {
		flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
		flagSet.Uint64(theFlag, 0, "The id")
		return flagSet
	}
	badFlagSet := func() *pflag.FlagSet {
		flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
		flagSet.String(theFlag, "", "The id")
		return flagSet
	}

	tests := []struct {
		name    string
		flags   []string
		flagSet *pflag.FlagSet
		args    []string
		expID   uint64
		expErr  string
	}{
		{
			name:    "unknown flag",
			flagSet: pflag.NewFlagSet("", pflag.ContinueOnError),
			expErr:  "flag accessed but not defined: " + theFlag,
		},
		{
			name:    "wrong flag type",
			flagSet: badFlagSet(),
			expErr:  "trying to get uint64 value of flag of type string",
		},
		{
			name:    "both flag and arg",
			flags:   []string{"--" + theFlag, "8"},
			flagSet: goodFlagSet(),
			args:    []string{"8"},
			expErr:  "cannot provide <order id> as both an arg (\"8\") and flag (--order 8)",
		},
		{
			name:    "just flag",
			flags:   []string{"--" + theFlag, "8"},
			flagSet: goodFlagSet(),
			expID:   8,
		},
		{
			name:    "just flag zero",
			flags:   []string{"--" + theFlag, "0"},
			flagSet: goodFlagSet(),
			expErr:  "no <order id> provided",
		},
		{
			name:    "just arg, bad",
			flagSet: goodFlagSet(),
			args:    []string{"8v8"},
			expErr:  "could not convert <order id> arg: strconv.ParseUint: parsing \"8v8\": invalid syntax",
		},
		{
			name:    "just arg, zero",
			flagSet: goodFlagSet(),
			args:    []string{"0"},
			expErr:  "no <order id> provided",
		},
		{
			name:    "just arg, good",
			flagSet: goodFlagSet(),
			args:    []string{"987"},
			expID:   987,
		},
		{
			name:    "neither flag nor arg",
			flagSet: goodFlagSet(),
			expErr:  "no <order id> provided",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var id uint64
			testFunc := func() {
				id, err = cli.ReadFlagOrderOrArg(tc.flagSet, tc.args)
			}
			require.NotPanics(t, testFunc, "ReadFlagOrderOrArg")
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadFlagOrderOrArg error")
			assert.Equal(t, int(tc.expID), int(id), "ReadFlagOrderOrArg id")
		})
	}
}

func TestReadFlagMarketOrArg(t *testing.T) {
	theFlag := cli.FlagMarket
	goodFlagSet := func() *pflag.FlagSet {
		flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
		flagSet.Uint32(theFlag, 0, "The id")
		return flagSet
	}
	badFlagSet := func() *pflag.FlagSet {
		flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
		flagSet.String(theFlag, "", "The id")
		return flagSet
	}

	tests := []struct {
		name    string
		flags   []string
		flagSet *pflag.FlagSet
		args    []string
		expID   uint32
		expErr  string
	}{
		{
			name:    "unknown flag",
			flagSet: pflag.NewFlagSet("", pflag.ContinueOnError),
			expErr:  "flag accessed but not defined: " + theFlag,
		},
		{
			name:    "wrong flag type",
			flagSet: badFlagSet(),
			expErr:  "trying to get uint32 value of flag of type string",
		},
		{
			name:    "both flag and arg",
			flags:   []string{"--" + theFlag, "8"},
			flagSet: goodFlagSet(),
			args:    []string{"8"},
			expErr:  "cannot provide <market id> as both an arg (\"8\") and flag (--market 8)",
		},
		{
			name:    "just flag",
			flags:   []string{"--" + theFlag, "8"},
			flagSet: goodFlagSet(),
			expID:   8,
		},
		{
			name:    "just flag zero",
			flags:   []string{"--" + theFlag, "0"},
			flagSet: goodFlagSet(),
			expErr:  "no <market id> provided",
		},
		{
			name:    "just arg, bad",
			flagSet: goodFlagSet(),
			args:    []string{"8v8"},
			expErr:  "could not convert <market id> arg: strconv.ParseUint: parsing \"8v8\": invalid syntax",
		},
		{
			name:    "just arg, zero",
			flagSet: goodFlagSet(),
			args:    []string{"0"},
			expErr:  "no <market id> provided",
		},
		{
			name:    "just arg, good",
			flagSet: goodFlagSet(),
			args:    []string{"987"},
			expID:   987,
		},
		{
			name:    "neither flag nor arg",
			flagSet: goodFlagSet(),
			expErr:  "no <market id> provided",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var id uint32
			testFunc := func() {
				id, err = cli.ReadFlagMarketOrArg(tc.flagSet, tc.args)
			}
			require.NotPanics(t, testFunc, "ReadFlagMarketOrArg")
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadFlagMarketOrArg error")
			assert.Equal(t, int(tc.expID), int(id), "ReadFlagMarketOrArg id")
		})
	}
}

func TestReadCoinsFlag(t *testing.T) {
	tests := []struct {
		testName string
		flags    []string
		name     string
		expCoins sdk.Coins
		expErr   string
	}{
		{
			testName: "unknown flag",
			name:     "unknown",
			expErr:   "flag accessed but not defined: unknown",
		},
		{
			testName: "wrong flag type",
			name:     flagInt,
			expErr:   "trying to get string value of flag of type int",
		},
		{
			testName: "nothing provided",
			name:     flagString,
			expErr:   "",
		},
		{
			testName: "invalid coins",
			flags:    []string{"--" + flagString, "2yupcoin,nopecoin"},
			name:     flagString,
			expErr:   "error parsing --" + flagString + " as coins: invalid coin expression: \"nopecoin\"",
		},
		{
			testName: "one coin",
			flags:    []string{"--" + flagString, "2grape"},
			name:     flagString,
			expCoins: sdk.NewCoins(sdk.NewInt64Coin("grape", 2)),
		},
		{
			testName: "three coins",
			flags:    []string{"--" + flagString, "8banana,5apple,14cherry"},
			name:     flagString,
			expCoins: sdk.NewCoins(
				sdk.NewInt64Coin("apple", 5), sdk.NewInt64Coin("banana", 8), sdk.NewInt64Coin("cherry", 14),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
			flagSet.String(flagString, "", "A string")
			flagSet.Int(flagInt, 0, "An int")
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var coins sdk.Coins
			testFunc := func() {
				coins, err = cli.ReadCoinsFlag(flagSet, tc.name)
			}
			require.NotPanics(t, testFunc, "ReadCoinsFlag(%q)", tc.name)
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadCoinsFlag(%q) error", tc.name)
			assert.Equal(t, tc.expCoins.String(), coins.String(), "ReadCoinsFlag(%q) coins", tc.name)
		})
	}
}

func TestParseCoins(t *testing.T) {
	tests := []struct {
		name     string
		coinsStr string
		expCoins sdk.Coins
		expErr   string
	}{
		{
			name:     "empty string",
			coinsStr: "",
			expCoins: nil,
			expErr:   "",
		},
		{
			name:     "one entry, bad",
			coinsStr: "bad",
			expErr:   "invalid coin expression: \"bad\"",
		},
		{
			name:     "one entry, good",
			coinsStr: "55good",
			expCoins: sdk.NewCoins(sdk.NewInt64Coin("good", 55)),
		},
		{
			name:     "three entries, first bad",
			coinsStr: "1234,555second,63third",
			expErr:   "invalid coin expression: \"1234\"",
		},
		{
			name:     "three entries, second bad",
			coinsStr: "1234first,second,55third",
			expErr:   "invalid coin expression: \"second\"",
		},
		{
			name:     "three entries, third bad",
			coinsStr: "1234first,555second,63x",
			expErr:   "invalid coin expression: \"63x\"",
		},
		{
			name:     "three entries, all good",
			coinsStr: "1234one,555two,63three",
			expCoins: sdk.NewCoins(
				sdk.NewInt64Coin("one", 1234),
				sdk.NewInt64Coin("three", 63),
				sdk.NewInt64Coin("two", 555),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var coins sdk.Coins
			var err error
			testFunc := func() {
				coins, err = cli.ParseCoins(tc.coinsStr)
			}
			require.NotPanics(t, testFunc, "ParseCoins(%q)", tc.coinsStr)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseCoins(%q) error", tc.coinsStr)
			assert.Equal(t, tc.expCoins.String(), coins.String(), "ParseCoins(%q) coins", tc.coinsStr)
		})
	}
}

func TestReadCoinFlag(t *testing.T) {
	tests := []struct {
		testName string
		flags    []string
		name     string
		expCoin  *sdk.Coin
		expErr   string
	}{
		{
			testName: "unknown flag",
			name:     "unknown",
			expErr:   "flag accessed but not defined: unknown",
		},
		{
			testName: "wrong flag type",
			name:     flagInt,
			expErr:   "trying to get string value of flag of type int",
		},
		{
			testName: "nothing provided",
			name:     flagString,
			expErr:   "",
		},
		{
			testName: "invalid coin",
			flags:    []string{"--" + flagString, "nopecoin"},
			name:     flagString,
			expErr:   "error parsing --" + flagString + " as a coin: invalid coin expression: \"nopecoin\"",
		},
		{
			testName: "zero coin",
			flags:    []string{"--" + flagString, "0zerocoin"},
			name:     flagString,
			expCoin:  &sdk.Coin{Denom: "zerocoin", Amount: sdkmath.NewInt(0)},
		},
		{
			testName: "normal coin",
			flags:    []string{"--" + flagString, "99banana"},
			name:     flagString,
			expCoin:  &sdk.Coin{Denom: "banana", Amount: sdkmath.NewInt(99)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
			flagSet.String(flagString, "", "A string")
			flagSet.Int(flagInt, 0, "An int")
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var coin *sdk.Coin
			testFunc := func() {
				coin, err = cli.ReadCoinFlag(flagSet, tc.name)
			}
			require.NotPanics(t, testFunc, "ReadCoinFlag(%q)", tc.name)
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadCoinFlag(%q) error", tc.name)
			if !assert.Equal(t, tc.expCoin, coin, "ReadCoinFlag(%q)", tc.name) && tc.expCoin != nil && coin != nil {
				t.Logf("Expected: %q", tc.expCoin)
				t.Logf("  Actual: %q", coin)
			}
		})
	}
}

func TestReadReqCoinFlag(t *testing.T) {
	tests := []struct {
		testName string
		flags    []string
		name     string
		expCoin  sdk.Coin
		expErr   string
	}{
		{
			testName: "unknown flag",
			name:     "unknown",
			expErr:   "flag accessed but not defined: unknown",
		},
		{
			testName: "wrong flag type",
			name:     flagInt,
			expErr:   "trying to get string value of flag of type int",
		},
		{
			testName: "nothing provided",
			name:     flagString,
			expErr:   "missing required --" + flagString + " flag",
		},
		{
			testName: "invalid coin",
			flags:    []string{"--" + flagString, "nopecoin"},
			name:     flagString,
			expErr:   "error parsing --" + flagString + " as a coin: invalid coin expression: \"nopecoin\"",
		},
		{
			testName: "zero coin",
			flags:    []string{"--" + flagString, "0zerocoin"},
			name:     flagString,
			expCoin:  sdk.Coin{Denom: "zerocoin", Amount: sdkmath.NewInt(0)},
		},
		{
			testName: "normal coin",
			flags:    []string{"--" + flagString, "99banana"},
			name:     flagString,
			expCoin:  sdk.Coin{Denom: "banana", Amount: sdkmath.NewInt(99)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
			flagSet.String(flagString, "", "A string")
			flagSet.Int(flagInt, 0, "An int")
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var coin sdk.Coin
			testFunc := func() {
				coin, err = cli.ReadReqCoinFlag(flagSet, tc.name)
			}
			require.NotPanics(t, testFunc, "ReadReqCoinFlag(%q)", tc.name)
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadReqCoinFlag(%q) error", tc.name)
			assert.Equal(t, tc.expCoin.String(), coin.String(), "ReadReqCoinFlag(%q)", tc.name)
		})
	}
}

func TestReadOrderIDsFlag(t *testing.T) {
	tests := []struct {
		testName string
		flags    []string
		name     string
		expIDs   []uint64
		expErr   string
	}{
		{
			testName: "unknown flag",
			name:     "unknown",
			expErr:   "flag accessed but not defined: unknown",
		},
		{
			testName: "wrong flag type",
			name:     flagString,
			expErr:   "trying to get string value of flag of type uintSlice",
		},
		{
			testName: "nothing provided",
			name:     flagUintSlice,
			expErr:   "",
		},
		{
			testName: "one val",
			flags:    []string{"--" + flagUintSlice, "15"},
			name:     flagUintSlice,
			expIDs:   []uint64{15},
		},
		{
			testName: "three vals",
			flags:    []string{"--" + flagUintSlice, "42,9001,3"},
			name:     flagUintSlice,
			expIDs:   []uint64{42, 9001, 3},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
			flagSet.UintSlice(flagUintSlice, nil, "A slice of uints")
			flagSet.String(flagString, "", "A string")
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var ids []uint64
			testFunc := func() {
				ids, err = cli.ReadOrderIDsFlag(flagSet, tc.name)
			}
			require.NotPanics(t, testFunc, "ReadOrderIDsFlag(%q)", tc.name)
			assertEqualSlices(t, tc.expIDs, ids, orderIDStringer, "ReadOrderIDsFlag(%q)", tc.name)
		})
	}
}

func TestReadAccessGrantsFlag(t *testing.T) {
	tests := []struct {
		testName  string
		flags     []string
		name      string
		expGrants []exchange.AccessGrant
		expErr    string
	}{
		{
			testName: "unknown flag",
			name:     "unknown",
			expErr:   "flag accessed but not defined: unknown",
		},
		{
			testName: "wrong flag type",
			name:     flagInt,
			expErr:   "trying to get stringSlice value of flag of type int",
		},
		{
			testName: "nothing provided",
			name:     flagStringSlice,
			expErr:   "",
		},
		{
			testName: "three vals, one bad",
			flags: []string{
				"--" + flagStringSlice, "addr1:all",
				"--" + flagStringSlice, "withdraw",
				"--" + flagStringSlice, "addr2:setids+update",
			},
			name: flagStringSlice,
			expGrants: []exchange.AccessGrant{
				{
					Address:     "addr1",
					Permissions: exchange.AllPermissions(),
				},
				{
					Address:     "addr2",
					Permissions: []exchange.Permission{exchange.Permission_set_ids, exchange.Permission_update},
				},
			},
			expErr: "could not parse \"withdraw\" as an <access grant>: expected format <address>:<permissions>",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
			flagSet.StringSlice(flagStringSlice, nil, "A string slice")
			flagSet.Int(flagInt, 0, "An int")
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var grants []exchange.AccessGrant
			testFunc := func() {
				grants, err = cli.ReadAccessGrantsFlag(flagSet, tc.name)
			}
			require.NotPanics(t, testFunc, "ReadAccessGrantsFlag(%q)", tc.name)
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadAccessGrantsFlag(%q) error", tc.name)
			assert.Equal(t, tc.expGrants, grants, "ReadAccessGrantsFlag(%q) grants", tc.name)
		})
	}
}

func TestParseAccessGrant(t *testing.T) {
	addr := "pb1v9jxgujlta047h6lta047h6lta047h6l5rpeqp" // = sdk.AccAddress("addr________________")

	tests := []struct {
		name   string
		val    string
		expAG  *exchange.AccessGrant
		expErr string
	}{
		{
			name:   "empty string",
			val:    "",
			expAG:  nil,
			expErr: "could not parse \"\" as an <access grant>: expected format <address>:<permissions>",
		},
		{
			name:   "zero colons",
			val:    "something",
			expErr: "could not parse \"something\" as an <access grant>: expected format <address>:<permissions>",
		},
		{
			name:   "two colons",
			val:    "part0:part1:part2",
			expErr: "could not parse \"part0:part1:part2\" as an <access grant>: expected format <address>:<permissions>",
		},
		{
			name:   "empty address",
			val:    ":part1",
			expErr: "invalid <access grant> \":part1\": both an <address> and <permissions> are required",
		},
		{
			name:   "empty permissions",
			val:    "part0:",
			expErr: "invalid <access grant> \"part0:\": both an <address> and <permissions> are required",
		},
		{
			name:   "unspecified",
			val:    "part0:unspecified",
			expErr: "could not parse permissions for \"part0\" from \"unspecified\": invalid permission: \"unspecified\"",
		},
		{
			name:  "all",
			val:   addr + ":all",
			expAG: &exchange.AccessGrant{Address: addr, Permissions: exchange.AllPermissions()},
		},
		{
			name: "one perm, enum name",
			val:  addr + ":PERMISSION_UPDATE",
			expAG: &exchange.AccessGrant{
				Address:     addr,
				Permissions: []exchange.Permission{exchange.Permission_update},
			},
		},
		{
			name: "one perm, simple name",
			val:  addr + ":cancel",
			expAG: &exchange.AccessGrant{
				Address:     addr,
				Permissions: []exchange.Permission{exchange.Permission_cancel},
			},
		},
		{
			name: "multiple perms, plus delim",
			val:  addr + ":Cancel+PERMISSION_SETTLE+setids",
			expAG: &exchange.AccessGrant{
				Address: addr,
				Permissions: []exchange.Permission{
					exchange.Permission_cancel,
					exchange.Permission_settle,
					exchange.Permission_set_ids,
				},
			},
		},
		{
			name: "multiple perms, dot delim",
			val:  addr + ":permissions.PERMISSION_ATTRIBUTES.withdraw",
			expAG: &exchange.AccessGrant{
				Address: addr,
				Permissions: []exchange.Permission{
					exchange.Permission_permissions,
					exchange.Permission_attributes,
					exchange.Permission_withdraw,
				},
			},
		},
		{
			name: "multiple perms, space delim",
			val:  addr + ":Set_Ids update settle permissions",
			expAG: &exchange.AccessGrant{
				Address: addr,
				Permissions: []exchange.Permission{
					exchange.Permission_set_ids,
					exchange.Permission_update,
					exchange.Permission_settle,
					exchange.Permission_permissions,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var ag *exchange.AccessGrant
			var err error
			testFunc := func() {
				ag, err = cli.ParseAccessGrant(tc.val)
			}
			require.NotPanics(t, testFunc, "ParseAccessGrant(%q)", tc.val)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseAccessGrant(%q) error", tc.val)
			assert.Equal(t, tc.expAG, ag, "ParseAccessGrant(%q) AccessGrant", tc.val)
		})
	}
}

func TestParseAccessGrants(t *testing.T) {
	addr1 := "pb1v9jxgu33ta047h6lta047h6lta047h6l0r6x5v" // = sdk.AccAddress("addr1_______________")
	addr2 := "pb1v9jxgu33ta047h6lta047h6lta047h6l0r6x5v" // = sdk.AccAddress("_addr2______________")
	addr3 := "pb10elxzerywge47h6lta047h6lta047h6l90x0zx" // = sdk.AccAddress("~~addr3_____________")

	tests := []struct {
		name      string
		vals      []string
		expGrants []exchange.AccessGrant
		expErr    string
	}{
		{
			name:   "nil",
			vals:   nil,
			expErr: "",
		},
		{
			name:   "empty",
			vals:   []string{},
			expErr: "",
		},
		{
			name:   "one, bad",
			vals:   []string{"not good"},
			expErr: "could not parse \"not good\" as an <access grant>: expected format <address>:<permissions>",
		},
		{
			name: "one, good",
			vals: []string{addr1 + ":update+permissions"},
			expGrants: []exchange.AccessGrant{{
				Address:     addr1,
				Permissions: []exchange.Permission{exchange.Permission_update, exchange.Permission_permissions},
			}},
		},
		{
			name: "three, all good",
			vals: []string{addr1 + ":settle", addr2 + ":setids", addr3 + ":permission_withdraw"},
			expGrants: []exchange.AccessGrant{
				{Address: addr1, Permissions: []exchange.Permission{exchange.Permission_settle}},
				{Address: addr2, Permissions: []exchange.Permission{exchange.Permission_set_ids}},
				{Address: addr3, Permissions: []exchange.Permission{exchange.Permission_withdraw}},
			},
		},
		{
			name: "three, first bad",
			vals: []string{":settle", addr2 + ":setids", addr3 + ":permission_withdraw"},
			expGrants: []exchange.AccessGrant{
				{Address: addr2, Permissions: []exchange.Permission{exchange.Permission_set_ids}},
				{Address: addr3, Permissions: []exchange.Permission{exchange.Permission_withdraw}},
			},
			expErr: "invalid <access grant> \":settle\": both an <address> and <permissions> are required",
		},
		{
			name: "three, second bad",
			vals: []string{addr1 + ":settle", addr2 + ":unspecified", addr3 + ":permission_withdraw"},
			expGrants: []exchange.AccessGrant{
				{Address: addr1, Permissions: []exchange.Permission{exchange.Permission_settle}},
				{Address: addr3, Permissions: []exchange.Permission{exchange.Permission_withdraw}},
			},
			expErr: "could not parse permissions for \"" + addr2 + "\" from \"unspecified\": invalid permission: \"unspecified\"",
		},
		{
			name: "three, third bad",
			vals: []string{addr1 + ":settle", addr2 + ":setids", "someaddr:"},
			expGrants: []exchange.AccessGrant{
				{Address: addr1, Permissions: []exchange.Permission{exchange.Permission_settle}},
				{Address: addr2, Permissions: []exchange.Permission{exchange.Permission_set_ids}},
			},
			expErr: "invalid <access grant> \"someaddr:\": both an <address> and <permissions> are required",
		},
		{
			name: "three, all bad",
			vals: []string{":settle", addr2 + ":unspecified", "someaddr:"},
			expErr: joinErrs(
				"invalid <access grant> \":settle\": both an <address> and <permissions> are required",
				"could not parse permissions for \""+addr2+"\" from \"unspecified\": invalid permission: \"unspecified\"",
				"invalid <access grant> \"someaddr:\": both an <address> and <permissions> are required",
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expGrants == nil {
				tc.expGrants = []exchange.AccessGrant{}
			}

			var grants []exchange.AccessGrant
			var err error
			testFunc := func() {
				grants, err = cli.ParseAccessGrants(tc.vals)
			}
			require.NotPanics(t, testFunc, "ParseAccessGrants(%q)", tc.vals)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseAccessGrants(%q) error", tc.vals)
			assert.Equal(t, tc.expGrants, grants, "ParseAccessGrants(%q) grants", tc.vals)
		})
	}
}

func TestReadFlatFeeFlag(t *testing.T) {
	tests := []struct {
		testName string
		flags    []string
		name     string
		expCoins []sdk.Coin
		expErr   string
	}{
		{
			testName: "unknown flag",
			name:     "unknown",
			expErr:   "flag accessed but not defined: unknown",
		},
		{
			testName: "wrong flag type",
			name:     flagInt,
			expErr:   "trying to get stringSlice value of flag of type int",
		},
		{
			testName: "nothing provided",
			name:     flagStringSlice,
			expErr:   "",
		},
		{
			testName: "three vals, one bad",
			flags:    []string{"--" + flagStringSlice, "apple,100pear", "--" + flagStringSlice, "777cherry"},
			name:     flagStringSlice,
			expCoins: []sdk.Coin{sdk.NewInt64Coin("pear", 100), sdk.NewInt64Coin("cherry", 777)},
			expErr:   "invalid coin expression: \"apple\"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
			flagSet.StringSlice(flagStringSlice, nil, "A string slice")
			flagSet.Int(flagInt, 0, "An int")
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var coins []sdk.Coin
			testFunc := func() {
				coins, err = cli.ReadFlatFeeFlag(flagSet, tc.name)
			}
			require.NotPanics(t, testFunc, "ReadFlatFeeFlag(%q)", tc.name)
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadFlatFeeFlag(%q) error", tc.name)
			assertEqualSlices(t, tc.expCoins, coins, sdk.Coin.String, "ReadFlatFeeFlag(%q) ratios", tc.name)
		})
	}
}

func TestParseFlatFeeOptions(t *testing.T) {
	tests := []struct {
		name     string
		vals     []string
		expCoins []sdk.Coin
		expErr   string
	}{
		{
			name:   "nil",
			vals:   nil,
			expErr: "",
		},
		{
			name:   "empty",
			vals:   []string{},
			expErr: "",
		},
		{
			name:   "one, bad",
			vals:   []string{"nope"},
			expErr: "invalid coin expression: \"nope\"",
		},
		{
			name:     "one, good",
			vals:     []string{"18banana"},
			expCoins: []sdk.Coin{sdk.NewInt64Coin("banana", 18)},
		},
		{
			name:     "one, zero",
			vals:     []string{"0durian"},
			expCoins: []sdk.Coin{sdk.NewInt64Coin("durian", 0)},
		},
		{
			name: "three, all good",
			vals: []string{"1apple", "2banana", "3cherry"},
			expCoins: []sdk.Coin{
				sdk.NewInt64Coin("apple", 1), sdk.NewInt64Coin("banana", 2), sdk.NewInt64Coin("cherry", 3),
			},
		},
		{
			name: "three, first bad",
			vals: []string{"notgonnacoin", "2banana", "3cherry"},
			expCoins: []sdk.Coin{
				sdk.NewInt64Coin("banana", 2), sdk.NewInt64Coin("cherry", 3),
			},
			expErr: "invalid coin expression: \"notgonnacoin\"",
		},
		{
			name: "three, second bad",
			vals: []string{"1apple", "12345", "3cherry"},
			expCoins: []sdk.Coin{
				sdk.NewInt64Coin("apple", 1), sdk.NewInt64Coin("cherry", 3),
			},
			expErr: "invalid coin expression: \"12345\"",
		},
		{
			name: "three, third bad",
			vals: []string{"1apple", "2banana", ""},
			expCoins: []sdk.Coin{
				sdk.NewInt64Coin("apple", 1), sdk.NewInt64Coin("banana", 2),
			},
			expErr: "invalid coin expression: \"\"",
		},
		{
			name: "three, all bad",
			vals: []string{"notgonnacoin", "12345", ""},
			expErr: joinErrs(
				"invalid coin expression: \"notgonnacoin\"",
				"invalid coin expression: \"12345\"",
				"invalid coin expression: \"\"",
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expCoins == nil {
				tc.expCoins = []sdk.Coin{}
			}
			var coins []sdk.Coin
			var err error
			testFunc := func() {
				coins, err = cli.ParseFlatFeeOptions(tc.vals)
			}
			require.NotPanics(t, testFunc, "ParseFlatFeeOptions(%q)", tc.vals)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseFlatFeeOptions(%q) error", tc.vals)
			assertEqualSlices(t, tc.expCoins, coins, sdk.Coin.String, "ParseFlatFeeOptions(%q) coins", tc.vals)
		})
	}
}

func TestReadFeeRatiosFlag(t *testing.T) {
	tests := []struct {
		testName  string
		flags     []string
		name      string
		expRatios []exchange.FeeRatio
		expErr    string
	}{
		{
			testName: "unknown flag",
			name:     "unknown",
			expErr:   "flag accessed but not defined: unknown",
		},
		{
			testName: "wrong flag type",
			name:     flagInt,
			expErr:   "trying to get stringSlice value of flag of type int",
		},
		{
			testName: "nothing provided",
			name:     flagStringSlice,
			expErr:   "",
		},
		{
			testName: "three vals, one bad",
			flags:    []string{"--" + flagStringSlice, "8apple:3apple,100pear:1apple", "--" + flagStringSlice, "cherry:777cherry"},
			name:     flagStringSlice,
			expRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("apple", 8), Fee: sdk.NewInt64Coin("apple", 3)},
				{Price: sdk.NewInt64Coin("pear", 100), Fee: sdk.NewInt64Coin("apple", 1)},
			},
			expErr: "cannot create FeeRatio from \"cherry:777cherry\": price: invalid coin expression: \"cherry\"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
			flagSet.StringSlice(flagStringSlice, nil, "A string slice")
			flagSet.Int(flagInt, 0, "An int")
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var ratios []exchange.FeeRatio
			testFunc := func() {
				ratios, err = cli.ReadFeeRatiosFlag(flagSet, tc.name)
			}
			require.NotPanics(t, testFunc, "ReadFeeRatiosFlag(%q)", tc.name)
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadFeeRatiosFlag(%q) error", tc.name)
			assertEqualSlices(t, tc.expRatios, ratios, exchange.FeeRatio.String, "ReadFeeRatiosFlag(%q) ratios", tc.name)
		})
	}
}

func TestParseFeeRatios(t *testing.T) {
	tests := []struct {
		name      string
		vals      []string
		expRatios []exchange.FeeRatio
		expErr    string
	}{
		{
			name:   "nil",
			vals:   nil,
			expErr: "",
		},
		{
			name:   "empty",
			vals:   []string{},
			expErr: "",
		},
		{
			name:   "one, bad",
			vals:   []string{"notaratio"},
			expErr: "cannot create FeeRatio from \"notaratio\": expected exactly one colon",
		},
		{
			name:      "one, good",
			vals:      []string{"10apple:3banana"},
			expRatios: []exchange.FeeRatio{{Price: sdk.NewInt64Coin("apple", 10), Fee: sdk.NewInt64Coin("banana", 3)}},
		},
		{
			name:      "one, zeros",
			vals:      []string{"0cherry:0durian"},
			expRatios: []exchange.FeeRatio{{Price: sdk.NewInt64Coin("cherry", 0), Fee: sdk.NewInt64Coin("durian", 0)}},
		},
		{
			name: "three, all good",
			vals: []string{"10apple:1cherry", "321banana:8grape", "66plum:7plum"},
			expRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("apple", 10), Fee: sdk.NewInt64Coin("cherry", 1)},
				{Price: sdk.NewInt64Coin("banana", 321), Fee: sdk.NewInt64Coin("grape", 8)},
				{Price: sdk.NewInt64Coin("plum", 66), Fee: sdk.NewInt64Coin("plum", 7)},
			},
		},
		{
			name: "three, first bad",
			vals: []string{"10apple", "321banana:8grape", "66plum:7plum"},
			expRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("banana", 321), Fee: sdk.NewInt64Coin("grape", 8)},
				{Price: sdk.NewInt64Coin("plum", 66), Fee: sdk.NewInt64Coin("plum", 7)},
			},
			expErr: "cannot create FeeRatio from \"10apple\": expected exactly one colon",
		},
		{
			name: "three, second bad",
			vals: []string{"10apple:1cherry", "8grape", "66plum:7plum"},
			expRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("apple", 10), Fee: sdk.NewInt64Coin("cherry", 1)},
				{Price: sdk.NewInt64Coin("plum", 66), Fee: sdk.NewInt64Coin("plum", 7)},
			},
			expErr: "cannot create FeeRatio from \"8grape\": expected exactly one colon",
		},
		{
			name: "three, third bad",
			vals: []string{"10apple:1cherry", "321banana:8grape", ""},
			expRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("apple", 10), Fee: sdk.NewInt64Coin("cherry", 1)},
				{Price: sdk.NewInt64Coin("banana", 321), Fee: sdk.NewInt64Coin("grape", 8)},
			},
			expErr: "cannot create FeeRatio from \"\": expected exactly one colon",
		},
		{
			name: "three, all bad",
			vals: []string{"10apple", "8grape", ""},
			expErr: joinErrs(
				"cannot create FeeRatio from \"10apple\": expected exactly one colon",
				"cannot create FeeRatio from \"8grape\": expected exactly one colon",
				"cannot create FeeRatio from \"\": expected exactly one colon",
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expRatios == nil {
				tc.expRatios = []exchange.FeeRatio{}
			}

			var ratios []exchange.FeeRatio
			var err error
			testFunc := func() {
				ratios, err = cli.ParseFeeRatios(tc.vals)
			}
			require.NotPanics(t, testFunc, "ParseFeeRatios(%q)", tc.vals)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseFeeRatios(%q) error", tc.vals)
			assertEqualSlices(t, tc.expRatios, ratios, exchange.FeeRatio.String, "ParseFeeRatios(%q) ratios", tc.vals)
		})
	}
}

func TestReadSplitsFlag(t *testing.T) {
	tests := []struct {
		testName  string
		flags     []string
		name      string
		expSplits []exchange.DenomSplit
		expErr    string
	}{
		{
			testName: "unknown flag",
			name:     "unknown",
			expErr:   "flag accessed but not defined: unknown",
		},
		{
			testName: "wrong flag type",
			name:     flagInt,
			expErr:   "trying to get stringSlice value of flag of type int",
		},
		{
			testName: "nothing provided",
			name:     flagStringSlice,
			expErr:   "",
		},
		{
			testName: "three vals, one bad",
			flags:    []string{"--" + flagStringSlice, "apple:3,banana:80q0", "--" + flagStringSlice, "cherry:777"},
			name:     flagStringSlice,
			expSplits: []exchange.DenomSplit{
				{Denom: "apple", Split: 3},
				{Denom: "cherry", Split: 777},
			},
			expErr: "could not parse \"banana:80q0\" amount: strconv.ParseUint: parsing \"80q0\": invalid syntax",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
			flagSet.StringSlice(flagStringSlice, nil, "A string slice")
			flagSet.Int(flagInt, 0, "An int")
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var splits []exchange.DenomSplit
			testFunc := func() {
				splits, err = cli.ReadSplitsFlag(flagSet, tc.name)
			}
			require.NotPanics(t, testFunc, "ReadSplitsFlag(%q)", tc.name)
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadSplitsFlag(%q) error", tc.name)
			assertEqualSlices(t, tc.expSplits, splits, splitStringer, "ReadSplitsFlag(%q) splits", tc.name)
		})
	}
}

func TestParseSplit(t *testing.T) {
	tests := []struct {
		name     string
		val      string
		expSplit *exchange.DenomSplit
		expErr   string
	}{
		{
			name:   "empty",
			val:    "",
			expErr: "invalid denom split \"\": expected format <denom>:<amount>",
		},
		{
			name:     "no colons",
			val:      "banana",
			expSplit: nil,
			expErr:   "invalid denom split \"banana\": expected format <denom>:<amount>",
		},
		{
			name:   "two colons",
			val:    "plum:8:123",
			expErr: "invalid denom split \"plum:8:123\": expected format <denom>:<amount>",
		},
		{
			name:   "empty denom",
			val:    ":444",
			expErr: "invalid denom split \":444\": both a <denom> and <amount> are required",
		},
		{
			name:   "empty amount",
			val:    "apple:",
			expErr: "invalid denom split \"apple:\": both a <denom> and <amount> are required",
		},
		{
			name:   "invalid amount",
			val:    "apple:banana",
			expErr: "could not parse \"apple:banana\" amount: strconv.ParseUint: parsing \"banana\": invalid syntax",
		},
		{
			name:     "good, zero",
			val:      "cherry:0",
			expSplit: &exchange.DenomSplit{Denom: "cherry", Split: 0},
		},
		{
			name:     "good, 10,000",
			val:      "pear:10000",
			expSplit: &exchange.DenomSplit{Denom: "pear", Split: 10000},
		},
		{
			name:     "good, 123",
			val:      "acorn:123",
			expSplit: &exchange.DenomSplit{Denom: "acorn", Split: 123},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var split *exchange.DenomSplit
			var err error
			testFunc := func() {
				split, err = cli.ParseSplit(tc.val)
			}
			require.NotPanics(t, testFunc, "ParseSplit(%q)", tc.val)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseSplit(%q) error", tc.val)
			if !assert.Equal(t, tc.expSplit, split, "ParseSplit(%q) split", tc.val) {
				t.Logf("Expected: %s:%d", tc.expSplit.Denom, tc.expSplit.Split)
				t.Logf("  Actual: %s:%d", split.Denom, split.Split)
			}
		})
	}
}

func TestParseSplits(t *testing.T) {
	tests := []struct {
		name      string
		vals      []string
		expSplits []exchange.DenomSplit
		expErr    string
	}{
		{
			name:   "nil",
			vals:   nil,
			expErr: "",
		},
		{
			name:   "empty",
			vals:   []string{},
			expErr: "",
		},
		{
			name:   "one, bad",
			vals:   []string{"nope"},
			expErr: "invalid denom split \"nope\": expected format <denom>:<amount>",
		},
		{
			name:      "one, good",
			vals:      []string{"yup:5"},
			expSplits: []exchange.DenomSplit{{Denom: "yup", Split: 5}},
		},
		{
			name: "three, all good",
			vals: []string{"first:1", "second:22", "third:333"},
			expSplits: []exchange.DenomSplit{
				{Denom: "first", Split: 1}, {Denom: "second", Split: 22}, {Denom: "third", Split: 333},
			},
		},
		{
			name: "three, first bad",
			vals: []string{"first", "second:22", "third:333"},
			expSplits: []exchange.DenomSplit{
				{Denom: "second", Split: 22}, {Denom: "third", Split: 333},
			},
			expErr: "invalid denom split \"first\": expected format <denom>:<amount>",
		},
		{
			name: "three, second bad",
			vals: []string{"first:1", ":22", "third:333"},
			expSplits: []exchange.DenomSplit{
				{Denom: "first", Split: 1}, {Denom: "third", Split: 333},
			},
			expErr: "invalid denom split \":22\": both a <denom> and <amount> are required",
		},
		{
			name: "three, third bad",
			vals: []string{"first:1", "second:22", "third:333x"},
			expSplits: []exchange.DenomSplit{
				{Denom: "first", Split: 1}, {Denom: "second", Split: 22},
			},
			expErr: "could not parse \"third:333x\" amount: strconv.ParseUint: parsing \"333x\": invalid syntax",
		},
		{
			name: "three, all bad",
			vals: []string{"first", ":22", "third:333x"},
			expErr: joinErrs(
				"invalid denom split \"first\": expected format <denom>:<amount>",
				"invalid denom split \":22\": both a <denom> and <amount> are required",
				"could not parse \"third:333x\" amount: strconv.ParseUint: parsing \"333x\": invalid syntax",
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expSplits == nil {
				tc.expSplits = []exchange.DenomSplit{}
			}

			var splits []exchange.DenomSplit
			var err error
			testFunc := func() {
				splits, err = cli.ParseSplits(tc.vals)
			}
			require.NotPanics(t, testFunc, "ParseSplits(%q)", tc.vals)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseSplits(%q) error", tc.vals)
			assertEqualSlices(t, tc.expSplits, splits, splitStringer, "ParseSplits(%q) splits", tc.vals)
		})
	}
}

func TestReadStringFlagOrArg(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		args     []string
		flagName string
		varName  string
		expStr   string
		expErr   string
	}{
		{
			name:     "unknown flag name",
			flagName: "other",
			varName:  "nope",
			expErr:   "flag accessed but not defined: other",
		},
		{
			name:     "wrong flag type",
			flagName: flagInt,
			varName:  "number",
			expErr:   "trying to get string value of flag of type int",
		},
		{
			name:     "both flag and arg",
			flags:    []string{"--" + flagString, "flagval"},
			args:     []string{"argval"},
			flagName: flagString,
			varName:  "value",
			expErr:   "cannot provide <value> as both an arg (\"argval\") and flag (--" + flagString + " \"flagval\")",
		},
		{
			name:     "only flag",
			flags:    []string{"--" + flagString, "flagval"},
			flagName: flagString,
			varName:  "value",
			expStr:   "flagval",
		},
		{
			name:     "only arg",
			args:     []string{"argval"},
			flagName: flagString,
			varName:  "value",
			expStr:   "argval",
		},
		{
			name:     "neither flag nor arg",
			flagName: flagString,
			varName:  "value",
			expErr:   "no <value> provided",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
			flagSet.String(flagString, "", "A string")
			flagSet.Int(flagInt, 0, "An int")
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var str string
			testFunc := func() {
				str, err = cli.ReadStringFlagOrArg(flagSet, tc.args, tc.flagName, tc.varName)
			}
			require.NotPanics(t, testFunc, "ReadStringFlagOrArg")
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadStringFlagOrArg error")
			assert.Equal(t, tc.expStr, str, "ReadStringFlagOrArg string")
		})
	}
}
