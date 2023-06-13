package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	tmconfig "github.com/tendermint/tendermint/config"

	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	"github.com/provenance-io/provenance/cmd/provenanced/config"
)

func GetPreUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "pre-upgrade",
		Short:        "Pre-Upgrade command",
		Long:         "Pre-upgrade command to implement custom pre-upgrade handling",
		Args:         cobra.NoArgs,
		Hidden:       true,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			err := UpdateConfig(cmd)
			if err != nil {
				cmd.PrintErrf("could not update config file(s): %v\n", err)
				os.Exit(30)
			}
		},
	}

	return cmd
}

// UpdateConfig writes the current config to files.
func UpdateConfig(cmd *cobra.Command) error {
	// This depends on the configs already being loaded.
	// Usually this is done with the root command's PersistentPreRunE.
	// If the config(s) change too much though, you might need to read/load
	// them using something else that correctly reads in the previous version.

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

	if tmCfg.Consensus.TimeoutCommit <= 4*time.Second {
		tmCfg.Consensus.TimeoutCommit = 5 * time.Second
	}

	return SafeSaveConfigs(cmd, appCfg, tmCfg, clientCfg, true)
}

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
