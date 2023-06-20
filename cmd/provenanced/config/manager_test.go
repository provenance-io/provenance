package config

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	tmcmds "github.com/tendermint/tendermint/cmd/cometbft/commands"
	tmconfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"

	"github.com/provenance-io/provenance/internal/pioconfig"
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
	encodingConfig := sdksim.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithHomeDir(s.Home)
	clientCtx.Viper = viper.New()
	serverCtx := server.NewContext(clientCtx.Viper, DefaultTmConfig(), log.NewNopLogger())

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

func (s *ConfigManagerTestSuite) logFile(path string) {
	if !FileExists(path) {
		s.T().Logf("File does not exist: %s", path)
		return
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		s.T().Logf("Error reading %s: %v", path, err)
		return
	}
	s.T().Logf("File: %s\nContents:\n%s", path, contents)
}

func (s *ConfigManagerTestSuite) TestConfigIndexEventsWriteRead() {
	// The IndexEvents field has some special handling that was broken at one point.
	// This test exists to make sure it doesn't break again.

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

	appConfig := DefaultAppConfig()
	tmConfig := DefaultTmConfig()
	clientConfig := DefaultClientConfig()
	generateAndWritePackedConfig(dCmd, appConfig, tmConfig, clientConfig, false)
	s.Require().NoError(loadPackedConfig(dCmd))

	ctx := client.GetClientContextFromCmd(dCmd)
	vpr := ctx.Viper
	s.Require().NotPanics(func() {
		appConfig2, err := serverconfig.GetConfig(vpr)
		s.Require().NoError(err, "GetConfig")
		s.Assert().Equal(*appConfig, appConfig2)
	})
}

func (s *ConfigManagerTestSuite) TestPackedConfigCosmosLoadGlobalLabels() {
	dCmd := s.makeDummyCmd()

	appConfig := serverconfig.DefaultConfig()
	appConfig.Telemetry.GlobalLabels = append(appConfig.Telemetry.GlobalLabels, []string{"key1", "value1"})
	appConfig.Telemetry.GlobalLabels = append(appConfig.Telemetry.GlobalLabels, []string{"key2", "value2"})
	tmConfig := DefaultTmConfig()
	clientConfig := DefaultClientConfig()
	generateAndWritePackedConfig(dCmd, appConfig, tmConfig, clientConfig, false)
	s.Require().NoError(loadPackedConfig(dCmd))

	ctx := client.GetClientContextFromCmd(dCmd)
	vpr := ctx.Viper
	s.Require().NotPanics(func() {
		appConfig2, err := serverconfig.GetConfig(vpr)
		s.Require().NoError(err, "GetConfig")
		s.Assert().Equal(appConfig.Telemetry.GlobalLabels, appConfig2.Telemetry.GlobalLabels)
	}, "GetConfig")
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
		assert.Equal(t, DefaultTmConfig().DBBackend, actual, "unmanaged field default value")
	})

	s.T().Run("unmanaged config is read with unpacked files", func(t *testing.T) {
		dCmd := s.makeDummyCmd()
		uFile := GetFullPathToUnmanagedConf(dCmd)
		SaveConfigs(dCmd, DefaultAppConfig(), DefaultTmConfig(), DefaultClientConfig(), false)
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
		SaveConfigs(dCmd, DefaultAppConfig(), DefaultTmConfig(), DefaultClientConfig(), false)
		require.NoError(t, PackConfig(dCmd), "packing config")
		require.NoError(t, os.WriteFile(uFile, []byte("other-custom-entry = 8\n"), 0o644), "writing unmanaged config")
		require.NoError(t, LoadConfigFromFiles(dCmd))
		ctx := client.GetClientContextFromCmd(dCmd)
		vpr := ctx.Viper
		actual := vpr.GetInt("other-custom-entry")
		assert.Equal(t, 8, actual, "unmanaged field value")
	})
}

