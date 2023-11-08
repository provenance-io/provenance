package cli

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/exchange"
)

var (
	// AuthorityAddr is the governance module's account address.
	AuthorityAddr = authtypes.NewModuleAddress(govtypes.ModuleName)

	// ExampleAddr is an example bech32 address to use in command descriptions and stuff.
	ExampleAddr = "pb1g4uxzmtsd3j5zerywf047h6lta047h6lycmzwe" // = sdk.AccAddress("ExampleAddr_________")
)

// A msgMaker is a function that makes a Msg from a client.Context, FlagSet, and set of args.
type msgMaker func(clientCtx client.Context, flagSet *pflag.FlagSet, args []string) (sdk.Msg, error)

// genericTxRunE returns a cobra.Command.RunE function that gets the client.Context, and FlagSet,
// then uses the provided maker to make the Msg that it then generates or broadcasts as a Tx.
func genericTxRunE(maker msgMaker) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		flagSet := cmd.Flags()
		msg, err := maker(clientCtx, flagSet, args)
		if err != nil {
			return err
		}

		cmd.SilenceUsage = true
		return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, msg)
	}
}

// govTxRunE returns a cobra.Command.RunE function that gets the client.Context, and FlagSet,
// then uses the provided maker to make the Msg. The Msg is then either generated or put in a
// governance proposal which is then broadcast as a Tx.
func govTxRunE(maker msgMaker) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		flagSet := cmd.Flags()
		msg, err := maker(clientCtx, flagSet, args)
		if err != nil {
			return err
		}

		cmd.SilenceUsage = true
		return govcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
	}
}

// MarkFlagsRequired marks the provided flags as required and panics if there's a problem.
func MarkFlagsRequired(cmd *cobra.Command, names ...string) {
	for _, name := range names {
		if err := cmd.MarkFlagRequired(name); err != nil {
			panic(fmt.Errorf("error marking --%s flag required on %s: %w", name, cmd.Name(), err))
		}
	}
}

// ReadCoinsFlag reads a string flag and converts it into sdk.Coins.
// If the flag wasn't provided, this returns nil, nil.
func ReadCoinsFlag(flagSet *pflag.FlagSet, name string) (sdk.Coins, error) {
	value, err := flagSet.GetString(name)
	if len(value) == 0 || err != nil {
		return nil, err
	}
	rv, err := sdk.ParseCoinsNormalized(value)
	if err != nil {
		return nil, fmt.Errorf("error parsing --%s value %q as coins: %w", name, value, err)
	}
	return rv, nil
}

// ReadCoinFlag reads a string flag and converts it into *sdk.Coin.
// If the flag wasn't provided, this returns nil, nil.
func ReadCoinFlag(flagSet *pflag.FlagSet, name string) (*sdk.Coin, error) {
	value, err := flagSet.GetString(name)
	if len(value) == 0 || err != nil {
		return nil, err
	}
	rv, err := exchange.ParseCoin(value)
	if err != nil {
		return nil, fmt.Errorf("error parsing --%s value %q as a coin: %w", name, value, err)
	}
	return &rv, nil
}

// ReadReqCoinFlag reads a string flag and converts it into a sdk.Coin and requires it to have a value.
func ReadReqCoinFlag(flagSet *pflag.FlagSet, name string) (sdk.Coin, error) {
	rv, err := ReadCoinFlag(flagSet, name)
	if err != nil {
		return sdk.Coin{}, err
	}
	if rv == nil {
		return sdk.Coin{}, fmt.Errorf("missing required --%s flag", name)
	}
	return *rv, nil
}

// ReadOrderIdsFlag reads a UintSlice flag and converts it into a []uint64.
func ReadOrderIdsFlag(flagSet *pflag.FlagSet, name string) ([]uint64, error) {
	ids, err := flagSet.GetUintSlice(name)
	if err != nil {
		return nil, err
	}
	rv := make([]uint64, len(ids))
	for i, id := range ids {
		rv[i] = uint64(id)
	}
	return rv, nil
}

// ReadAddrOrDefault gets the requested flag or, if it wasn't provided, gets the --from address.
func ReadAddrOrDefault(clientCtx client.Context, flagSet *pflag.FlagSet, name string) (string, error) {
	rv, err := flagSet.GetString(name)
	if len(rv) > 0 || err != nil {
		return rv, err
	}

	rv = clientCtx.FromAddress.String()
	if len(rv) > 0 {
		return rv, nil
	}

	return "", fmt.Errorf("no %s provided", name)
}

// ReadAccessGrants reads a StringSlice flag and converts it to a slice of AccessGrants.
func ReadAccessGrants(flagSet *pflag.FlagSet, name string) ([]exchange.AccessGrant, error) {
	vals, err := flagSet.GetStringSlice(name)
	if len(vals) == 0 || err != nil {
		return nil, err
	}
	return ParseAccessGrants(vals)
}

