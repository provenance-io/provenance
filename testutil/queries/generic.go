package queries

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/gogoproto/proto"
)

// AssertGetRequest does an HTTP get on the provided url and unmarshalls the response into the provided emptyResp.
// The returned bool will be true on success, or false if something goes wrong.
//
// The url should start with a / and should just contain the portion as defined in the proto.
func AssertGetRequest[T proto.Message](t *testing.T, n *network.Network, url string, emptyResp T) (T, bool) {
	t.Helper()
	if !assert.NotEmpty(t, n.Validators, "Network.Validators") {
		return emptyResp, false
	}
	val := n.Validators[0]

	url = val.APIAddress + url
	respBz, err := testutil.GetRequestWithHeaders(url, nil)
	if !assert.NoError(t, err, "failed to execute GET %q", url) {
		return emptyResp, false
	}
	t.Logf("GET %q\nResponse: %s", url, string(respBz))

	err = val.ClientCtx.Codec.UnmarshalJSON(respBz, emptyResp)
	if !assert.NoError(t, err, "failed to unmarshal response as %T", emptyResp) {
		return emptyResp, false
	}

	return emptyResp, true
}

// createQueryCmd creates a command that will execute a query on the provided url,
// unmarshal the response into the empty response, and output the result as either
// yaml or json depending on the --output flag.
func createQueryCmd[T proto.Message](n *network.Network, cmdName, url string, emptyResp T) *cobra.Command {
	if n == nil || len(n.Validators) == 0 {
		panic("network must have at least one validator")
	}
	url = n.Validators[0].APIAddress + url

	cmd := &cobra.Command{
		Use:          "generic-" + cmdName,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			respBz, err := testutil.GetRequestWithHeaders(url, nil)
			if err != nil {
				return fmt.Errorf("failed to execute GET %q: %w", url, err)
			}

			err = clientCtx.Codec.UnmarshalJSON(respBz, emptyResp)
			if err != nil {
				_ = clientCtx.PrintString("Response from GET " + url + "\n")
				_ = clientCtx.PrintBytes(respBz)
				return fmt.Errorf("failed to unmarshal response as %T: %w", emptyResp, err)
			}

			return clientCtx.PrintProto(emptyResp)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