func (s *ConfigManagerTestSuite) TestServerGetConfigGlobalLabels() {
	// This test exists because the telemetry.global-labels field used to be handled specially in
	// serverconfig.GetConfig, so we had to have some work-around special handling for that field
	// in the reflector. Now it exists to make sure it doesn't break again.
	globalLabels := [][]string{
		{"keya", "valuea"},
		{"keyb", "valueb"},
		{"keyc", "valuec"},
	}
	telemetry := map[string]interface{}{
		"global-labels": globalLabels,
	}
	cfgMap := map[string]interface{}{
		"telemetry": telemetry,
	}

	vpr := viper.New()
	s.Require().NoError(vpr.MergeConfigMap(cfgMap), "MergeConfigMap")
	cfg, err := serverconfig.GetConfig(vpr)
	s.Require().NoError(err, "GetConfig")
	actual := cfg.Telemetry.GlobalLabels
	s.Assert().Equal(globalLabels, actual, "cfg.Telemetry.GlobalLabels")
}

func (s *ConfigManagerTestSuite) TestConfigMinGasPrices() {
	configDir := GetFullPathToConfigDir(s.makeDummyCmd())
	s.Require().NoError(os.MkdirAll(configDir, 0o755), "making config dir")

	pioconfig.SetProvenanceConfig("manager", 42)
	defaultMinGasPrices := pioconfig.GetProvenanceConfig().ProvenanceMinGasPrices
	s.Require().NotEqual("", defaultMinGasPrices, "ProvenanceMinGasPrices")

	s.Run("DefaultAppConfig has MinGasPrices", func() {
		cfg := DefaultAppConfig()
		actual := cfg.MinGasPrices
		s.Assert().Equal(defaultMinGasPrices, actual, "MinGasPrices")
	})

	s.Run("no files", func() {
		cmd := s.makeDummyCmd()
		s.Require().NoError(LoadConfigFromFiles(cmd), "LoadConfigFromFiles")
		cfg, err := ExtractAppConfig(cmd)
		s.Require().NoError(err, "ExtractAppConfig")
		actual := cfg.MinGasPrices
		s.Assert().Equal(defaultMinGasPrices, actual, "MinGasPrices")
	})

	s.Run("tm and client files but no app file", func() {
		cmd1 := s.makeDummyCmd()
		SaveConfigs(cmd1, nil, DefaultTmConfig(), DefaultClientConfig(), false)
		appCfgFile := GetFullPathToAppConf(cmd1)
		_, err := os.Stat(appCfgFile)
		fileExists := !os.IsNotExist(err)
		s.Require().False(fileExists, "file exists: %s", appCfgFile)

		cmd2 := s.makeDummyCmd()
		s.Require().NoError(LoadConfigFromFiles(cmd2), "LoadConfigFromFiles")
		cfg, err := ExtractAppConfig(cmd2)
		s.Require().NoError(err, "ExtractAppConfig")
		actual := cfg.MinGasPrices
		s.Assert().Equal(defaultMinGasPrices, actual, "MinGasPrices")
	})

	s.Run("all files exist min gas price empty", func() {
		cmd1 := s.makeDummyCmd()
		appCfg := DefaultAppConfig()
		appCfg.MinGasPrices = ""
		SaveConfigs(cmd1, appCfg, DefaultTmConfig(), DefaultClientConfig(), false)
		appCfgFile := GetFullPathToAppConf(cmd1)
		_, err := os.Stat(appCfgFile)
		fileExists := !os.IsNotExist(err)
		s.Require().True(fileExists, "file exists: %s", appCfgFile)

		cmd2 := s.makeDummyCmd()
		s.Require().NoError(LoadConfigFromFiles(cmd2), "LoadConfigFromFiles")
		cfg, err := ExtractAppConfig(cmd2)
		s.Require().NoError(err, "ExtractAppConfig")
		actual := cfg.MinGasPrices
		s.Assert().Equal("", actual, "MinGasPrices")
	})

	s.Run("all files exist min gas price something else", func() {
		cmd1 := s.makeDummyCmd()
		appCfg := DefaultAppConfig()
		appCfg.MinGasPrices = "something else"
		SaveConfigs(cmd1, appCfg, DefaultTmConfig(), DefaultClientConfig(), false)
		appCfgFile := GetFullPathToAppConf(cmd1)
		_, err := os.Stat(appCfgFile)
		fileExists := !os.IsNotExist(err)
		s.Require().True(fileExists, "file exists: %s", appCfgFile)

		cmd2 := s.makeDummyCmd()
		s.Require().NoError(LoadConfigFromFiles(cmd2), "LoadConfigFromFiles")
		cfg, err := ExtractAppConfig(cmd2)
		s.Require().NoError(err, "ExtractAppConfig")
		actual := cfg.MinGasPrices
		s.Assert().Equal("something else", actual, "MinGasPrices")
	})

	s.Run("packed config without min-gas-prices", func() {
		cmd1 := s.makeDummyCmd()
		SaveConfigs(cmd1, DefaultAppConfig(), DefaultTmConfig(), DefaultClientConfig(), false)
		s.Require().NoError(PackConfig(cmd1), "PackConfig")
		packedCfgFile := GetFullPathToPackedConf(cmd1)
		_, err := os.Stat(packedCfgFile)
		fileExists := !os.IsNotExist(err)
		s.Require().True(fileExists, "file exists: %s", packedCfgFile)

		// Just to be sure, rewrite the file as just "{}".
		s.Require().NoError(os.WriteFile(packedCfgFile, []byte("{}"), 0o644), "writing packed config")

		cmd2 := s.makeDummyCmd()
		s.Require().NoError(LoadConfigFromFiles(cmd2), "LoadConfigFromFiles")
		cfg, err := ExtractAppConfig(cmd2)
		s.Require().NoError(err, "ExtractAppConfig")
		actual := cfg.MinGasPrices
		s.Assert().Equal(defaultMinGasPrices, actual)
	})

	s.Run("packed config with min-gas-prices", func() {
		cmd1 := s.makeDummyCmd()
		SaveConfigs(cmd1, DefaultAppConfig(), DefaultTmConfig(), DefaultClientConfig(), false)
		s.Require().NoError(PackConfig(cmd1), "PackConfig")
		packedCfgFile := GetFullPathToPackedConf(cmd1)
		_, err := os.Stat(packedCfgFile)
		fileExists := !os.IsNotExist(err)
		s.Require().True(fileExists, "file exists: %s", packedCfgFile)

		// rewrite the packed file to include min-gas-prices
		s.Require().NoError(os.WriteFile(packedCfgFile, []byte(`{"minimum-gas-prices":"65blue"}`), 0o644), "writing packed config")

		cmd2 := s.makeDummyCmd()
		s.Require().NoError(LoadConfigFromFiles(cmd2), "LoadConfigFromFiles")
		cfg, err := ExtractAppConfig(cmd2)
		s.Require().NoError(err, "ExtractAppConfig")
		actual := cfg.MinGasPrices
		s.Assert().Equal("65blue", actual)
	})
}

