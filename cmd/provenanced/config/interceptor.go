package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	tmconfig "github.com/tendermint/tendermint/config"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
)

// InterceptConfigsPreRunHandler performs a pre-run function for all commands.
// It will finish setting up the client context and create the server context.
// It will create a Viper literal and the configs will be read and parsed or created from defaults.
// The viper literal is used to read and parse configurations. Command handlers can
// fetch the server or client contexts to get the Tendermint, App/Cosmos, or Client
// configurations, or to get access to viper.
func InterceptConfigsPreRunHandler(cmd *cobra.Command) error {
	// The result of client.GetClientContextFromCmd(cmd) is not a pointer.
	// I need it for both viper and the home dir.
	// Don't use the clientCtx variable after these three lines, though since it's info will probably be stale.
	clientCtx := client.GetClientContextFromCmd(cmd)
	vpr := clientCtx.Viper
	// Set the home dir in viper so that all the appOpts stuff gets the correct value.
	vpr.Set(flags.FlagHome, clientCtx.HomeDir)

	// And now, set up Viper a little more.
	if err := bindFlagsAndEnv(cmd, vpr); err != nil {
		return err
	}

	// Create a new Server context with the same viper as the client context, a default config, and no logger.
	serverCtx := server.NewContext(vpr, tmconfig.DefaultConfig(), nil)
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
		recover() //nolint:errcheck
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
