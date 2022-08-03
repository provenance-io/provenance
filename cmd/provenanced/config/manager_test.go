package config

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/simapp"

	tmconfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
)

type ConfigManagerTestSuite struct {
	suite.Suite

	Home string
}

func TestConfigManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigManagerTestSuite))
}

func (s *ConfigManagerTestSuite) SetupTest() {
	s.Home = s.T().TempDir()
	s.T().Logf("%s Home: %s", s.T().Name(), s.Home)
}

// makeDummyCmd creates a dummy command with a context in it that can be used to test all the manager stuff.
func (s *ConfigManagerTestSuite) makeDummyCmd() *cobra.Command {
	encodingConfig := simapp.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithHomeDir(s.Home)
	clientCtx.Viper = viper.New()
	serverCtx := server.NewContext(clientCtx.Viper, tmconfig.DefaultConfig(), log.NewNopLogger())

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	dummyCmd := &cobra.Command{
		Use:   "dummy",
		Short: "Just used for testing. Doesn't do anything.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	dummyCmd.SetOut(ioutil.Discard)
	dummyCmd.SetErr(ioutil.Discard)
	dummyCmd.SetArgs([]string{})
	var err error
	dummyCmd, err = dummyCmd.ExecuteContextC(ctx)
	s.Require().NoError(err, "dummy command execution")
	return dummyCmd
}

func (s *ConfigManagerTestSuite) TestConfigIndexEventsWriteReadCanary() {
	// This test will pass as long as a certain bug still exists in the Cosmos WriteConfigFile function.
	// Issue: https://github.com/cosmos/cosmos-sdk/issues/10016
	// If you're looking at it because it is failing, we might need to remove some work-around code.
	// Use the TestConfigIndexEventsWriteRead test to figure out if it's actually fixed.
	// If it's actually fixed, you can delete the appConfigIndexEventsWorkAround func, the call to it, and this test.

	confFile := filepath.Join(s.Home, "app.toml")
	appConfig := serverconfig.DefaultConfig()
	appConfig.IndexEvents = []string{"key1", "key2"}
	serverconfig.WriteConfigFile(confFile, appConfig)

	// Read that file into viper.
	vpr := viper.New()
	vpr.SetConfigFile(confFile)
	err := vpr.ReadInConfig()
	s.Require().EqualError(err, "While parsing config: toml: incomplete number", "reading config file into viper")
}

func (s *ConfigManagerTestSuite) TestConfigIndexEventsWriteRead() {
	// This return is here so that the test can stay here, uncommented, but not actually run on its own.
	// To actually run the test, delete it. If it still passes, WOOO! Keep this test without the premature return.
	return

	// Create config with two IndexEvents entries, and write it to a file.
	confFile := filepath.Join(s.Home, "app.toml")
	appConfig := serverconfig.DefaultConfig()
	appConfig.IndexEvents = []string{"key1", "key2"}
	serverconfig.WriteConfigFile(confFile, appConfig)

	// Read that file into viper.
	vpr := viper.New()
	vpr.SetConfigFile(confFile)
	err := vpr.ReadInConfig()
	s.Require().NoError(err, "reading config file into viper")
	vprIndexEvents := vpr.GetStringSlice("index-events")
	s.Require().Equal(appConfig.IndexEvents, vprIndexEvents, "viper's index events")
}

func (s *ConfigManagerTestSuite) TestManagerWriteAppConfigWithIndexEventsThenReadIt() {
	// This test is just making sure that writing/reading index events works in our stuff.
	dCmd := s.makeDummyCmd()

	appConfig := serverconfig.DefaultConfig()
	appConfig.IndexEvents = []string{"key1", "key2"}
	SaveConfigs(dCmd, appConfig, nil, nil, false)

	err := LoadConfigFromFiles(dCmd)
	s.Require().NoError(err, "loading config from files")

	appConfig2, err2 := ExtractAppConfig(dCmd)
	s.Require().NoError(err2, "extracging app config")
	s.Require().Equal(appConfig.IndexEvents, appConfig2.IndexEvents, "index events before/after")
}

func (s *ConfigManagerTestSuite) TestPackedConfigCosmosLoadDefaults() {
	dCmd := s.makeDummyCmd()

	appConfig := serverconfig.DefaultConfig()
	tmConfig := tmconfig.DefaultConfig()
	clientConfig := DefaultClientConfig()
	generateAndWritePackedConfig(dCmd, appConfig, tmConfig, clientConfig, false)
	s.Require().NoError(loadPackedConfig(dCmd))

	ctx := client.GetClientContextFromCmd(dCmd)
	vpr := ctx.Viper
	s.Require().NotPanics(func() {
		appConfig2 := serverconfig.GetConfig(vpr)
		s.Assert().Equal(*appConfig, appConfig2)
	})
}

func (s *ConfigManagerTestSuite) TestPackedConfigCosmosLoadGlobalLabels() {
	dCmd := s.makeDummyCmd()

	appConfig := serverconfig.DefaultConfig()
	appConfig.Telemetry.GlobalLabels = append(appConfig.Telemetry.GlobalLabels, []string{"key1", "value1"})
	appConfig.Telemetry.GlobalLabels = append(appConfig.Telemetry.GlobalLabels, []string{"key2", "value2"})
	tmConfig := tmconfig.DefaultConfig()
	clientConfig := DefaultClientConfig()
	generateAndWritePackedConfig(dCmd, appConfig, tmConfig, clientConfig, false)
	s.Require().NoError(loadPackedConfig(dCmd))

	ctx := client.GetClientContextFromCmd(dCmd)
	vpr := ctx.Viper
	s.Require().NotPanics(func() {
		appConfig2 := serverconfig.GetConfig(vpr)
		s.Assert().Equal(appConfig.Telemetry.GlobalLabels, appConfig2.Telemetry.GlobalLabels)
	})
}

func (s *ConfigManagerTestSuite) TestUnmanagedConfig() {
	s.T().Run("unmanaged config is read with no other config files", func(t *testing.T) {
		dCmd := s.makeDummyCmd()
		configDir := GetFullPathToConfigDir(dCmd)
		uFile := GetFullPathToUnmanagedConf(dCmd)
		require.NoError(t, os.MkdirAll(configDir, 0o755), "making config dir")
		require.NoError(t, os.WriteFile(uFile, []byte("banana = \"bananas\"\n"), 0o644), "writing unmanaged config")
		require.NoError(t, LoadConfigFromFiles(dCmd))
		ctx := client.GetClientContextFromCmd(dCmd)
		vpr := ctx.Viper
		actual := vpr.GetString("banana")
		assert.Equal(t, "bananas", actual, "unmanaged field value")
	})

	s.T().Run("unmanaged config entry does not override other config", func(t *testing.T) {
		dCmd := s.makeDummyCmd()
		configDir := GetFullPathToConfigDir(dCmd)
		uFile := GetFullPathToUnmanagedConf(dCmd)
		require.NoError(t, os.MkdirAll(configDir, 0o755), "making config dir")
		require.NoError(t, os.WriteFile(uFile, []byte("db_backend = \"still bananas\"\n"), 0o644), "writing unmanaged config")
		require.NoError(t, LoadConfigFromFiles(dCmd))
		ctx := client.GetClientContextFromCmd(dCmd)
		vpr := ctx.Viper
		actual := vpr.GetString("db_backend")
		assert.NotEqual(t, "still bananas", actual, "unmanaged field value")
		assert.Equal(t, tmconfig.DefaultConfig().DBBackend, actual, "unmanaged field default value")
	})

	s.T().Run("unmanaged config is read with unpacked files", func(t *testing.T) {
		dCmd := s.makeDummyCmd()
		uFile := GetFullPathToUnmanagedConf(dCmd)
		SaveConfigs(dCmd, serverconfig.DefaultConfig(), tmconfig.DefaultConfig(), DefaultClientConfig(), false)
		require.NoError(t, os.WriteFile(uFile, []byte("my-custom-entry = \"stuff\"\n"), 0o644), "writing unmanaged config")
		require.NoError(t, LoadConfigFromFiles(dCmd))
		ctx := client.GetClientContextFromCmd(dCmd)
		vpr := ctx.Viper
		actual := vpr.GetString("my-custom-entry")
		assert.Equal(t, "stuff", actual, "unmanaged field value")
	})

	s.T().Run("unmanaged config is read with packed config", func(t *testing.T) {
		dCmd := s.makeDummyCmd()
		uFile := GetFullPathToUnmanagedConf(dCmd)
		SaveConfigs(dCmd, serverconfig.DefaultConfig(), tmconfig.DefaultConfig(), DefaultClientConfig(), false)
		require.NoError(t, PackConfig(dCmd), "packing config")
		require.NoError(t, os.WriteFile(uFile, []byte("other-custom-entry = 8\n"), 0o644), "writing unmanaged config")
		require.NoError(t, LoadConfigFromFiles(dCmd))
		ctx := client.GetClientContextFromCmd(dCmd)
		vpr := ctx.Viper
		actual := vpr.GetInt("other-custom-entry")
		assert.Equal(t, 8, actual, "unmanaged field value")
	})
}
