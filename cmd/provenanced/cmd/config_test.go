package cmd_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/provenance-io/provenance/cmd/provenanced/cmd"
	"github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
)

func TestClientConfigList(t *testing.T) {
	command := cmd.ClientConfigCmd()
	home := t.TempDir()
	logger := log.NewNopLogger()
	cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
	require.NoError(t, err)

	appCodec := simapp.MakeTestEncodingConfig().Marshaler
	err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home).WithViper("")
	clientCtx, err = config.ReadFromClientConfig(clientCtx)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	command.SetArgs([]string{})
	b := bytes.NewBufferString("")
	command.SetOut(b)
	err = command.ExecuteContext(ctx)
	require.NoErrorf(t, err, "%s - unexpected error", command.Name())
	out, err := ioutil.ReadAll(b)
	require.NoErrorf(t, err, "%s - unexpected error", command.Name())
	outStr := strings.Trim(string(out), "\n")
	require.Equal(t, "{\n\t\"chain-id\": \"\",\n\t\"keyring-backend\": \"test\",\n\t\"output\": \"text\",\n\t\"node\": \"tcp://localhost:26657\",\n\t\"broadcast-mode\": \"block\"\n}", outStr)
}

func TestClientConfigCmdSet(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		initial string
		updated string
	}{
		{
			name:    "set chain-id config",
			args:    []string{"chain-id"},
			initial: "",
			updated: "new-chain",
		},
		{
			name:    "set node config",
			args:    []string{"node"},
			initial: "tcp://localhost:26657",
			updated: "tcp://127.0.0.1:26657",
		},
		{
			name:    "set output config",
			args:    []string{"output"},
			initial: "text",
			updated: "json",
		},
		{
			name:    "set broadcast-mode config",
			args:    []string{"broadcast-mode"},
			initial: "block",
			updated: "sync",
		},
		{
			name:    "set keyring-backend config",
			args:    []string{"keyring-backend"},
			initial: "test",
			updated: "os",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			command := cmd.ClientConfigCmd()
			home := t.TempDir()
			logger := log.NewNopLogger()
			cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
			require.NoError(t, err)

			appCodec := simapp.MakeTestEncodingConfig().Marshaler
			err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
			require.NoError(t, err)

			serverCtx := server.NewContext(viper.New(), cfg, logger)
			clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home).WithViper("")
			clientCtx, err = config.ReadFromClientConfig(clientCtx)
			require.NoError(t, err)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			// Get the initial value
			command.SetArgs(tc.args)
			b := bytes.NewBufferString("")
			command.SetOut(b)
			err = command.ExecuteContext(ctx)
			require.NoErrorf(t, err, "%s - unexpected error", command.Name())
			out, err := ioutil.ReadAll(b)
			require.NoErrorf(t, err, "%s - unexpected error", command.Name())
			outStr := strings.Trim(string(out), "\n")
			require.Equal(t, tc.initial, outStr)

			// Set the updated value
			command.SetArgs(append(tc.args, tc.updated))
			err = command.ExecuteContext(ctx)
			require.NoErrorf(t, err, "%s - unexpected error", command.Name())

			// Get the update value
			command.SetArgs(tc.args)
			b = bytes.NewBufferString("")
			command.SetOut(b)
			err = command.ExecuteContext(ctx)
			require.NoErrorf(t, err, "%s - unexpected error", command.Name())
			out, err = ioutil.ReadAll(b)
			require.NoErrorf(t, err, "%s - unexpected error", command.Name())
			outStr = strings.Trim(string(out), "\n")
			require.Equal(t, tc.updated, outStr)
		})
	}
}
