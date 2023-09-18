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
		extensions   []string
	}{
		{
			name:         "failure - no flags specified",
			target:       "tmp",
			createTarget: true,
			err:          "at least one doc type must be specified",
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
			extensions:   []string{".yaml"},
		},
		{
			name:         "success - rest is generated",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--rest"},
			extensions:   []string{".rst"},
		},
		{
			name:         "success - manpage is generated",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--manpage"},
			extensions:   []string{".1"},
		},
		{
			name:         "success - markdown is generated",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--markdown"},
			extensions:   []string{".md"},
		},
		{
			name:         "success - multiple types supported",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--markdown", "--yaml"},
			extensions:   []string{".md", ".yaml"},
		},
		{
			name:         "success - generates a new directory",
			target:       "tmp2",
			createTarget: false,
			flags:        []string{"--yaml"},
			extensions:   []string{".md", ".yaml"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()

			targetPath := filepath.Join(home, tc.target)
			if tc.createTarget {
				require.NoError(t, os.Mkdir(targetPath, 0755), "Mkdir successfully created directory")
			}

			logger := log.NewNopLogger()
			cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
			require.NoError(t, err, "Created default tendermint config")

			appCodec := sdksim.MakeTestEncodingConfig().Codec
			err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
			require.NoError(t, err, "Executed init command")

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
				require.Error(t, err, "should throw an error")
				require.Equal(t, tc.err, err.Error(), "should return the correct error")
				files, err := os.ReadDir(targetPath)
				require.NoError(t, err, "ReadDir should not return an error")
				require.Equal(t, 0, len(files), "should not generate files when failed")
			} else {
				err := cmd.ExecuteContext(ctx)
				require.NoError(t, err, "should not return an error")

				files, err := os.ReadDir(targetPath)
				require.NoError(t, err, "ReadDir should not return an error")
				require.NotZero(t, len(files), "should generate files when successful")

				for _, file := range files {
					ext := filepath.Ext(file.Name())

					contains := false
					for _, extension := range tc.extensions {
						contains = contains || ext == extension
					}
					require.True(t, contains, "should generate files with correct extension")
				}
			}

			require.NoError(t, os.RemoveAll(targetPath), "RemoveAll should be able to remove the temporary target directory")
		})
	}
}
