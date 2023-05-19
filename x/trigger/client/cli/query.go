package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/trigger/types"
)

var cmdStart = fmt.Sprintf("%s query trigger", version.AppName)

// GetQueryCmd is the top-level command for trigger CLI queries.
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the triggers module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(GetTriggersCmd())
	return queryCmd
}

// GetTriggersCmd queries for one trigger by id or all depending on the input
func GetTriggersCmd() *cobra.Command {
	const all = "all"
	cmd := &cobra.Command{
		Use:     "list {trigger_id|\"all\"}",
		Aliases: []string{"ls", "l"},
		Short:   "Query the current triggers",
		Long: fmt.Sprintf(`%[1]s trigger {trigger_id} - gets the trigger for a given id.
%[1]s reward-program all - gets all the triggers`, cmdStart),
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s trigger 1
%[1]s trigger all`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			var request types.QueryTriggersRequest
			arg0 := strings.TrimSpace(args[0])
			if arg0 != all {
				return queryTriggerByID(clientCtx, queryClient, arg0)
			}

			var response *types.QueryTriggersResponse
			if response, err = queryClient.Triggers(
				context.Background(),
				&request,
			); err != nil {
				return fmt.Errorf("failed to query triggers: %w", err)
			}

			return clientCtx.PrintProto(response)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all")
	return cmd
}

// queryTriggerByID queries for one trigger by id.
func queryTriggerByID(client client.Context, queryClient types.QueryClient, arg string) error {
	triggerID, err := strconv.Atoi(arg)
	if err != nil {
		return fmt.Errorf("invalid argument arg : %s", arg)
	}

	var response *types.QueryTriggerByIDResponse
	if response, err = queryClient.TriggerByID(
		context.Background(),
		&types.QueryTriggerByIDRequest{Id: uint64(triggerID)},
	); err != nil {
		return fmt.Errorf("failed to query trigger %d: %w", triggerID, err)
	}

	if response.GetTrigger() == nil {
		return fmt.Errorf("trigger %d does not exist", triggerID)
	}

	return client.PrintProto(response)
}
