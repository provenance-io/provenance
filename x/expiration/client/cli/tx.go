package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/expiration/types"
)

const (
	FlagSigners = "signers"
)

// NewTxCmd is the top-level command for expiration CLI transactions
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"exp"},
		Short:                      "Transaction commands for the expiration module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		AddExpirationCmd(),
		ExtendExpirationCmd(),
		DeleteExpirationCmd(),
	)
	return txCmd
}

// AddExpirationCmd creates a command for adding an expiration
func AddExpirationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [module-asset-id] [owner] [block-height] [deposit] [messages, optional]",
		Aliases: []string{"a"},
		Short:   "Add module asset expiration on the provenance blockchain",
		Long: `Add module asset expiration on the provenance blockchain
module-asset-id - the id of the module asset
owner			- the owner of the module asset
block-height	- the block height the module asset will expire and be removed from the provenance blockchain
deposit			- deposit held for storing assets on provenance blockchain
messages		- comma separated list of messages to add to the expiration (optional)`,
		Example: strings.TrimSpace(
			fmt.Sprintf(`
				$ %[1]s tx expiration add pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 1000000 10000nhash
				$ %[1]s tx expiration add pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 1000000 10000nhash message1,message2
				`,
				version.AppName,
			)),
		Args: cobra.RangeArgs(4, 5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			blockHeight, err := parseBlockHeight(args[2])
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoinNormalized(args[3])
			if err != nil {
				return err
			}
			//messages := parseMessages(args[4])

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			expiration := types.Expiration{
				ModuleAssetId: strings.TrimSpace(args[0]),
				Owner:         strings.TrimSpace(args[1]),
				BlockHeight:   blockHeight,
				Deposit:       deposit,
				// TODO	How do we add message of type 'Any'?
				//      Should we only support string messages?
				//Messages:      messages,
			}

			msg := types.NewMsgAddExpirationRequest(expiration, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// ExtendExpirationCmd creates a command for extending an expiration
func ExtendExpirationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extend [module-asset-id] [owner] [block-height] [deposit] [messages, optional]",
		Aliases: []string{"e"},
		Short:   "Extend module asset expiration on the provenance blockchain",
		Long: `Extend module asset expiration on the provenance blockchain
module-asset-id - the id of the module asset
owner			- the owner of the module asset
block-height	- the block height the module asset will expire and be removed from the provenance blockchain
deposit			- deposit held for storing assets on provenance blockchain
messages		- comma separated list of messages to add to the expiration (optional)`,
		Example: strings.TrimSpace(
			fmt.Sprintf(`
				$ %[1]s tx expiration extend pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 1000000 10000nhash
				$ %[1]s tx expiration extend pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 1000000 10000nhash message1,message2
				`,
				version.AppName,
			)),
		Args: cobra.RangeArgs(4, 5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			blockHeight, err := parseBlockHeight(args[2])
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoinNormalized(args[3])
			if err != nil {
				return err
			}
			//messages := parseMessages(args[4])

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			expiration := types.Expiration{
				ModuleAssetId: strings.TrimSpace(args[0]),
				Owner:         strings.TrimSpace(args[1]),
				BlockHeight:   blockHeight,
				Deposit:       deposit,
				//Messages:      messages,
			}

			msg := types.NewMsgExtendExpirationRequest(expiration, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// DeleteExpirationCmd creates a command for deleting an expiration
func DeleteExpirationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [module-asset-id]",
		Aliases: []string{"d"},
		Short:   "Extend module asset expiration on the provenance blockchain",
		Example: fmt.Sprintf(`$ %s tx expiration delete pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			moduleAssetId := args[0]
			msg := types.NewMsgDeleteExpirationRequest(moduleAssetId, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func parseBlockHeight(s string) (int64, error) {
	blockHeight, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	if blockHeight < 0 {
		return 0, fmt.Errorf("block height cannot be negative: %d", blockHeight)
	}
	return blockHeight, nil
}

func parseMessages(messages string) []string {
	if len(messages) == 0 {
		return nil
	} else {
		return strings.Split(messages, ",")
	}
}

func addSignerFlagCmd(cmd *cobra.Command) {
	cmd.Flags().String(FlagSigners, "", "comma delimited list of bech32 addresses")
}

// parseSigners checks signers flag for signers, else uses the from address
func parseSigners(cmd *cobra.Command, client *client.Context) ([]string, error) {
	flagSet := cmd.Flags()
	if flagSet.Changed(FlagSigners) {
		signerList, _ := flagSet.GetString(FlagSigners)
		signers := strings.Split(signerList, ",")
		for _, signer := range signers {
			_, err := sdk.AccAddressFromBech32(signer)
			if err != nil {
				fmt.Printf("signer address must be a Bech32 string: %v", err)
				return nil, err
			}
		}
		return signers, nil
	}
	return []string{client.GetFromAddress().String()}, nil
}
