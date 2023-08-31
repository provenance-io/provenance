package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/provenance-io/provenance/internal/pioconfig"
)

const (
	// CustomDenomFlag flag to take in custom denom, defaults to nhash if not passed in.
	CustomDenomFlag = "custom-denom"
	// CustomMsgFeeFloorPriceFlag flag to take in custom msg floor fees, defaults to 1905nhash if not passed in.
	CustomMsgFeeFloorPriceFlag = "msgfee-floor-price"
)

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
	// I'd just have to call it again later anyway because deeper stuff will probably update it.
	vpr := client.GetClientContextFromCmd(cmd).Viper

	// And now, set up Viper a little more.
	if err := bindFlagsAndEnv(cmd, vpr); err != nil {
		return err
	}

	// Set the pio config now so that the proper default is set for the rest of the stuff.
	SetPioConfigFromFlags(cmd.Flags())

	// Create a new Server context with the same viper as the client context, a default config, and no logger.
	serverCtx := server.NewContext(vpr, DefaultTmConfig(), nil)
	if err := server.SetCmdServerContext(cmd, serverCtx); err != nil {
		return err
	}

	// Read the configs into viper and the contexts.
	return LoadConfigFromFiles(cmd)
}

func SetPioConfigFromFlags(flagSet *pflag.FlagSet) {
	// Ignoring errors here in the off chance that the flags weren't defined originally.
	customDenom, _ := flagSet.GetString(CustomDenomFlag)
	customMsgFeeFloor, _ := flagSet.GetInt64(CustomMsgFeeFloorPriceFlag)
	pioconfig.SetProvenanceConfig(customDenom, customMsgFeeFloor)
}

// Binds viper flags using the PIO ENV prefix.
func bindFlagsAndEnv(cmd *cobra.Command, v *viper.Viper) (err error) {
	defer func() {
		recover() //nolint:errcheck // err already set to needed return value.
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
