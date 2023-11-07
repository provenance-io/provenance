package cli

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/exchange"
)

var (
	AuthorityAddr = authtypes.NewModuleAddress(govtypes.ModuleName)

	ExampleAddr1 = "pb1g4uxzmtsd3j5zerywgc47h6lta047h6lwwxvlw" // = sdk.AccAddress("ExampleAddr1________")
	ExampleAddr2 = "pb1tazhsctdwpkx2styv3eryh6lta047h6l63dw8r" // = sdk.AccAddress("_ExampleAddr2_______")
	ExampleAddr3 = "pb195k527rpd4cxce2pv3j8yv6lta047h6l3kaj79" // = sdk.AccAddress("--ExampleAddr3______")
	ExampleAddr4 = "pb10el8u3tcv9khqmr9g9jxgu35ta047h6l9hc7xs" // = sdk.AccAddress("~~~ExampleAddr4_____")
	ExampleAddr5 = "pb1857n60290psk6urvv4qkgerjx4047h6l5vynnz" // = sdk.AccAddress("====ExampleAddr5____")
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

		return govcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
	}
}

// MarkFlagRequired this marks a flag as required and panics if there's a problem.
func MarkFlagRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(fmt.Errorf("error marking --%s flag required on %s: %w", name, cmd.Name(), err))
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
	rv := make([]exchange.AccessGrant, 0, len(vals))
	for _, val := range vals {
		ag, err := ParseAccessGrant(val)
		if err != nil {
			errs = append(errs, err)
		}
		if ag != nil {
			rv = append(rv, *ag)
		}
	}
	return rv, errors.Join(errs...)
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
	rv := make([]exchange.FeeRatio, 0, len(vals))
	for _, val := range vals {
		ratio, err := exchange.ParseFeeRatio(val)
		if err != nil {
			errs = append(errs, err)
		}
		if ratio != nil {
			rv = append(rv, *ratio)
		}
	}
	return rv, errors.Join(errs...)
}
