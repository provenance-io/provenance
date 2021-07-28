package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	tmcfg "github.com/tendermint/tendermint/config"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

// InterceptConfigsPreRunHandler performs a pre-run function for the root daemon
// application command. It will create a Viper literal and a default server
// Context. The server Tendermint configuration will either be read and parsed
// or created and saved to disk, where the server Context is updated to reflect
// the Tendermint configuration. It takes custom app config template and config
// settings to create a custom Tendermint configuration. If the custom template
// is empty, it uses default-template provided by the server. The Viper literal
// is used to read and parse the application configuration. Command handlers can
// fetch the server Context to get the Tendermint configuration or to get access
// to Viper.
// NOTE: This function is duplicated here from the SDK due to forced override of ENV
// prefix using the binary name which breaks provenanced configuration.
func InterceptConfigsPreRunHandler(cmd *cobra.Command, customAppConfigTemplate string, customAppConfig interface{}) error {
	serverCtx := server.NewDefaultContext()

	// Configure the viper instance
	serverCtx.Viper.BindPFlags(cmd.Flags())
	serverCtx.Viper.BindPFlags(cmd.PersistentFlags())
	serverCtx.Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	serverCtx.Viper.AutomaticEnv()

	// intercept configuration files, using both Viper instances separately
	config, err := interceptConfigs(serverCtx.Viper, customAppConfigTemplate, customAppConfig)
	if err != nil {
		return err
	}

	// return value is a tendermint configuration object
	serverCtx.Config = config
	if err = bindFlags(cmd, serverCtx.Viper); err != nil {
		return err
	}

	var logWriter io.Writer
	if strings.ToLower(serverCtx.Viper.GetString(flags.FlagLogFormat)) == tmcfg.LogFormatPlain {
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
	} else {
		logWriter = os.Stderr
	}

	logLvlStr := serverCtx.Viper.GetString(flags.FlagLogLevel)
	logLvl, err := zerolog.ParseLevel(logLvlStr)
	if err != nil {
		return fmt.Errorf("failed to parse log level (%s): %w", logLvlStr, err)
	}

	serverCtx.Logger = server.ZeroLogWrapper{zerolog.New(logWriter).Level(logLvl).With().Timestamp().Logger()}

	return server.SetCmdServerContext(cmd, serverCtx)
}

// interceptConfigs parses and updates a Tendermint configuration file or
// creates a new one and saves it. It also parses and saves the application
// configuration file. The Tendermint configuration file is parsed given a root
// Viper object, whereas the application is parsed with the private package-aware
// viperCfg object.
func interceptConfigs(rootViper *viper.Viper, customAppTemplate string, customConfig interface{}) (*tmcfg.Config, error) {
	rootDir := rootViper.GetString(flags.FlagHome)
	configPath := filepath.Join(rootDir, "config")
	tmCfgFile := filepath.Join(configPath, "config.toml")

	conf := tmcfg.DefaultConfig()

	switch _, err := os.Stat(tmCfgFile); {
	case os.IsNotExist(err):
		tmcfg.EnsureRoot(rootDir)

		if err = conf.ValidateBasic(); err != nil {
			return nil, fmt.Errorf("error in config file: %v", err)
		}

		conf.RPC.PprofListenAddress = "localhost:6060"
		conf.Consensus.TimeoutCommit = 4 * time.Second
		tmcfg.WriteConfigFile(tmCfgFile, conf)

	case err != nil:
		return nil, err

	default:
		rootViper.SetConfigType("toml")
		rootViper.SetConfigName("config")
		rootViper.AddConfigPath(configPath)

		if err := rootViper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read in %s: %w", tmCfgFile, err)
		}
	}

	// Read into the configuration whatever data the viper instance has for it.
	// This may come from the configuration file above but also any of the other
	// sources viper uses.
	if err := rootViper.Unmarshal(conf); err != nil {
		return nil, err
	}

	conf.SetRoot(rootDir)

	appCfgFilePath := filepath.Join(configPath, "app.toml")
	if _, err := os.Stat(appCfgFilePath); os.IsNotExist(err) {
		if customAppTemplate != "" {
			serverconfig.SetConfigTemplate(customAppTemplate)

			if err = rootViper.Unmarshal(&customConfig); err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", appCfgFilePath, err)
			}

			serverconfig.WriteConfigFile(appCfgFilePath, customConfig)
		} else {
			appConf, err := serverconfig.ParseConfig(rootViper)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", appCfgFilePath, err)
			}

			serverconfig.WriteConfigFile(appCfgFilePath, appConf)
		}
	}

	rootViper.SetConfigType("toml")
	rootViper.SetConfigName("app")
	rootViper.AddConfigPath(configPath)

	if err := rootViper.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("failed to merge configuration: %w", err)
	}

	return conf, nil
}

// Binds viper flags using the PIO ENV prefix.
func bindFlags(cmd *cobra.Command, v *viper.Viper) (err error) {
	defer func() {
		recover() // nolint:errcheck
	}()

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		err = v.BindEnv(f.Name, fmt.Sprintf("PIO_%s", strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))))
		if err != nil {
			panic(err)
		}

		err = v.BindPFlag(f.Name, f)
		if err != nil {
			panic(err)
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			err = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				panic(err)
			}
		}
	})

	return err
}