func (s *ConfigManagerTestSuite) TestDefaultTmConfig() {
	cfg := DefaultTmConfig()

	s.Run("consensus.commit_timeout", func() {
		exp := 5 * time.Second
		act := cfg.Consensus.TimeoutCommit
		s.Assert().Equal(exp, act, "cfg.Consensus.TimeoutCommit")
	})
}

func (s *ConfigManagerTestSuite) TestPackedConfigTmLoadDefaults() {
	dCmd := s.makeDummyCmd()
	dCmd.Flags().String("home", s.Home, "home dir")

	appConfig := DefaultAppConfig()
	tmConfig := DefaultTmConfig()
	tmConfig.SetRoot(s.Home)
	clientConfig := DefaultClientConfig()
	generateAndWritePackedConfig(dCmd, appConfig, tmConfig, clientConfig, false)
	s.logFile(GetFullPathToPackedConf(dCmd))
	s.Require().NoError(loadPackedConfig(dCmd), "loadPackedConfig")

	s.Run("tmcmds.ParseConfig", func() {
		var tmConfig2 *tmconfig.Config
		var err error
		s.Require().NotPanics(func() {
			tmConfig2, err = tmcmds.ParseConfig(dCmd)
		})
		s.Require().NoError(err, "tmcmds.ParseConfig")
		s.Assert().Equal(tmConfig, tmConfig2)
	})

	s.Run("ExtractTmConfig", func() {
		var tmConfig2 *tmconfig.Config
		var err error
		s.Require().NotPanics(func() {
			tmConfig2, err = ExtractTmConfig(dCmd)
		})
		s.Require().NoError(err, "ExtractTmConfig")
		s.Assert().Equal(tmConfig, tmConfig2)
	})
}
