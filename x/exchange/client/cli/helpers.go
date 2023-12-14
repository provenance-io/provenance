package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	// govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli" // TODO[1760]: gov-cli
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/x/exchange"
)

var (
	// AuthorityAddr is the governance module's account address.
	// It's not converted to a string here because the global HRP probably isn't set when this is being defined.
	AuthorityAddr = authtypes.NewModuleAddress(govtypes.ModuleName)

	// ExampleAddr is an example bech32 address to use in command descriptions and stuff.
	ExampleAddr = "pb1g4uxzmtsd3j5zerywf047h6lta047h6lycmzwe" // = sdk.AccAddress("ExampleAddr_________")
)

// A msgMaker is a function that makes a Msg from a client.Context, FlagSet, and set of args.
//
// R is the type of the Msg.
type msgMaker[R sdk.Msg] func(clientCtx client.Context, flagSet *pflag.FlagSet, args []string) (R, error)

// genericTxRunE returns a cobra.Command.RunE function that gets the client.Context and FlagSet,
// then uses the provided maker to make the Msg that it then generates or broadcasts as a Tx.
//
// R is the type of the Msg.
func genericTxRunE[R sdk.Msg](maker msgMaker[R]) func(cmd *cobra.Command, args []string) error {
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

// govTxRunE returns a cobra.Command.RunE function that gets the client.Context and FlagSet,
// then uses the provided maker to make the Msg. The Msg is then put into a governance
// proposal and either generated or broadcast as a Tx.
//
// R is the type of the Msg.
func govTxRunE[R sdk.Msg](maker msgMaker[R]) func(cmd *cobra.Command, args []string) error {
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
		// return govcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg) // TODO[1760]: gov-cli
		_ = msg
		return fmt.Errorf("not yet updated")
	}
}

// queryReqMaker is a function that creates a query request message.
//
// R is the type of request message.
type queryReqMaker[R any] func(clientCtx client.Context, flagSet *pflag.FlagSet, args []string) (*R, error)

// queryEndpoint is a grpc_query endpoint function.
//
// R is the type of request message.
// S is the type of response message.
type queryEndpoint[R any, S proto.Message] func(queryClient exchange.QueryClient, ctx context.Context, req *R, opts ...grpc.CallOption) (S, error)

// genericQueryRunE returns a cobra.Command.RunE function that gets the query context and FlagSet,
// then uses the provided maker to make the query request message. A query client is created and
// that message is then given to the provided endpoint func to get the response which is then printed.
//
// R is the type of request message.
// S is the type of response message.
func genericQueryRunE[R any, S proto.Message](reqMaker queryReqMaker[R], endpoint queryEndpoint[R, S]) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		req, err := reqMaker(clientCtx, cmd.Flags(), args)
		if err != nil {
			return err
		}

		cmd.SilenceUsage = true
		queryClient := exchange.NewQueryClient(clientCtx)
		res, err := endpoint(queryClient, cmd.Context(), req)
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(res)
	}
}

// AddUseArgs adds the given strings to the cmd's Use, separated by a space.
func AddUseArgs(cmd *cobra.Command, args ...string) {
	cmd.Use = cmd.Use + " " + strings.Join(args, " ")
}

// AddUseDetails appends each provided section to the Use field with an empty line between them.
func AddUseDetails(cmd *cobra.Command, sections ...string) {
	if len(sections) > 0 {
		cmd.Use = cmd.Use + "\n\n" + strings.Join(sections, "\n\n")
	}
	cmd.DisableFlagsInUseLine = true
}