// ReadFlatFee reads a StringSlice flag and converts it into a slice of sdk.Coin.
func ReadFlatFee(flagSet *pflag.FlagSet, name string) ([]sdk.Coin, error) {
	vals, err := flagSet.GetStringSlice(name)
	if len(vals) == 0 || err != nil {
		return nil, err
	}
	return ParseFlatFeeOptions(vals)
}

// ReadFeeRatios reads a StringSlice flag and converts it into a slice of exchange.FeeRatio.
func ReadFeeRatios(flagSet *pflag.FlagSet, name string) ([]exchange.FeeRatio, error) {
	vals, err := flagSet.GetStringSlice(name)
	if len(vals) == 0 || err != nil {
		return nil, err
	}
	return ParseFeeRatios(vals)
}

// ReadFlagSplits reads a StringSlice flag and converts it into a slice of exchange.DenomSplit.
func ReadFlagSplits(flagSet *pflag.FlagSet, name string) ([]exchange.DenomSplit, error) {
	vals, err := flagSet.GetStringSlice(FlagSplit)
	if len(vals) == 0 || err != nil {
		return nil, err
	}
	return ParseSplits(vals)
}

// permSepRx is a regexp that matches characters that can be used to separate permissions.
var permSepRx = regexp.MustCompile(`[ +-.]`)

// ParseAccessGrant parses an AccessGrant from a string with the format "<address>:<perm 1>[+<perm 2>...]".
func ParseAccessGrant(val string) (*exchange.AccessGrant, error) {
	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("could not parse %q as an <access grant>: expected format <address>:<permissions>", val)
	}

	addr := strings.TrimSpace(parts[0])
	perms := strings.ToLower(strings.TrimSpace(parts[1]))
	if len(addr) == 0 || len(perms) == 0 {
		return nil, fmt.Errorf("invalid <access grant>: both an <address> and <permissions> are required")
	}

	rv := &exchange.AccessGrant{Address: addr}

	if perms == "all" {
		rv.Permissions = exchange.AllPermissions()
		return rv, nil
	}

	permVals := permSepRx.Split(perms, -1)
	var err error
	rv.Permissions, err = exchange.ParsePermissions(permVals...)
	if err != nil {
		return nil, fmt.Errorf("could not parse permissions for %q from %q: %w", rv.Address, parts[1], err)
	}

	return rv, nil
}

// ParseAccessGrants parses an AccessGrant from each of the provided vals.
func ParseAccessGrants(vals []string) ([]exchange.AccessGrant, error) {
	var errs []error
	grants := make([]exchange.AccessGrant, 0, len(vals))
	for _, val := range vals {
		ag, err := ParseAccessGrant(val)
		if err != nil {
			errs = append(errs, err)
		}
		if ag != nil {
			grants = append(grants, *ag)
		}
	}
	return grants, errors.Join(errs...)
}

// ParseFlatFeeOptions parses an sdk.Coin from each of the provided vals.
func ParseFlatFeeOptions(vals []string) ([]sdk.Coin, error) {
	var errs []error
	rv := make([]sdk.Coin, 0, len(vals))
	for _, val := range vals {
		coin, err := exchange.ParseCoin(val)
		if err != nil {
			errs = append(errs, err)
		} else {
			rv = append(rv, coin)
		}
	}
	return rv, errors.Join(errs...)
}

// ParseFeeRatios parses a FeeRatio from each of the provided vals.
func ParseFeeRatios(vals []string) ([]exchange.FeeRatio, error) {
	var errs []error
	ratios := make([]exchange.FeeRatio, 0, len(vals))
	for _, val := range vals {
		ratio, err := exchange.ParseFeeRatio(val)
		if err != nil {
			errs = append(errs, err)
		}
		if ratio != nil {
			ratios = append(ratios, *ratio)
		}
	}
	return ratios, errors.Join(errs...)
}

// ParseSplit parses a DenomSplit from a string with teh format "<denom>:<amount>".
func ParseSplit(val string) (*exchange.DenomSplit, error) {
	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid denom split %q: expected format <denom>:<amount>")
	}

	denom := strings.TrimSpace(parts[0])
	amountStr := strings.TrimSpace(parts[1])
	if len(denom) == 0 || len(amountStr) == 0 {
		return nil, fmt.Errorf("invalid denom split %q: both a <denom> and <amount> are required", val)
	}

	amount, err := strconv.ParseUint(amountStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("could not parse %s amount %q into uint32: %w", denom, amountStr, err)
	}

	return &exchange.DenomSplit{Denom: denom, Split: uint32(amount)}, nil
}

