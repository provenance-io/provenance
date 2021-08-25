package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	tmcfg "github.com/tendermint/tendermint/config"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
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

// InterceptConfigsPreRunHandler performs a pre-run function for all commands.
// It will finish setting up the client context and create the server context.
// It will create a Viper literal and the configs will be read and parsed or created from defaults.
// The viper literal is used to read and parse configurations. Command handlers can
// fetch the server or client contexts to get the Tendermint, App/Cosmos, or Client
// configurations, or to get access to viper.
func InterceptConfigsPreRunHandler(cmd *cobra.Command) error {
	// The result of client.GetClientContextFromCmd(cmd) is not a pointer.
	// Since I'm just getting the Viper pointer from it (for now), I'm not
	// pulling the context into its own variable.
	// I'd just have to call it again later anyway.
	vpr := client.GetClientContextFromCmd(cmd).Viper

	// And now, set up Viper a little more.
	if err := bindFlagsAndEnv(cmd, vpr); err != nil {
		return err
	}

	// Create the logger that we'll want for the new server context
	var logWriter io.Writer
	if strings.ToLower(vpr.GetString(flags.FlagLogFormat)) == tmcfg.LogFormatPlain {
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
	} else {
		logWriter = os.Stderr
	}

	logLvlStr := vpr.GetString(flags.FlagLogLevel)
	logLvl, perr := zerolog.ParseLevel(logLvlStr)
	if perr != nil {
		return fmt.Errorf("failed to parse log level (%s): %w", logLvlStr, perr)
	}

	serverLogger := server.ZeroLogWrapper{Logger: zerolog.New(logWriter).Level(logLvl).With().Timestamp().Logger()}

	// Create a new Server context with the same viper as the client context, a default config, and the just defined logger.
	serverCtx := server.NewContext(vpr, tmcfg.DefaultConfig(), serverLogger)
	// The server context is provided as a pointer so we should be okay to just set it up here and keep defining it.
	if err := server.SetCmdServerContext(cmd, serverCtx); err != nil {
		return err
	}

	// Read the configs into viper and the contexts.
	if err := LoadConfigFromFiles(cmd); err != nil {
		return err
	}

	return nil
}

// Binds viper flags using the PIO ENV prefix.
func bindFlagsAndEnv(cmd *cobra.Command, v *viper.Viper) (err error) {
	defer func() {
		recover() // nolint:errcheck
	}()

	replacer := strings.NewReplacer(".", "_", "-", "_")

	v.SetEnvKeyReplacer(replacer)
	v.AutomaticEnv()

	if err = v.BindPFlags(cmd.Flags()); err != nil {
		return err
	}
	if err = v.BindPFlags(cmd.PersistentFlags()); err != nil {
		return err
	}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to PIO_FAVORITE_COLOR
		err = v.BindEnv(f.Name, "PIO_"+strings.ToUpper(replacer.Replace(f.Name)))
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

// NewDefaultReplacer creates a new strings.Replacer using some default replacements.
func NewDefaultReplacer() *strings.Replacer {
	return strings.NewReplacer(".", "_", "-", "_")
}
