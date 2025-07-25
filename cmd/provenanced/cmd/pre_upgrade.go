package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	cmtconfig "github.com/cometbft/cometbft/config"

	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	cmderrors "github.com/provenance-io/provenance/cmd/errors"
	"github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/provenance-io/provenance/internal/pioconfig"
)

var (
	ErrFail      error = cmderrors.ExitCodeError(30)
	ErrFailRetry error = cmderrors.ExitCodeError(31)
)

// GetPreUpgradeCmd returns the pre-upgrade command which cosmovisor runs before
// starting a node after an upgrade. Anyone not using cosmovisor should manually
// run this after swapping executables and before restarting the node.
func GetPreUpgradeCmd() *cobra.Command {
	// https://docs.cosmos.network/main/build/building-apps/app-upgrade#pre-upgrade-handling
	// The exit code meanings are dictated by cosmovisor.
	cmd := &cobra.Command{
		Use:   "pre-upgrade", // cosmovisor requires it to be this.
		Short: "Pre-Upgrade command",
		Long: `A command to be run as part of an upgrade.
It should be run using the new version before restarting the node.

Exit code meanings:
   0 - Success. The node should be started as usual.
   1 - Returned only if this command does not exist. The node should be started as usual.
  30 - Execution failed and the node should not be restarted.
  31 - Execution failed, but this command should be re-attempted until it returns either 0 or 30.`,
		Hidden:       true, // This isn't a command that we need to advertise in provenanced help.
		SilenceUsage: true, // No need to print usage if the command fails.
		// Cosmovisor doesn't provide any args, and none are expected. But we want an
		// exit code of 30 here (instead of 1), so we're not using cobra.NoArgs.
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.Join(fmt.Errorf("expected 0 args, received %d", len(args)), ErrFail)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			err := UpdateConfig(cmd)
			if err != nil {
				return errors.Join(fmt.Errorf("could not update config file(s): %w", err), ErrFail)
			}
			cmd.Printf("pre-upgrade successful\n")
			return nil
		},
	}

	return cmd
}

// UpdateConfig writes the current config to files.
// During a pre-upgrade, this, at the very least, updates the config file using
// the most recent template. It might also force-change some config values.
func UpdateConfig(cmd *cobra.Command) error {
	// Load all the config objects.
	// This depends on the configs already being read into viper.
	// Usually that is done with the root command's PersistentPreRunE.
	// If the configs change too much, though, you might need to read/load
	// them using something else that correctly handles in the previous version.
	appCfg, err := config.ExtractAppConfig(cmd)
	if err != nil {
		return err
	}
	cmtCfg, err := config.ExtractCmtConfig(cmd)
	if err != nil {
		return err
	}
	clientCfg, err := config.ExtractClientConfig(cmd)
	if err != nil {
		return err
	}

	if clientCfg.BroadcastMode == "block" {
		cmd.Printf("Updating the broadcast_mode config value to \"sync\" (from %q, which is no longer an option).\n", clientCfg.BroadcastMode)
		clientCfg.BroadcastMode = "sync"
	}

	piocfg := pioconfig.GetProvConfig()
	if appCfg.MinGasPrices != piocfg.ProvMinGasPrices {
		cmd.Printf("Updating the minimum-gas-prices config value to %q (from %q, to accommodate flat fees).\n", piocfg.ProvMinGasPrices, appCfg.MinGasPrices)
		appCfg.MinGasPrices = piocfg.ProvMinGasPrices
	}

	return SafeSaveConfigs(cmd, appCfg, cmtCfg, clientCfg, true)
}

// SafeSaveConfigs calls config.SaveConfigs but returns an error instead of panicking.
func SafeSaveConfigs(cmd *cobra.Command,
	appConfig *serverconfig.Config,
	cmtConfig *cmtconfig.Config,
	clientConfig *config.ClientConfig,
	verbose bool,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, isErr := r.(error); isErr {
				err = fmt.Errorf("error saving config file(s): %w", e)
			} else {
				err = fmt.Errorf("error saving config file(s): %v", r)
			}
		}
	}()
	config.SaveConfigs(cmd, appConfig, cmtConfig, clientConfig, verbose)
	return nil
}
