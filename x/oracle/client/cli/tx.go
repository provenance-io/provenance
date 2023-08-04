package cli

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/provenance-io/provenance/x/oracle/types"
	"github.com/spf13/cobra"
)

// NewTxCmd is the top-level command for oracle CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"t"},
		Short:                      "Transaction commands for the oracle module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdSendQuery(),
		GetCmdOracleUpdate(),
	)

	return txCmd
}

// GetCmdOracleUpdate is a command to update the address of the module's oracle
func GetCmdOracleUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <address>",
		Short:   "Update the module's oracle address",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"u"},
		Example: fmt.Sprintf(`%[1]s tx oracle update pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateOracle(
				clientCtx.GetFromAddress().String(),
				args[0],
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdSendQuery is a command to send a query to another chain's oracle
func GetCmdSendQuery() *cobra.Command {
	decoder := newArgDecoder(asciiDecodeString)
	cmd := &cobra.Command{
		Use:     "send-query <channel-id> <json>",
		Short:   "Send a query to an oracle on a remote chain via IBC",
		Args:    cobra.ExactArgs(2),
		Aliases: []string{"sq"},
		Example: fmt.Sprintf(`%[1]s tx oracle send-query channel-1 '{"query_version":{}}'`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			channelID := args[0]

			queryData, err := decoder.DecodeString(args[1])
			if err != nil {
				return fmt.Errorf("decode query: %s", err)
			}
			if !json.Valid(queryData) {
				return errors.New("query data must be json")
			}

			msg := types.NewMsgQueryOracle(
				clientCtx.GetFromAddress().String(),
				channelID,
				queryData,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

type argumentDecoder struct {
	// dec is the default decoder
	dec                func(string) ([]byte, error)
	asciiF, hexF, b64F bool
}

func newArgDecoder(def func(string) ([]byte, error)) *argumentDecoder {
	return &argumentDecoder{dec: def}
}

// DecodeString decodes the supplied json string
func (a *argumentDecoder) DecodeString(s string) ([]byte, error) {
	found := -1
	for i, v := range []*bool{&a.asciiF, &a.hexF, &a.b64F} {
		if !*v {
			continue
		}
		if found != -1 {
			return nil, errors.New("multiple decoding flags used")
		}
		found = i
	}
	switch found {
	case 0:
		return asciiDecodeString(s)
	case 1:
		return hex.DecodeString(s)
	case 2:
		return base64.StdEncoding.DecodeString(s)
	default:
		return a.dec(s)
	}
}

func asciiDecodeString(s string) ([]byte, error) {
	return []byte(s), nil
}