// ParseSplits parses a DenomSplit from each of the provided vals.
func ParseSplits(vals []string) ([]exchange.DenomSplit, error) {
	var errs []error
	splits := make([]exchange.DenomSplit, 0, len(vals))
	for _, val := range vals {
		split, err := ParseSplit(val)
		if err != nil {
			errs = append(errs, err)
		}
		if split != nil {
			splits = append(splits, *split)
		}
	}
	return splits, errors.Join(errs...)
}

// AddUseArgs adds the given args to the cmd's Use.
func AddUseArgs(cmd *cobra.Command, args ...string) {
	cmd.Use = cmd.Use + " " + strings.Join(args, " ")
}

// AddUseDetails appends each provided section to the Use field with an empty line between them.
func AddUseDetails(cmd *cobra.Command, sections ...string) {
	cmd.Use = cmd.Use + "\n\n" + strings.Join(sections, "\n\n")
	cmd.DisableFlagsInUseLine = true
}

// SimplePerms returns a string containing all the Permission.SimpleString() values.
func SimplePerms() string {
	allPerms := exchange.AllPermissions()
	simple := make([]string, len(allPerms))
	for i, perm := range allPerms {
		simple[i] = perm.SimpleString()
	}
	return strings.Join(simple, "  ")
}

// ReqSignerDesc returns a description of how the --<name> flag is used and sort of required.
func ReqSignerDesc(name string) string {
	return fmt.Sprintf(`If --%[1]s <%[1]s> is provided, that is used as the %[1]s.
If no --%[1]s is provided, the --%[2]s account address is used as the %[1]s.
A %[1]s is required.`,
		name, flags.FlagFrom,
	)
}

// ReqSignerUse is the Use string for a signer flag.
func ReqSignerUse(name string) string {
	return fmt.Sprintf("{--%s|--%s} <%s>", flags.FlagFrom, name, name)
}

// ReqFlagUse returns the string "--name <opt>" if an opt is provided, or just "--name" if not.
func ReqFlagUse(name string, opt string) string {
	if len(opt) > 0 {
		return fmt.Sprintf("--%s <%s>", name, opt)
	}
	return "--" + name
}

// OptFlagUse wraps a ReqFlagUse in [], e.g. "[--name <opt>]"
func OptFlagUse(name string, opt string) string {
	return "[" + ReqFlagUse(name, opt) + "]"
}

var (
	// UseFlagsBreak is a string to use to start a new line of flags in the Use.
	UseFlagsBreak = "\n     "

	// RepeatableDesc is a description of how repeatable flags/values can be provided.
	RepeatableDesc = "If a flag is repeatable, multiple entries can be separated by commas\nand/or the flag can be provided multiple times."

	// ReqAdminUse is the Use string of the --admin flag.
	ReqAdminUse = fmt.Sprintf("{--%s|--%s} <admin>", flags.FlagFrom, FlagAdmin)

	// ReqAdminDesc is a description of how the --admin, --authority, and --from flags work and are sort of required.
	ReqAdminDesc = fmt.Sprintf(`If --%[1]s <admin> is provided, that is used as the admin.
If no --%[1]s is provided, but the --%[2]s flag was, the governance module account is used as the admin.
Otherwise the --%[3]s account address is used as the admin.
An admin is required.`,
		FlagAdmin, FlagAuthority, flags.FlagFrom,
	)

	// ReqEnableDisableUse is a use string for the --enable and --disable flags.
	ReqEnableDisableUse = fmt.Sprintf("{--%s|--%s}", FlagEnable, FlagDisable)

	// ReqEnableDisableDesc is a description of the --enable and --disable flags.
	ReqEnableDisableDesc = fmt.Sprintf("One of --%s or --%s must be provided, but not both.", FlagEnable, FlagDisable)

	// AccessGrantsDesc is a description of the <asset grant> format.
	AccessGrantsDesc = fmt.Sprintf(`An <access grant> has the format "<address>:<permissions>"
In <permissions>, separate each permission with a + (plus), - (dash), or . (period).
An <access grant> of "<address>:all" will have all of the permissions.

Example <access grant>: %s:settle+update

Valid permissions entries: %s
The full Permission enum names are also valid.`,
		ExampleAddr,
		SimplePerms(),
	)

	// FeeRatioDesc is a description of the <fee ratio> format.
	FeeRatioDesc = `A <fee ratio> has the format "<price coin>:<fee coin>".
Both <price coin> and <fee coin> have the format "<amount><denom>".

Example <fee ratio>: 100nhash:1nhash`

	AuthorityDesc = fmt.Sprintf("If --%s is not provided, the governance module account is used as the authority.", FlagAuthority)
)
