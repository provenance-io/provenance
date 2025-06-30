package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	"github.com/provenance-io/provenance/internal/provcli"
	holdmodule "github.com/provenance-io/provenance/x/hold"
	"github.com/provenance-io/provenance/x/hold/types"
)

const (
	FlagAddresses     = "addresses"
	FlagAddressesFile = "addresses-file"
)

// NewTxCmd returns the top-level command for hold CLI transactions
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        holdmodule.ModuleName,
		Short:                      "Transaction commands for the hold module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		GetCmdUnlockVestingAccountsProposal(),
	)
	return txCmd
}

// GetCmdUnlockVestingAccountsProposal creates a governance proposal to unlock vesting accounts
func GetCmdUnlockVestingAccountsProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock-vesting-accounts-proposal",
		Args:  cobra.NoArgs,
		Short: "Submit a governance proposal to unlock vesting accounts",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a governance proposal to convert vesting accounts back to base accounts.

Supports both command-line addresses and JSON files for large address sets.

Examples:
Command-line addresses:
$ %[1]s tx gov submit-proposal unlock-vesting-accounts-proposal \
  --title "Unlock Vesting Accounts" \
  --description "Convert vesting accounts to base accounts" \
  --addresses addr1,addr2,addr3 \
  --deposit 10000000nhash

JSON file input:
$ %[1]s tx gov submit-proposal unlock-vesting-accounts-proposal \
  --title "Unlock Vesting Accounts" \
  --description "Convert vesting accounts to base accounts" \
  --addresses-file addresses.json \
  --deposit 10000000nhash

JSON file format (array of strings):
["addr1", "addr2", "addr3"]
`, holdmodule.ModuleName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			addresses, err := getAddresses(cmd)
			if err != nil {
				return err
			}
			for _, addr := range addresses {
				if _, err := sdk.AccAddressFromBech32(addr); err != nil {
					return sdkErrors.ErrInvalidAddress.Wrapf("invalid address %q: %w", addr, err)
				}
			}

			flagSet := cmd.Flags()
			authority := provcli.GetAuthority(flagSet)
			unlockMsg := types.NewMsgUnlockVestingAccounts(authority, addresses)

			return provcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, unlockMsg)
		},
	}

	govcli.AddGovPropFlagsToCmd(cmd)
	provcli.AddAuthorityFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().StringSlice(FlagAddresses, []string{}, "Comma-separated list of addresses to unlock")
	cmd.Flags().String(FlagAddressesFile, "", "Path to a JSON file containing an array of addresses to unlock")

	return cmd
}

// getAddresses retrieves addresses from flags or file and enforces mutual exclusivity
func getAddresses(cmd *cobra.Command) ([]string, error) {
	filePath, err := cmd.Flags().GetString(FlagAddressesFile)
	if err != nil {
		return nil, sdkErrors.ErrIO.Wrapf("get flag %s: %w", FlagAddressesFile, err)
	}

	addressList, err := cmd.Flags().GetStringSlice(FlagAddresses)
	if err != nil {
		return nil, sdkErrors.ErrInvalidAddress.Wrapf("get flag %s: %w", FlagAddresses, err)
	}

	if filePath != "" && len(addressList) > 0 {
		return nil, sdkErrors.ErrInvalidAddress.Wrapf("only one of --addresses or --addresses-file can be specified")
	}

	if filePath != "" {
		return parseAddressFile(filePath)
	}

	if len(addressList) == 0 {
		return nil, sdkErrors.ErrInvalidAddress.Wrapf("no addresses provided")
	}

	return addressList, nil
}

// parseAddressFile reads addresses from a JSON file
func parseAddressFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, sdkErrors.ErrInvalidType.Wrapf("read file: %w", err)
	}
	var addresses []string
	if err := json.Unmarshal(data, &addresses); err != nil {
		return nil, sdkErrors.ErrInvalidType.Wrapf("parse JSON: %w", err)
	}
	if len(addresses) == 0 {
		return nil, sdkErrors.ErrInvalidAddress.Wrapf("no addresses in file")
	}
	return addresses, nil
}
