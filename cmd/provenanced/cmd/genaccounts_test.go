package cmd_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"

	"github.com/provenance-io/provenance/app"
	provenancecmd "github.com/provenance-io/provenance/cmd/provenanced/cmd"
)

var testMbm = module.NewBasicManager(genutil.AppModuleBasic{})

func TestAddGenesisAccountCmd(t *testing.T) {
	_, _, addr1 := testdata.KeyTestPubAddr()
	tests := []struct {
		name      string
		addr      string
		denom     string
		expectErr bool
	}{
		{
			name:      "invalid address",
			addr:      "",
			denom:     "1000atom",
			expectErr: true,
		},
		{
			name:      "valid address",
			addr:      addr1.String(),
			denom:     "1000atom",
			expectErr: false,
		},
		{
			name:      "multiple denoms",
			addr:      addr1.String(),
			denom:     "1000atom, 2000stake",
			expectErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			logger := log.NewNopLogger()
			cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
			require.NoError(t, err)

			appCodec := simapp.MakeTestEncodingConfig().Marshaler
			err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
			require.NoError(t, err)

			serverCtx := server.NewContext(viper.New(), cfg, logger)
			clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			cmd := provenancecmd.AddGenesisAccountCmd(home)
			cmd.SetArgs([]string{
				tc.addr,
				tc.denom,
				fmt.Sprintf("--%s=home", flags.FlagHome)})

			if tc.expectErr {
				require.Error(t, cmd.ExecuteContext(ctx))
			} else {
				require.NoError(t, cmd.ExecuteContext(ctx))
			}
		})
	}
}

func TestAddGenesisMsgFeeCmd(t *testing.T) {
	tests := []struct {
		name         string
		msgType      string
		fee          string
		coinDenom    string
		expectErrMsg string
	}{
		{
			name:         "invalid msg type",
			msgType:      "InvalidMsgType",
			fee:          "1000jackthecat",
			coinDenom:    "vspn",
			expectErrMsg: "unable to resolve type URL /InvalidMsgType",
		},
		{
			name:         "invalid fee",
			msgType:      "/provenance.name.v1.MsgBindNameRequest",
			fee:          "not-a-fee",
			coinDenom:    "vspn",
			expectErrMsg: "failed to parse coin: invalid decimal coin expression: not-a-fee",
		},
		{
			name:         "valid msg type and fee",
			msgType:      "/provenance.name.v1.MsgBindNameRequest",
			fee:          "1000jackthecat",
			coinDenom:    "vspn",
			expectErrMsg: "",
		},
		{
			name:         "invalid fee",
			msgType:      "provenance.name.v1.MsgBindNameRequest",
			fee:          "1000jackthecat",
			coinDenom:    "vspn",
			expectErrMsg: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			logger := log.NewNopLogger()
			cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
			require.NoError(t, err)

			appCodec := simapp.MakeTestEncodingConfig().Marshaler
			err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
			require.NoError(t, err)

			serverCtx := server.NewContext(viper.New(), cfg, logger)
			clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			cmd := provenancecmd.AddGenesisCustomFloorPriceDenom(home, app.MakeEncodingConfig().InterfaceRegistry)
			cmd.SetArgs([]string{
				tc.msgType,
				tc.fee,
				tc.coinDenom,
				fmt.Sprintf("--%s=home", flags.FlagHome)})

			if len(tc.expectErrMsg) > 0 {
				err := cmd.ExecuteContext(ctx)
				require.Error(t, err)
				require.Equal(t, tc.expectErrMsg, err.Error())
			} else {
				require.NoError(t, cmd.ExecuteContext(ctx))
			}
		})
	}
}
