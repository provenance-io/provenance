package cmd_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	provenancecmd "github.com/provenance-io/provenance/cmd/provenanced/cmd"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
)

func TestDocGen(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		createTarget bool
		flags        []string
		err          string
	}{
		{
			name:         "failure - no flags specified",
			target:       "tmp",
			createTarget: true,
			err:          "at least one doc type must be specified.",
		},
		{
			name:         "failure - unsupported flag format",
			target:       "tmp",
			flags:        []string{"--bad"},
			createTarget: true,
			err:          "unknown flag: --bad",
		},
		{
			name:         "success - yaml is generated",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--yaml"},
		},
		{
			name:         "success - rest is generated",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--rest"},
		},
		{
			name:         "success - manpage is generated",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--manpage"},
		},
		{
			name:         "success - markdown is generated",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--markdown"},
		},
		{
			name:         "success - multiple types supported",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--markdown", "--yaml"},
		},
		{
			name:         "success - generates a new directory",
			target:       "tmp2",
			createTarget: false,
			flags:        []string{"--yaml"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()

			targetPath := filepath.Join(home, tc.target)
			if tc.createTarget {
				require.NoError(t, os.Mkdir(targetPath, 0755))
			}

			logger := log.NewNopLogger()
			cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
			require.NoError(t, err)

			appCodec := sdksim.MakeTestEncodingConfig().Codec
			err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
			require.NoError(t, err)

			serverCtx := server.NewContext(viper.New(), cfg, logger)
			clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			cmd := provenancecmd.GetDocGenCmd()
			args := append([]string{targetPath}, tc.flags...)
			cmd.SetArgs(args)

			if len(tc.err) > 0 {
				err := cmd.ExecuteContext(ctx)
				require.Error(t, err)
				require.Equal(t, tc.err, err.Error())
			} else {
				err := cmd.ExecuteContext(ctx)
				require.NoError(t, err)
				// We probably want to check generation of files
			}

			require.NoError(t, os.RemoveAll(targetPath))
		})
	}
}