// AddQueryExample appends an example to a query command's examples.
func AddQueryExample(cmd *cobra.Command, args ...string) {
	if len(cmd.Example) > 0 {
		cmd.Example += "\n"
	}
	cmd.Example += fmt.Sprintf("%s query %s %s", version.AppName, exchange.ModuleName, cmd.Name())
	if len(args) > 0 {
		cmd.Example += " " + strings.Join(args, " ")
	}
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
	return fmt.Sprintf(`If --%[1]s <%[1]s> is provided, that is used as the <%[1]s>.
If no --%[1]s is provided, the --%[2]s account address is used as the <%[1]s>.
A <%[1]s> is required.`,
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

// OptFlagUse wraps a ReqFlagUse in [], e.g. "[--name <opt>]".
func OptFlagUse(name string, opt string) string {
	return "[" + ReqFlagUse(name, opt) + "]"
}

// ProposalFileDesc is a description of the --proposal flag and expected file.
func ProposalFileDesc(msgType sdk.Msg) string {
	return fmt.Sprintf(`The file provided with the --%[1]s flag should be a json-encoded Tx.
The Tx should have a message with a %[2]s that contains a %[3]s.
Such a message can be generated using the --generate-only flag on the tx endpoint.

Example (with just the important bits):
{
  "body": {
    "messages": [
      {
        "@type": "%[2]s",
        "messages": [
          {
            "@type": "%[3]s",
            "authority": "...",
            <other %[4]T fields>
          }
        ],
      }
    ],
  },
}

If other message flags are provided with --%[1]s, they will overwrite just that field.
`,
		FlagProposal, sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}), sdk.MsgTypeURL(msgType), msgType,
	)
}

var (
	// UseFlagsBreak is a string to use to start a new line of flags in the Use string of a command.
	UseFlagsBreak = "\n     "

	// RepeatableDesc is a description of how repeatable flags/values can be provided.
	RepeatableDesc = "If a flag is repeatable, multiple entries can be separated by commas\nand/or the flag can be provided multiple times."

	// ReqAdminUse is the Use string of the --admin flag.
	ReqAdminUse = fmt.Sprintf("{--%s|--%s} <admin>", flags.FlagFrom, FlagAdmin)

	// ReqAdminDesc is a description of how the --admin, --authority, and --from flags work and are sort of required.
	ReqAdminDesc = fmt.Sprintf(`If --%[1]s <admin> is provided, that is used as the <admin>.
If no --%[1]s is provided, but the --%[2]s flag was, the governance module account is used as the <admin>.
Otherwise the --%[3]s account address is used as the <admin>.
An <admin> is required.`,
		FlagAdmin, FlagAuthority, flags.FlagFrom,
	)

	// ReqEnableDisableUse is a use string for the --enable and --disable flags.
	ReqEnableDisableUse = fmt.Sprintf("{--%s|--%s}", FlagEnable, FlagDisable)

	// ReqEnableDisableDesc is a description of the --enable and --disable flags.
	ReqEnableDisableDesc = fmt.Sprintf("One of --%s or --%s must be provided, but not both.", FlagEnable, FlagDisable)

	// AccessGrantsDesc is a description of the <asset grant> format.
	AccessGrantsDesc = fmt.Sprintf(`An <access grant> has the format "<address>:<permissions>"
In <permissions>, separate each permission with a + (plus) or . (period).
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

	// AuthorityDesc is a description of the authority flag.
	AuthorityDesc = fmt.Sprintf("If --%s <authority> is not provided, the governance module account is used as the <authority>.", FlagAuthority)

	// ReqAskBidUse is a use string of the --ask and --bid flags when one is required.
	ReqAskBidUse = fmt.Sprintf("{--%s|--%s}", FlagAsk, FlagBid)

	// ReqAskBidDesc is a description of the --ask and --bid flags when one is required.
	ReqAskBidDesc = fmt.Sprintf("One of --%s or --%s must be provided, but not both.", FlagAsk, FlagBid)

	// OptAsksBidsUse is a use string of the optional mutually exclusive --asks and --bids flags.
	OptAsksBidsUse = fmt.Sprintf("[--%s|--%s]", FlagAsks, FlagBids)

	// OptAsksBidsDesc is a description of the --asks and --bids flags when they're optional.
	OptAsksBidsDesc = fmt.Sprintf("At most one of --%s or --%s can be provided.", FlagAsks, FlagBids)
)
