package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	tmconfig "github.com/tendermint/tendermint/config"

	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	"github.com/provenance-io/provenance/cmd/provenanced/config"
)

var (
	ErrFail      error = server.ErrorCode{Code: 30}
	ErrFailRetry error = server.ErrorCode{Code: 31}
)

// GetPreUpgradeCmd returns the pre-upgrade command which cosmovisor runs before
// starting a node after an upgrade. Anyone not using cosmovisor should manually
// run this after swapping executables and before restarting the node.
func GetPreUpgradeCmd() *cobra.Command {
	// https://docs.cosmos.network/main/building-apps/app-upgrade#pre-upgrade-handling
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
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.Join(fmt.Errorf("expected 0 args, received %d", len(args)), ErrFail)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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
	tmCfg, err := config.ExtractTmConfig(cmd)
	if err != nil {
		return err
	}
	clientCfg, err := config.ExtractClientConfig(cmd)
	if err != nil {
		return err
	}

	if clientCfg.ChainID == "pio-mainnet-1" {
		// Update the timeout commit if it's too low.
		timeoutCommit := config.DefaultConsensusTimeoutCommit
		if tmCfg.Consensus.TimeoutCommit < timeoutCommit/2 {
			cmd.Printf("Updating consensus.timeout_commit config value to %q (from %q)\n",
				timeoutCommit, tmCfg.Consensus.TimeoutCommit)
			tmCfg.Consensus.TimeoutCommit = timeoutCommit
		}
	}

	return SafeSaveConfigs(cmd, appCfg, tmCfg, clientCfg, true)
}

// SafeSaveConfigs calls config.SaveConfigs but returns an error instead of panicking.
func SafeSaveConfigs(cmd *cobra.Command,
	appConfig *serverconfig.Config,
	tmConfig *tmconfig.Config,
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
	config.SaveConfigs(cmd, appConfig, tmConfig, clientConfig, verbose)
	return nil
}
