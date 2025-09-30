package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	"github.com/provenance-io/provenance/internal/provcli"
	hold "github.com/provenance-io/provenance/x/hold"
)

const (
	FlagAddresses     = "addresses"      //nolint:revive
	FlagAddressesFile = "addresses-file" //nolint:revive
)

// exampleTxCmdBase is the base command that gets a user to one of the query commands in here.
var exampleTxCmdBase = fmt.Sprintf("%s tx %s", version.AppName, hold.ModuleName)

// NewTxCmd returns the top-level command for hold CLI transactions
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        hold.ModuleName,
		Short:                      "Transaction commands for the hold module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		GetCmdUnlockVestingAccounts(),
	)
	return txCmd
}

// GetCmdUnlockVestingAccounts creates a governance proposal to unlock vesting accounts
func GetCmdUnlockVestingAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unlock-vesting-accounts [--addresses <addr1>[,<addr2> ...]] [--addresses-file <filename>] [gov prop and tx flags]",
		Aliases: []string{"unlock-vesting", "unlock-vesting-account"},
		Args:    cobra.NoArgs,
		Short:   "Submit a governance proposal to unlock vesting accounts",
		Long: strings.TrimSpace(fmt.Sprintf(`Submit a governance proposal to convert vesting accounts back to base accounts.

At least one address must be provided using either the --%[2]s and/or --%[3]s flags.

Examples:
Command-line addresses:
$ %[1]s unlock-vesting-accounts \
  --title "Unlock Vesting Accounts" \
  --description "Convert vesting accounts to base accounts" \
  --addresses addr1,addr2,addr3 \
  --deposit 10000000nhash

File input:
$ %[1]s unlock-vesting-accounts \
  --title "Unlock Vesting Accounts" \
  --description "Convert vesting accounts to base accounts" \
  --addresses-file addresses.txt \
  --deposit 10000000nhash

The provided file should contain bech32 address strings, one per line.
`, exampleTxCmdBase, FlagAddresses, FlagAddressesFile)),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()
			msg := &hold.MsgUnlockVestingAccountsRequest{
				Authority: provcli.GetAuthority(flagSet),
			}
			msg.Addresses, err = getAddressesFromFlags(flagSet)
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true
			return provcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
		},
	}

	govcli.AddGovPropFlagsToCmd(cmd)
	provcli.AddAuthorityFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().StringSlice(FlagAddresses, []string{}, "Comma-separated list of addresses to unlock")
	cmd.Flags().String(FlagAddressesFile, "", "Path to a file containing the addresses to unlock")

	return cmd
}

// getAddressesFromFlags retrieves addresses from flags and/or file.
func getAddressesFromFlags(flagSet *pflag.FlagSet) ([]string, error) {
	addrs, err := readAddressesFlag(flagSet)
	if err != nil {
		return nil, err
	}
	fileAddrs, err := readAddressesFileFlag(flagSet)
	if err != nil {
		return nil, err
	}

	return append(fileAddrs, addrs...), nil
}

// readAddressesFlag returns the addresses provided with the --addresses flag.
func readAddressesFlag(flagSet *pflag.FlagSet) ([]string, error) {
	addrs, err := flagSet.GetStringSlice(FlagAddresses)
	if err != nil {
		return nil, fmt.Errorf("could not read --%q flag: %w", FlagAddresses, err)
	}
	return addrs, nil
}

// readAddressesFileFlag will get the value of the --addresses-file flag, read the file, and return the its contents.
func readAddressesFileFlag(flagSet *pflag.FlagSet) ([]string, error) {
	path, err := flagSet.GetString(FlagAddressesFile)
	if err != nil {
		return nil, fmt.Errorf("could not read --%q flag: %w", FlagAddressesFile, err)
	}
	if len(path) == 0 {
		return nil, nil
	}

	data, err := os.ReadFile(path) //nolint:gosec // G304
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	addrs := make([]string, 0, len(lines))
	for _, line := range lines {
		l2 := strings.TrimSpace(line)
		if len(l2) != 0 {
			addrs = append(addrs, l2)
		}
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("no addresses found in file: %s", path)
	}
	return addrs, nil
}
